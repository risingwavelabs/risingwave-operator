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

package scaleview

import (
	"fmt"

	"github.com/samber/lo"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type RisingWaveScaleViewHelper struct {
	risingwave *risingwavev1alpha1.RisingWave
	component  string
}

func findGroup[T any](groups []T, target string, name func(*T) string) (*T, int) {
	for i, group := range groups {
		if name(&group) == target {
			return &groups[i], i
		}
	}
	return nil, 0
}

// getReplicasGlobalPtr returns the global pointer to the respective group.
func (r *RisingWaveScaleViewHelper) getReplicasGlobalPtr() *int32 {
	switch r.component {
	case consts.ComponentMeta:
		return &r.risingwave.Spec.Global.Replicas.Meta
	case consts.ComponentCompute:
		return &r.risingwave.Spec.Global.Replicas.Compute
	case consts.ComponentFrontend:
		return &r.risingwave.Spec.Global.Replicas.Frontend
	case consts.ComponentCompactor:
		return &r.risingwave.Spec.Global.Replicas.Compactor
	default:
		panic(fmt.Sprintf("Unknown component %v", r.component))
	}
}

func (r *RisingWaveScaleViewHelper) getReplicasPtr(group string) (*int32, int) {
	if group == "" {
		return r.getReplicasGlobalPtr(), 0
	}

	var gPtr *risingwavev1alpha1.RisingWaveComponentGroup
	idx := -1

	switch r.component {
	case consts.ComponentMeta:
		gPtr, idx = findGroup(r.risingwave.Spec.Components.Meta.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
	case consts.ComponentFrontend:
		gPtr, idx = findGroup(r.risingwave.Spec.Components.Frontend.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
	case consts.ComponentCompactor:
		gPtr, idx = findGroup(r.risingwave.Spec.Components.Compactor.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
	case consts.ComponentCompute:
		cgPtr, i := findGroup(r.risingwave.Spec.Components.Compute.Groups, group, func(g *risingwavev1alpha1.RisingWaveComputeGroup) string { return g.Name })
		if cgPtr == nil {
			return nil, 0
		}
		return &cgPtr.Replicas, i
	default:
		panic(fmt.Sprintf("Unknown component %v", r.component))
	}
	if gPtr == nil {
		return nil, 0
	}
	return &gPtr.Replicas, idx
}

// ListComponentGroups lists the groups under `.spec.components`. The default group "" is not included.
func (r *RisingWaveScaleViewHelper) ListComponentGroups() []string {
	switch r.component {
	case consts.ComponentMeta:
		return lo.Map(r.risingwave.Spec.Components.Meta.Groups, func(g risingwavev1alpha1.RisingWaveComponentGroup, _ int) string { return g.Name })
	case consts.ComponentFrontend:
		return lo.Map(r.risingwave.Spec.Components.Frontend.Groups, func(g risingwavev1alpha1.RisingWaveComponentGroup, _ int) string { return g.Name })
	case consts.ComponentCompute:
		return lo.Map(r.risingwave.Spec.Components.Compute.Groups, func(g risingwavev1alpha1.RisingWaveComputeGroup, _ int) string { return g.Name })
	case consts.ComponentCompactor:
		return lo.Map(r.risingwave.Spec.Components.Compactor.Groups, func(g risingwavev1alpha1.RisingWaveComponentGroup, _ int) string { return g.Name })
	default:
		panic("never reach here")
	}
}

// GetGroupIndex gets the index of the given group in the list under `.spec.components.{component}.groups`.
func (r *RisingWaveScaleViewHelper) GetGroupIndex(group string) (int, bool) {
	replicasPtr, i := r.getReplicasPtr(group)
	if replicasPtr == nil {
		return 0, false
	}
	return i, true
}

// ReadReplicas reads the replicas of the given group. It returns 0 and false if the group is not found.
func (r *RisingWaveScaleViewHelper) ReadReplicas(group string) (int32, bool) {
	replicasPtr, _ := r.getReplicasPtr(group)
	if replicasPtr == nil {
		return 0, false
	}
	return *replicasPtr, true
}

// WriteReplicas writes the replicas to the given group. It returns true if the group is found and the value is changed.
func (r *RisingWaveScaleViewHelper) WriteReplicas(group string, replicas int32) bool {
	replicasPtr, _ := r.getReplicasPtr(group)
	if replicasPtr == nil || *replicasPtr == replicas {
		return false
	}
	*replicasPtr = replicas
	return true
}

func NewRisingWaveScaleViewHelper(risingwave *risingwavev1alpha1.RisingWave, component string) *RisingWaveScaleViewHelper {
	return &RisingWaveScaleViewHelper{
		risingwave: risingwave,
		component:  component,
	}
}
