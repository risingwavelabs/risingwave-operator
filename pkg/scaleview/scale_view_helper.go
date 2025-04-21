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

// Package scaleview provides utilities for operating in-memory RisingWave and RisingWaveScaleView objects. For the
// detailed design of RisingWaveScaleView, please refer to the RFC-0004.
package scaleview

import (
	"fmt"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

// RisingWaveScaleViewHelper is a helper struct to help get and update the replicas in the scale view lock records.
type RisingWaveScaleViewHelper struct {
	risingwave *risingwavev1alpha1.RisingWave
	component  string
}

func findNodeGroup[T any](groups []T, target string, name func(*T) string) (*T, int) {
	for i, group := range groups {
		if name(&group) == target {
			return &groups[i], i
		}
	}

	return nil, 0
}

func (r *RisingWaveScaleViewHelper) findReplicaPtrFromNodeGroups(nodeGroups []risingwavev1alpha1.RisingWaveNodeGroup, group string) (*int32, int) {
	ptr, i := findNodeGroup(nodeGroups, group, func(g *risingwavev1alpha1.RisingWaveNodeGroup) string { return g.Name })
	if ptr == nil {
		return nil, 0
	}

	return &ptr.Replicas, i
}

// findReplicaPtrGroup finds the pointer to the required group.
func (r *RisingWaveScaleViewHelper) findReplicaPtrGroup(group string) (*int32, int) {
	switch r.component {
	case consts.ComponentMeta:
		return r.findReplicaPtrFromNodeGroups(r.risingwave.Spec.Components.Meta.NodeGroups, group)
	case consts.ComponentFrontend:
		return r.findReplicaPtrFromNodeGroups(r.risingwave.Spec.Components.Frontend.NodeGroups, group)
	case consts.ComponentCompactor:
		return r.findReplicaPtrFromNodeGroups(r.risingwave.Spec.Components.Compactor.NodeGroups, group)
	case consts.ComponentCompute:
		return r.findReplicaPtrFromNodeGroups(r.risingwave.Spec.Components.Compute.NodeGroups, group)
	case consts.ComponentStandalone:
		panic("not supported")
	default:
		panic(fmt.Sprintf("Unknown component %v", r.component))
	}
}

// getReplicasPtr returns the pointer and group index to the required group or to the global group. Returns nil ptr if group is not found.
func (r *RisingWaveScaleViewHelper) getReplicasPtr(group string) (*int32, int) {
	return r.findReplicaPtrGroup(group)
}

// ListComponentGroups lists the groups under `.spec.components`.
func (r *RisingWaveScaleViewHelper) ListComponentGroups() []string {
	var nodeGroups []risingwavev1alpha1.RisingWaveNodeGroup

	switch r.component {
	case consts.ComponentMeta:
		nodeGroups = r.risingwave.Spec.Components.Meta.NodeGroups
	case consts.ComponentFrontend:
		nodeGroups = r.risingwave.Spec.Components.Frontend.NodeGroups
	case consts.ComponentCompute:
		nodeGroups = r.risingwave.Spec.Components.Compute.NodeGroups
	case consts.ComponentCompactor:
		nodeGroups = r.risingwave.Spec.Components.Compactor.NodeGroups
	default:
		panic(fmt.Sprintf("unknown component %v", r.component))
	}

	names := make([]string, 0, len(nodeGroups))
	for _, ng := range nodeGroups {
		names = append(names, ng.Name)
	}

	return names
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

// NewRisingWaveScaleViewHelper creates a new RisingWaveScaleViewHelper.
func NewRisingWaveScaleViewHelper(risingwave *risingwavev1alpha1.RisingWave, component string) *RisingWaveScaleViewHelper {
	return &RisingWaveScaleViewHelper{
		risingwave: risingwave,
		component:  component,
	}
}
