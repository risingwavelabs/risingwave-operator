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

package testutils

import (
	"fmt"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/utils/ptr"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// Scheme for test only.
var Scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = risingwavev1alpha1.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = prometheusv1.AddToScheme(Scheme)
	_ = kruiseappsv1alpha1.AddToScheme(Scheme)
	_ = kruiseappsv1beta1.AddToScheme(Scheme)
}

// FakeRisingWaveOpenKruiseEnabled returns a new fake RisingWave with OpenKruise enabled.
func FakeRisingWaveOpenKruiseEnabled() *risingwavev1alpha1.RisingWave {
	risingwaveCopy := fakeRisingWave.DeepCopy()
	risingwaveCopy.Spec.EnableOpenKruise = ptr.To(true)
	return risingwaveCopy
}

// FakeRisingWaveOpenKruiseDisabled returns a new fake RisingWave with OpenKruise disabled.
func FakeRisingWaveOpenKruiseDisabled() *risingwavev1alpha1.RisingWave {
	risingwaveCopy := fakeRisingWave.DeepCopy()
	risingwaveCopy.Spec.EnableOpenKruise = ptr.To(false)
	return risingwaveCopy
}

// FakeRisingWave returns a new fake Risingwave.
func FakeRisingWave() *risingwavev1alpha1.RisingWave {
	return fakeRisingWave.DeepCopy()
}

var fakeRisingWave = &risingwavev1alpha1.RisingWave{
	TypeMeta: metav1.TypeMeta{
		Kind:       "RisingWave",
		APIVersion: "risingwave.risingwavelabs.com/v1alpha1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:       "fake-risingwave",
		Namespace:  "default",
		Generation: 2,
		UID:        uuid.NewUUID(),
	},
	Spec: risingwavev1alpha1.RisingWaveSpec{
		MetaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
			Memory: ptr.To(true),
		},
		StateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
			Memory: ptr.To(true),
		},
		FrontendServiceType: corev1.ServiceTypeClusterIP,
		Image:               "ghcr.io/risingwavelabs/risingwave:latest",
		Components: risingwavev1alpha1.RisingWaveComponentsSpec{
			Meta: risingwavev1alpha1.RisingWaveComponent{
				NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("1Gi"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("100Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			Frontend: risingwavev1alpha1.RisingWaveComponent{
				NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("1Gi"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("100Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			Compute: risingwavev1alpha1.RisingWaveComponent{
				NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("1Gi"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("100Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
			Compactor: risingwavev1alpha1.RisingWaveComponent{
				NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Resources: corev1.ResourceRequirements{
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("1"),
											corev1.ResourceMemory: resource.MustParse("1Gi"),
										},
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("100Mi"),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
	Status: risingwavev1alpha1.RisingWaveStatus{
		ObservedGeneration: 1,
		MetaStore: risingwavev1alpha1.RisingWaveMetaStoreStatus{
			Backend: risingwavev1alpha1.RisingWaveMetaStoreBackendTypeMemory,
		},
		StateStore: risingwavev1alpha1.RisingWaveStateStoreStatus{
			Backend: risingwavev1alpha1.RisingWaveStateStoreBackendTypeMemory,
		},
		Conditions: []risingwavev1alpha1.RisingWaveCondition{
			{
				Type:               risingwavev1alpha1.RisingWaveConditionRunning,
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Now(),
			},
		},
		ComponentReplicas: risingwavev1alpha1.RisingWaveComponentsReplicasStatus{
			Meta: risingwavev1alpha1.ComponentReplicasStatus{
				Target:  1,
				Running: 1,
				Groups: []risingwavev1alpha1.ComponentGroupReplicasStatus{
					{
						Name:    "",
						Target:  1,
						Running: 1,
					},
				},
			},
			Frontend: risingwavev1alpha1.ComponentReplicasStatus{
				Target:  1,
				Running: 1,
				Groups: []risingwavev1alpha1.ComponentGroupReplicasStatus{
					{
						Name:    "",
						Target:  1,
						Running: 1,
					},
				},
			},
			Compute: risingwavev1alpha1.ComponentReplicasStatus{
				Target:  1,
				Running: 1,
				Groups: []risingwavev1alpha1.ComponentGroupReplicasStatus{
					{
						Name:    "",
						Target:  1,
						Running: 1,
					},
				},
			},
			Compactor: risingwavev1alpha1.ComponentReplicasStatus{
				Target:  1,
				Running: 1,
				Groups: []risingwavev1alpha1.ComponentGroupReplicasStatus{
					{
						Name:    "",
						Target:  1,
						Running: 1,
					},
				},
			},
		},
	},
}

// DeepEqual returns true when the two objects are semantically equal.
func DeepEqual[T any](x, y T) bool {
	return equality.Semantic.DeepEqual(x, y)
}

// NewFakeRisingWaveScaleViewFor creates a new fake RisingWaveScaleView object for the target RisingWave and component. It applies
// the given mutations before returning.
func NewFakeRisingWaveScaleViewFor(risingwave *risingwavev1alpha1.RisingWave, component string, mutates ...func(*risingwavev1alpha1.RisingWave, *risingwavev1alpha1.RisingWaveScaleView)) *risingwavev1alpha1.RisingWaveScaleView {
	r := &risingwavev1alpha1.RisingWaveScaleView{
		TypeMeta: metav1.TypeMeta{
			Kind:       "RisingWaveScaleView",
			APIVersion: "risingwave.risingwavelabs.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:       "fake-risingwave-scaleview-" + rand.String(4),
			Namespace:  "default",
			Generation: 1,
			UID:        uuid.NewUUID(),
		},
		Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
			TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
				Name:      risingwave.Name,
				Component: component,
			},
		},
	}
	for _, m := range mutates {
		m(risingwave, r)
	}
	return r
}

// FakeRisingWaveWithMutate creates a new fake RisingWave and applies the mutate function. It's for test purposes.
func FakeRisingWaveWithMutate(mutate func(wave *risingwavev1alpha1.RisingWave)) *risingwavev1alpha1.RisingWave {
	r := FakeRisingWave()
	mutate(r)
	return r
}

// GetNodeGroupName returns the group name for a specified index.
func GetNodeGroupName(i int) string {
	if i == 0 {
		return ""
	}
	return fmt.Sprintf("group-%d", i)
}
