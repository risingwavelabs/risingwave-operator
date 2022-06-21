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

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveManager struct {
	client            client.Client
	risingwave        *risingwavev1alpha1.RisingWave // Immutable.
	mutableRisingWave *risingwavev1alpha1.RisingWave // Mutable copy of original.

	mu sync.RWMutex
}

// RisingWave returns the risingwave immutable reference.
func (mgr *RisingWaveManager) RisingWave() *risingwavev1alpha1.RisingWave {
	return mgr.risingwave
}

// IsObservedGenerationOutdated tells whether the observed generation is outdated.
func (mgr *RisingWaveManager) IsObservedGenerationOutdated() bool {
	return mgr.risingwave.Status.ObservedGeneration < mgr.risingwave.Generation
}

// SyncObservedGeneration updates the observed generation to the current generation.
func (mgr *RisingWaveManager) SyncObservedGeneration() {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.mutableRisingWave.Status.ObservedGeneration = mgr.risingwave.Generation
}

func (mgr *RisingWaveManager) ObjectStorageType() risingwavev1alpha1.ObjectStorageType {
	objectStorage := mgr.risingwave.Spec.ObjectStorage
	switch {
	case objectStorage.Memory:
		return risingwavev1alpha1.MemoryType
	case objectStorage.MinIO != nil:
		return risingwavev1alpha1.MinIOType
	case objectStorage.S3 != nil:
		return risingwavev1alpha1.S3Type
	default:
		return risingwavev1alpha1.UnknownType
	}
}

func (mgr *RisingWaveManager) GetCondition(conditionType risingwavev1alpha1.RisingWaveType) *risingwavev1alpha1.RisingWaveCondition {
	for _, cond := range mgr.risingwave.Status.Conditions {
		if cond.Type == conditionType {
			return cond.DeepCopy()
		}
	}
	return nil
}

func (mgr *RisingWaveManager) RemoveCondition(conditionType risingwavev1alpha1.RisingWaveType) {
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
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	// Set the last transition time to now if it's a new condition or status changed.
	lastObservedCondition := mgr.GetCondition(condition.Type)
	if lastObservedCondition == nil || lastObservedCondition.Status != condition.Status {
		condition.LastTransitionTime = metav1.Now()
	}

	// Add or update the conidtion.
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

func NewRisingWaveManager(client client.Client, risingwave *risingwavev1alpha1.RisingWave) *RisingWaveManager {
	return &RisingWaveManager{
		client:            client,
		risingwave:        risingwave,
		mutableRisingWave: risingwave.DeepCopy(),
	}
}
