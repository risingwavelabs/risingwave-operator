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
	"reflect"
	"testing"

	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func TestScaleViewLockManager_splitReplicasIntoGroups(t *testing.T) {
	testcases := map[string]struct {
		sv       risingwavev1alpha1.RisingWaveScaleView
		expected map[string]int32
	}{
		"one": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(10),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "",
						},
					},
				},
			},
			expected: map[string]int32{
				"": 10,
			},
		},
		"two-unlimited": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(9),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "a",
						},
						{
							Group: "b",
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 5,
				"b": 4,
			},
		},
		"two-but-one-limited": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(9),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "a",
						},
						{
							Group:       "b",
							MaxReplicas: pointer.Int32(2),
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 7,
				"b": 2,
			},
		},
		"three-with-priorities-1": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(9),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    10,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 3,
				"b": 2,
				"c": 4,
			},
		},
		"three-with-priorities-2": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(9),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:    "a",
							Priority: 10,
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 9,
				"b": 0,
				"c": 0,
			},
		},
		"three-with-priorities-3": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(4),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    10,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 3,
				"b": 1,
				"c": 0,
			},
		},
		"three-with-priorities-4": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: pointer.Int32(4),
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    5,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 2,
				"b": 2,
				"c": 0,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r := SplitReplicas(&tc.sv)
			if !reflect.DeepEqual(r, tc.expected) {
				t.Fatalf("wrong result: %s", testutils.JsonMustPrettyPrint(r))
			}
		})
	}
}
