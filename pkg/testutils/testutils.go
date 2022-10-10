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

package testutils

import (
	"fmt"

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
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

// Scheme for test only.
var Scheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(Scheme)
	_ = risingwavev1alpha1.AddToScheme(Scheme)
	_ = apiextensionsv1.AddToScheme(Scheme)
	_ = prometheusv1.AddToScheme(Scheme)
}

// Fake RisingWave.
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
		Storages: risingwavev1alpha1.RisingWaveStoragesSpec{
			Meta: risingwavev1alpha1.RisingWaveMetaStorage{
				Memory: pointer.Bool(true),
			},
			Object: risingwavev1alpha1.RisingWaveObjectStorage{
				Memory: pointer.Bool(true),
			},
		},
		Global: risingwavev1alpha1.RisingWaveGlobalSpec{
			Replicas: risingwavev1alpha1.RisingWaveGlobalReplicas{
				Meta:      1,
				Compute:   1,
				Frontend:  1,
				Compactor: 1,
			},
			ServiceType: corev1.ServiceTypeClusterIP,
			RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
				Image:           "ghcr.io/risingwavelabs/risingwave:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				NodeSelector: map[string]string{
					"kubernetes.io/os":   "linux",
					"kubernetes.io/arch": "amd64",
				},
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
		Components: risingwavev1alpha1.RisingWaveComponentsSpec{
			Meta: risingwavev1alpha1.RisingWaveComponentMeta{
				Ports: risingwavev1alpha1.RisingWaveComponentMetaPorts{
					RisingWaveComponentCommonPorts: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultMetaServicePort,
						MetricsPort: consts.DefaultMetaMetricsPort,
					},
					DashboardPort: consts.DefaultMetaDashboardPort,
				},
			},
			Frontend: risingwavev1alpha1.RisingWaveComponentFrontend{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultFrontendServicePort,
					MetricsPort: consts.DefaultFrontendMetricsPort,
				},
			},
			Compute: risingwavev1alpha1.RisingWaveComponentCompute{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultComputeServicePort,
					MetricsPort: consts.DefaultComputeMetricsPort,
				},
			},
			Compactor: risingwavev1alpha1.RisingWaveComponentCompactor{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultCompactorServicePort,
					MetricsPort: consts.DefaultCompactorMetricsPort,
				},
			},
		},
	},
	Status: risingwavev1alpha1.RisingWaveStatus{
		ObservedGeneration: 1,
		Storages: risingwavev1alpha1.RisingWaveStoragesStatus{
			Meta: risingwavev1alpha1.RisingWaveMetaStorageStatus{
				Type: risingwavev1alpha1.MetaStorageTypeMemory,
			},
			Object: risingwavev1alpha1.RisingWaveObjectStorageStatus{
				Type: risingwavev1alpha1.ObjectStorageTypeMemory,
			},
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

func FakeRisingWave() *risingwavev1alpha1.RisingWave {
	return fakeRisingWave.DeepCopy()
}

func GetGroupName(index int) string {
	return fmt.Sprintf("group-%d", index)
}

var fakeRisingWaveComponentOnly = &risingwavev1alpha1.RisingWave{
	TypeMeta: metav1.TypeMeta{
		Kind:       "RisingWave",
		APIVersion: "risingwave.risingwavelabs.com/v1alpha1",
	},
	ObjectMeta: metav1.ObjectMeta{
		Name:       "fake-risingwave-component-only",
		Namespace:  "default",
		Generation: 2,
		UID:        uuid.NewUUID(),
	},
	Spec: risingwavev1alpha1.RisingWaveSpec{
		Storages: risingwavev1alpha1.RisingWaveStoragesSpec{
			Meta: risingwavev1alpha1.RisingWaveMetaStorage{
				Memory: pointer.Bool(true),
			},
			Object: risingwavev1alpha1.RisingWaveObjectStorage{
				Memory: pointer.Bool(true),
			},
		},
		Global: risingwavev1alpha1.RisingWaveGlobalSpec{
			ServiceType: corev1.ServiceTypeClusterIP,
			RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
				Image:           "ghcr.io/risingwavelabs/risingwave:latest",
				ImagePullPolicy: corev1.PullIfNotPresent,
				NodeSelector: map[string]string{
					"kubernetes.io/os":   "linux",
					"kubernetes.io/arch": "amd64",
				},
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
		Components: risingwavev1alpha1.RisingWaveComponentsSpec{
			Meta: risingwavev1alpha1.RisingWaveComponentMeta{
				Ports: risingwavev1alpha1.RisingWaveComponentMetaPorts{
					RisingWaveComponentCommonPorts: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultMetaServicePort,
						MetricsPort: consts.DefaultMetaMetricsPort,
					},
					DashboardPort: consts.DefaultMetaDashboardPort,
				},
				Groups: []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     GetGroupName(0),
						Replicas: 1,
					},
				},
			},
			Frontend: risingwavev1alpha1.RisingWaveComponentFrontend{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultFrontendServicePort,
					MetricsPort: consts.DefaultFrontendMetricsPort,
				},
				Groups: []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     GetGroupName(0),
						Replicas: 1,
					},
				},
			},
			Compute: risingwavev1alpha1.RisingWaveComponentCompute{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultComputeServicePort,
					MetricsPort: consts.DefaultComputeMetricsPort,
				},
				Groups: []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     GetGroupName(0),
						Replicas: 1,
					},
				},
			},
			Compactor: risingwavev1alpha1.RisingWaveComponentCompactor{
				Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
					ServicePort: consts.DefaultCompactorServicePort,
					MetricsPort: consts.DefaultCompactorMetricsPort,
				},
				Groups: []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     GetGroupName(0),
						Replicas: 1,
					},
				},
			},
		},
	},
	Status: risingwavev1alpha1.RisingWaveStatus{
		ObservedGeneration: 1,
		Storages: risingwavev1alpha1.RisingWaveStoragesStatus{
			Meta: risingwavev1alpha1.RisingWaveMetaStorageStatus{
				Type: risingwavev1alpha1.MetaStorageTypeMemory,
			},
			Object: risingwavev1alpha1.RisingWaveObjectStorageStatus{
				Type: risingwavev1alpha1.ObjectStorageTypeMemory,
			},
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
						Name:    GetGroupName(0),
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
						Name:    GetGroupName(0),
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
						Name:    GetGroupName(0),
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
						Name:    GetGroupName(0),
						Target:  1,
						Running: 1,
					},
				},
			},
		},
	},
}

func FakeRisingWaveComponentOnly() *risingwavev1alpha1.RisingWave {
	return fakeRisingWaveComponentOnly.DeepCopy()
}

func DeepEqual[T any](x, y T) bool {
	return equality.Semantic.DeepEqual(x, y)
}

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

func FakeRisingWaveWithMutate(mutate func(wave *risingwavev1alpha1.RisingWave)) *risingwavev1alpha1.RisingWave {
	r := FakeRisingWave()
	mutate(r)
	return r
}
