/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package object

import (
	"context"
	"sync"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// RisingWaveReader is a reader for RisingWave object.
type RisingWaveReader struct {
	risingwave *risingwavev1alpha1.RisingWave // Immutable.
}

// NewRisingWaveReader creates a RisingWaveReader.
func NewRisingWaveReader(risingwave *risingwavev1alpha1.RisingWave) *RisingWaveReader {
	return &RisingWaveReader{risingwave: risingwave}
}

// RisingWave returns the RisingWave immutable reference.
func (r *RisingWaveReader) RisingWave() *risingwavev1alpha1.RisingWave {
	return r.risingwave
}

// IsObservedGenerationOutdated tells whether the observed generation is outdated.
func (r *RisingWaveReader) IsObservedGenerationOutdated() bool {
	return r.risingwave.Status.ObservedGeneration < r.risingwave.Generation
}

func (r *RisingWaveReader) GetCondition(conditionType risingwavev1alpha1.RisingWaveConditionType) *risingwavev1alpha1.RisingWaveCondition {
	for _, cond := range r.risingwave.Status.Conditions {
		if cond.Type == conditionType {
			return cond.DeepCopy()
		}
	}
	return nil
}

func (r *RisingWaveReader) DoesConditionExistAndEqual(conditionType risingwavev1alpha1.RisingWaveConditionType, value bool) bool {
	cond := r.GetCondition(conditionType)
	return cond != nil && (cond.Status == metav1.ConditionTrue) == value
}

type RisingWaveManager struct {
	RisingWaveReader

	client client.Client

	mu                 sync.RWMutex
	mutableRisingWave  *risingwavev1alpha1.RisingWave // Mutable copy of original.
	openkruiseAvailble bool                           // availability and admistrative switch of openkruise
}

// RisingWaveAfterImage returns a copy of the mutable RisingWave.
func (mgr *RisingWaveManager) RisingWaveAfterImage() *risingwavev1alpha1.RisingWave {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	return mgr.mutableRisingWave.DeepCopy()
}

// SyncObservedGeneration updates the observed generation to the current generation.
func (mgr *RisingWaveManager) SyncObservedGeneration() {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.mutableRisingWave.Status.ObservedGeneration = mgr.mutableRisingWave.Generation
}

func (mgr *RisingWaveManager) RemoveCondition(conditionType risingwavev1alpha1.RisingWaveConditionType) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	conditions := mgr.mutableRisingWave.Status.Conditions
	for i, cond := range conditions {
		if cond.Type == conditionType {
			// Remove it.
			conditions = append(conditions[:i], conditions[i+1:]...)
			mgr.mutableRisingWave.Status.Conditions = conditions
			return
		}
	}
}

func (mgr *RisingWaveManager) UpdateCondition(condition risingwavev1alpha1.RisingWaveCondition) {
	// Set the last transition time to now if it's a new condition or status changed.
	lastObservedCondition := mgr.GetCondition(condition.Type)
	if lastObservedCondition == nil || lastObservedCondition.Status != condition.Status {
		condition.LastTransitionTime = metav1.Now()
	}

	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	
	conditions := mgr.mutableRisingWave.Status.Conditions
	_, curIndex, found := lo.FindIndexOf(conditions, func(cond risingwavev1alpha1.RisingWaveCondition) bool {
		return cond.Type == condition.Type
	})
	if found {
		conditions[curIndex] = condition
	} else {
		conditions = append(conditions, condition)
		mgr.mutableRisingWave.Status.Conditions = conditions
	}
}

func (mgr *RisingWaveManager) UpdateStatus(f func(*risingwavev1alpha1.RisingWaveStatus)) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	f(&mgr.mutableRisingWave.Status)
}

func (mgr *RisingWaveManager) UpdateRemoteRisingWaveStatus(ctx context.Context) error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// Do nothing if not changed.
	if equality.Semantic.DeepEqual(&mgr.mutableRisingWave.Status, &mgr.risingwave.Status) {
		return nil
	}

	return mgr.client.Status().Update(ctx, mgr.mutableRisingWave)
}

func (mgr *RisingWaveManager) IsOpenKruiseAvailable() bool {
	return mgr.openkruiseAvailble
}

func (mgr *RisingWaveManager) IsOpenKruiseEnabled() bool {
	risingwave := mgr.RisingWaveReader.RisingWave()
	return mgr.IsOpenKruiseAvailable() && risingwave.Spec.EnableOpenKruise != nil && *risingwave.Spec.EnableOpenKruise
}

func NewRisingWaveManager(client client.Client, risingwave *risingwavev1alpha1.RisingWave, openkruiseAvailble bool) *RisingWaveManager {
	return &RisingWaveManager{
		RisingWaveReader: RisingWaveReader{
			risingwave: risingwave,
		},
		client:             client,
		mutableRisingWave:  risingwave.DeepCopy(),
		openkruiseAvailble: openkruiseAvailble,
	}
}
