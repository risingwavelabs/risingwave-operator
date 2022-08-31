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

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func NewFakeClient() client.Client {
	risingwave := &v1alpha1.RisingWave{}
	c := fake.NewClientBuilder().
		WithScheme(testutils.Schema).
		WithObjects(risingwave).
		Build()
	return c
}

var fakeRW = &v1alpha1.RisingWave{
	ObjectMeta: metav1.ObjectMeta{
		Namespace: "test-namespace",
		Name:      "test-name",
	},
	Spec: v1alpha1.RisingWaveSpec{
		Global: v1alpha1.RisingWaveGlobalSpec{
			Replicas: v1alpha1.RisingWaveGlobalReplicas{
				Meta:      1,
				Frontend:  2,
				Compute:   3,
				Compactor: 4,
			},
			RisingWaveComponentGroupTemplate: v1alpha1.RisingWaveComponentGroupTemplate{
				Image: "test.image.global",
				Resources: corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("1"),
						corev1.ResourceMemory: resource.MustParse("1Gi"),
					}},
			},
		},
		Storages: v1alpha1.RisingWaveStoragesSpec{
			Meta: v1alpha1.RisingWaveMetaStorage{
				Memory: pointer.Bool(true),
			},
			Object: v1alpha1.RisingWaveObjectStorage{
				Memory: pointer.Bool(true),
			},
		},
		Components: v1alpha1.RisingWaveComponentsSpec{
			Meta: v1alpha1.RisingWaveComponentMeta{
				Groups: []v1alpha1.RisingWaveComponentGroup{
					{
						Name: "meta-group-1",
						RisingWaveComponentGroupTemplate: &v1alpha1.RisingWaveComponentGroupTemplate{
							Image: "test.image.meta",
						},
					},
				},
			},
			Frontend: v1alpha1.RisingWaveComponentFrontend{
				Groups: []v1alpha1.RisingWaveComponentGroup{
					{
						Name:     "frontend-group-1",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &v1alpha1.RisingWaveComponentGroupTemplate{
							Image: "test.image.frontend",
						},
					},
					{
						Name:     "frontend-group-2",
						Replicas: 2,
						RisingWaveComponentGroupTemplate: &v1alpha1.RisingWaveComponentGroupTemplate{
							Image: "test.image.frontend",
						},
					},
				},
			},

			Compute: v1alpha1.RisingWaveComponentCompute{
				Groups: []v1alpha1.RisingWaveComputeGroup{
					{
						Name:     "compute-group-1",
						Replicas: 2,
						RisingWaveComputeGroupTemplate: &v1alpha1.RisingWaveComputeGroupTemplate{
							RisingWaveComponentGroupTemplate: v1alpha1.RisingWaveComponentGroupTemplate{
								Image: "test.image.compute",
							},
						},
					},
				},
			},
			Compactor: v1alpha1.RisingWaveComponentCompactor{
				Groups: []v1alpha1.RisingWaveComponentGroup{
					{
						Name: "compactor-group-1",
						RisingWaveComponentGroupTemplate: &v1alpha1.RisingWaveComponentGroupTemplate{
							Image: "test.image.compactor",
						},
					},
				},
			},
		},
	},
}

func FakeRW() *v1alpha1.RisingWave {
	return fakeRW.DeepCopy()
}
