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

package util

import "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"

// ConvertRisingwave will mv the risingwave Global replicas into each component group.
func ConvertRisingwave(rw *v1alpha1.RisingWave) *v1alpha1.RisingWave {
	var newRW = rw.DeepCopy()

	// set the global replicas is nil.
	newRW.Spec.Global.Replicas = v1alpha1.RisingWaveGlobalReplicas{}

	convertReplicas(rw, newRW)
	return newRW
}

func convertReplicas(oldRW, newRW *v1alpha1.RisingWave) {
	rs := oldRW.Spec.Global.Replicas
	global := oldRW.Spec.Global
	if rs.Meta != 0 {
		newRW.Spec.Components.Meta.Groups = append(newRW.Spec.Components.Meta.Groups, v1alpha1.RisingWaveComponentGroup{
			Name:     DefaultGroup,
			Replicas: rs.Meta,

			RisingWaveComponentGroupTemplate: global.RisingWaveComponentGroupTemplate.DeepCopy(),
		})
	}

	if rs.Compute != 0 {
		newRW.Spec.Components.Compute.Groups = append(newRW.Spec.Components.Compute.Groups, v1alpha1.RisingWaveComputeGroup{
			Name:     DefaultGroup,
			Replicas: rs.Compute,

			RisingWaveComputeGroupTemplate: &v1alpha1.RisingWaveComputeGroupTemplate{
				RisingWaveComponentGroupTemplate: *global.RisingWaveComponentGroupTemplate.DeepCopy(),
			},
		})
	}

	if rs.Frontend != 0 {
		newRW.Spec.Components.Frontend.Groups = append(newRW.Spec.Components.Frontend.Groups, v1alpha1.RisingWaveComponentGroup{
			Name:     DefaultGroup,
			Replicas: rs.Frontend,

			RisingWaveComponentGroupTemplate: global.RisingWaveComponentGroupTemplate.DeepCopy(),
		})
	}

	if rs.Compactor != 0 {
		newRW.Spec.Components.Compactor.Groups = append(newRW.Spec.Components.Compactor.Groups, v1alpha1.RisingWaveComponentGroup{
			Name:     DefaultGroup,
			Replicas: rs.Compactor,

			RisingWaveComponentGroupTemplate: global.RisingWaveComponentGroupTemplate.DeepCopy(),
		})
	}

}
