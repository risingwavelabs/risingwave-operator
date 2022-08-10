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

package util

import "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"

func IsValidGroup(rw *v1alpha1.RisingWave, group string) bool {
	for _, g := range rw.Spec.Components.Compute.Groups {
		if g.Name == group {
			return true
		}
	}
	for _, g := range rw.Spec.Components.Meta.Groups {
		if g.Name == group {
			return true
		}
	}
	for _, g := range rw.Spec.Components.Compactor.Groups {
		if g.Name == group {
			return true
		}
	}
	for _, g := range rw.Spec.Components.Frontend.Groups {
		if g.Name == group {
			return true
		}
	}
	return false
}
