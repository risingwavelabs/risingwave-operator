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

package utils

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func Test_SortEnvVarSlice(t *testing.T) {
	testcases := map[string]struct {
		envVar       []corev1.EnvVar
		expectEnvVar []corev1.EnvVar
	}{
		"env-default": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
				{
					Name:  "ENV_A",
					Value: "valueA",
				},
			},

			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_A",
					Value: "valueA",
				},
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
			},
		},
		"env-dependencies": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
			},
			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
			},
		},
		"env-escaped": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
				{
					Name:  "ENV_A",
					Value: "$$(ENV_B)_suffix",
				},
			},

			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_A",
					Value: "$$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
			},
		},
		"env-dependencies-escaped": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_C",
					Value: "$$(ENV_A)_$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
			},
			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_C",
					Value: "$$(ENV_A)_$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "valueB",
				},
			},
		},
		"env-fake-transitive": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_C",
					Value: "$(ENV_D)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "$(ENV_C)_suffix",
				},
			},
			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
				{
					Name:  "ENV_C",
					Value: "$(ENV_D)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "$(ENV_C)_suffix",
				},
			},
		},
		"env-input-alphabetical": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_C",
					Value: "$(ENV_D)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "$(ENV_C)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_C)_suffix",
				},
			},
			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_C",
					Value: "$(ENV_D)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_C)_suffix",
				},
				{
					Name:  "ENV_B",
					Value: "$(ENV_C)_suffix",
				},
			},
		},
		"env-circular-dependencies": {
			envVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "$(ENV_A)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
			},
			expectEnvVar: []corev1.EnvVar{
				{
					Name:  "ENV_B",
					Value: "$(ENV_A)_suffix",
				},
				{
					Name:  "ENV_A",
					Value: "$(ENV_B)_suffix",
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			TopologicalSort(tc.envVar)
			if !equality.Semantic.DeepEqual(tc.envVar, tc.expectEnvVar) {
				t.Fail()
			}
		})
	}
}
