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
	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type ComponentGroupReplicasManager struct {
	risingwave *risingwavev1alpha1.RisingWave
	component  string
}

func findGroup[T any](groups []T, target string, name func(*T) string) (*T, int) {
	for i, group := range groups {
		if name(&group) == target {
			return &group, i
		}
	}
	return nil, 0
}

func (r *ComponentGroupReplicasManager) getReplicasPtr(group string) (*int32, int) {
	switch r.component {
	case consts.ComponentMeta:
		if group == "" {
			return &r.risingwave.Spec.Global.Replicas.Meta, 0
		} else {
			g, i := findGroup(r.risingwave.Spec.Components.Meta.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
			if g == nil {
				return nil, 0
			}
			return &g.Replicas, i
		}
	case consts.ComponentFrontend:
		if group == "" {
			return &r.risingwave.Spec.Global.Replicas.Frontend, 0
		} else {
			g, i := findGroup(r.risingwave.Spec.Components.Frontend.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
			if g == nil {
				return nil, 0
			}
			return &g.Replicas, i
		}
	case consts.ComponentCompactor:
		if group == "" {
			return &r.risingwave.Spec.Global.Replicas.Compactor, 0
		} else {
			g, i := findGroup(r.risingwave.Spec.Components.Compactor.Groups, group, func(g *risingwavev1alpha1.RisingWaveComponentGroup) string { return g.Name })
			if g == nil {
				return nil, 0
			}
			return &g.Replicas, i
		}
	case consts.ComponentCompute:
		if group == "" {
			return &r.risingwave.Spec.Global.Replicas.Compute, 0
		} else {
			g, i := findGroup(r.risingwave.Spec.Components.Compute.Groups, group, func(g *risingwavev1alpha1.RisingWaveComputeGroup) string { return g.Name })
			if g == nil {
				return nil, 0
			}
			return &g.Replicas, i
		}
	default:
		panic("never reach here")
	}
}

func (r *ComponentGroupReplicasManager) GetGroupIndex(group string) (int, bool) {
	replicasPtr, i := r.getReplicasPtr(group)
	if replicasPtr == nil {
		return 0, false
	}
	return i, true
}

func (r *ComponentGroupReplicasManager) ReadReplicas(group string) (int32, bool) {
	replicasPtr, _ := r.getReplicasPtr(group)
	if replicasPtr == nil {
		return 0, false
	}
	return *replicasPtr, true
}

func (r *ComponentGroupReplicasManager) WriteReplicas(group string, replicas int32) bool {
	replicasPtr, _ := r.getReplicasPtr(group)

	if replicasPtr != nil {
		if *replicasPtr == replicas {
			return false
		}
		*replicasPtr = replicas
	}
	return replicasPtr != nil
}

func NewComponentGroupReplicasManager(risingwave *risingwavev1alpha1.RisingWave, component string) *ComponentGroupReplicasManager {
	return &ComponentGroupReplicasManager{
		risingwave: risingwave,
		component:  component,
	}
}