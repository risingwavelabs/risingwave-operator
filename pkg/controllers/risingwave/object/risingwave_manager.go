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

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveManager struct {
	client           client.Client
	risingwave       *risingwavev1alpha1.RisingWave       // Status mutable.
	risingwaveStatus *risingwavev1alpha1.RisingWaveStatus // Immutable copy of original.
}

func (mgr *RisingWaveManager) RisingWave() *risingwavev1alpha1.RisingWave {
	return mgr.risingwave
}

func (mgr *RisingWaveManager) IsObservedGenerationOutdated() bool {
	return mgr.risingwaveStatus.ObservedGeneration < mgr.risingwave.Generation
}

func (mgr *RisingWaveManager) SyncObservedGeneration() {
	mgr.risingwave.Generation = mgr.risingwaveStatus.ObservedGeneration
}

func (mgr *RisingWaveManager) GetCondition(conditionType risingwavev1alpha1.RisingWaveType) *risingwavev1alpha1.RisingWaveCondition {
	for _, cond := range mgr.risingwaveStatus.Conditions {
		if cond.Type == conditionType {
			return cond.DeepCopy()
		}
	}
	return nil
}

func (mgr *RisingWaveManager) RemoveCondition(conditionType risingwavev1alpha1.RisingWaveType) {
	conditions := mgr.risingwave.Status.Conditions
	for i, cond := range conditions {
		if cond.Type == conditionType {
			// Remove it.
			conditions = append(conditions[:i], conditions[i+1:]...)
			mgr.risingwave.Status.Conditions = conditions
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

	// Add or update the conidtion.
	conditions := mgr.risingwave.Status.Conditions
	_, curIndex, found := lo.FindIndexOf(conditions, func(cond risingwavev1alpha1.RisingWaveCondition) bool {
		return cond.Type == condition.Type
	})
	if found {
		conditions[curIndex] = condition
	} else {
		conditions = append(conditions, condition)
		mgr.risingwave.Status.Conditions = conditions
	}
}

func (mgr *RisingWaveManager) UpdateRisingWaveStatus(ctx context.Context) error {
	// Do nothing if not changed.
	if equality.Semantic.DeepEqual(mgr.risingwaveStatus, &mgr.risingwave.Status) {
		return nil
	}

	return mgr.client.Status().Update(ctx, mgr.risingwave)
}

func NewRisingWaveManager(client client.Client, risingwave *risingwavev1alpha1.RisingWave) *RisingWaveManager {
	return &RisingWaveManager{
		client:           client,
		risingwave:       risingwave,
		risingwaveStatus: risingwave.Status.DeepCopy(),
	}
}
