/*
 * Copyright 2023 RisingWave Labs
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
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
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
	return r.risingwave.DeepCopy()
}

// StateStoreRootPath returns the value of `Root` fields of state stores if defined.
//
//nolint:all
func (r *RisingWaveReader) StateStoreRootPath() string {
	stateStore := r.risingwave.Spec.StateStore
	switch {
	case stateStore.AzureBlob != nil:
		return stateStore.AzureBlob.Root
	case stateStore.AliyunOSS != nil:
		return stateStore.AliyunOSS.Root
	case stateStore.GCS != nil:
		return stateStore.GCS.Root
	case stateStore.HDFS != nil:
		return stateStore.HDFS.Root
	case stateStore.WebHDFS != nil:
		return stateStore.WebHDFS.Root
	default:
		return ""
	}
}

// IsObservedGenerationOutdated tells whether the observed generation is outdated.
func (r *RisingWaveReader) IsObservedGenerationOutdated() bool {
	return r.risingwave.Status.ObservedGeneration < r.risingwave.Generation
}

// GetCondition gets the condition object by the given type. It returns nil when not found.
func (r *RisingWaveReader) GetCondition(conditionType risingwavev1alpha1.RisingWaveConditionType) *risingwavev1alpha1.RisingWaveCondition {
	for _, cond := range r.risingwave.Status.Conditions {
		if cond.Type == conditionType {
			return cond.DeepCopy()
		}
	}

	return nil
}

// DoesConditionExistAndEqual returns true if the condition is found and its value equals to the given one.
func (r *RisingWaveReader) DoesConditionExistAndEqual(conditionType risingwavev1alpha1.RisingWaveConditionType, value bool) bool {
	cond := r.GetCondition(conditionType)

	return cond != nil && (cond.Status == metav1.ConditionTrue) == value
}

// GetNodeGroups gets the node groups of the given component. It panics when the component is unknown.
func (r *RisingWaveReader) GetNodeGroups(component string) []risingwavev1alpha1.RisingWaveNodeGroup {
	switch component {
	case consts.ComponentMeta:
		return r.risingwave.Spec.Components.Meta.NodeGroups
	case consts.ComponentCompactor:
		return r.risingwave.Spec.Components.Compactor.NodeGroups
	case consts.ComponentFrontend:
		return r.risingwave.Spec.Components.Frontend.NodeGroups
	case consts.ComponentCompute:
		return r.risingwave.Spec.Components.Compute.NodeGroups
	case consts.ComponentStandalone:
		panic("not supported")
	default:
		panic("unknown component: " + component)
	}
}

// GetNodeGroup gets the node groups of the given component and group.
// It panics when the component is unknown and returns nil when the node group isn't found.
func (r *RisingWaveReader) GetNodeGroup(component, group string) *risingwavev1alpha1.RisingWaveNodeGroup {
	nodeGroups := r.GetNodeGroups(component)
	for _, nodeGroup := range nodeGroups {
		if nodeGroup.Name == group {
			return nodeGroup.DeepCopy()
		}
	}

	return nil
}

// RisingWaveManager is a struct to help manipulate the RisingWave object in memory. It is concurrent-safe.
type RisingWaveManager struct {
	RisingWaveReader

	client client.Client

	mu                sync.RWMutex
	mutableRisingWave *risingwavev1alpha1.RisingWave // Mutable copy of original.

	openkruiseAvailable bool // Availability and administrative switch of openkruise
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

// RemoveCondition removes the condition if the condition type matches.
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

// UpdateCondition updates the conditions in the mutable copy and sets the last transition time automatically
// according to the latest conditions from the original status. It will append a new condition status when there's no such
// condition before.
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

// UpdateStatus receives a function to mutate the RisingWaveStatus and runs it within a lock.
func (mgr *RisingWaveManager) UpdateStatus(f func(*risingwavev1alpha1.RisingWaveStatus)) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	f(&mgr.mutableRisingWave.Status)
}

// UpdateRemoteRisingWaveStatus updates the remote RisingWave object with the mutable copy.
func (mgr *RisingWaveManager) UpdateRemoteRisingWaveStatus(ctx context.Context) error {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()

	// Do nothing if not changed.
	if equality.Semantic.DeepEqual(&mgr.mutableRisingWave.Status, &mgr.risingwave.Status) {
		return nil
	}

	return mgr.client.Status().Update(ctx, mgr.mutableRisingWave)
}

// IsOpenKruiseAvailable returns true when the OpenKruise is available.
func (mgr *RisingWaveManager) IsOpenKruiseAvailable() bool {
	return mgr.openkruiseAvailable
}

// IsOpenKruiseEnabled returns true when the OpenKruise is available and enabled on the target RisingWave.
func (mgr *RisingWaveManager) IsOpenKruiseEnabled() bool {
	risingwave := mgr.RisingWave()

	return mgr.IsOpenKruiseAvailable() && risingwave.Spec.EnableOpenKruise != nil && *risingwave.Spec.EnableOpenKruise
}

// IsStandaloneModeEnabled returns true when the standalone mode is enabled.
func (r *RisingWaveReader) IsStandaloneModeEnabled() bool {
	return ptr.Deref(r.risingwave.Spec.EnableStandaloneMode, false)
}

// IsAdvertisingWithIP returns true when the advertising with IP is enabled.
func (r *RisingWaveReader) IsAdvertisingWithIP() bool {
	return ptr.Deref(r.risingwave.Spec.EnableAdvertisingWithIP, false)
}

// KeepLock resets the current scale views record in the status with the given array.
func (mgr *RisingWaveManager) KeepLock(aliveScaleView []risingwavev1alpha1.RisingWaveScaleViewLock) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()

	mgr.mutableRisingWave.Status.ScaleViews = aliveScaleView
}

// NewRisingWaveManager creates a new RisingWaveManager with given arguments.
func NewRisingWaveManager(client client.Client, risingwave *risingwavev1alpha1.RisingWave, openkruiseAvailable bool) *RisingWaveManager {
	return &RisingWaveManager{
		RisingWaveReader: RisingWaveReader{
			risingwave: risingwave,
		},
		client:              client,
		mutableRisingWave:   risingwave.DeepCopy(),
		openkruiseAvailable: openkruiseAvailable,
	}
}
