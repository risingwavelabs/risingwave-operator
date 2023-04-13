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

package factory

import (
	"math"
	"strconv"
	"time"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type risingWaveComponentGroup interface {
	risingwavev1alpha1.RisingWaveComputeGroup |
		risingwavev1alpha1.RisingWaveComponentGroup
}

type kubeObjectsUpgradeStrategy interface {
	*appsv1.DeploymentStrategy |
		*appsv1.StatefulSetUpdateStrategy |
		*kruiseappsv1alpha1.CloneSetUpdateStrategy |
		*kruiseappsv1beta1.StatefulSetUpdateStrategy
}

type testCaseType interface {
	baseTestCase |
		deploymentTestCase |
		cloneSetTestCase |
		computeStatefulSetTestCase |
		computeAdvancedSTSTestCase |
		servicesTestCase |
		serviceMetadataTestCase |
		objectStoragesTestCase |
		configMapTestCase |
		computeArgsTestCase |
		metaStorageTestCase |
		metaStatefulSetTestCase |
		metaAdvancedSTSTestCase
}

type kubeObject interface {
	*corev1.Service |
		*corev1.Secret |
		*corev1.ConfigMap |
		*corev1.Pod |
		*corev1.PersistentVolumeClaim |
		*corev1.PersistentVolume |
		*appsv1.Deployment |
		*appsv1.StatefulSet |
		*kruiseappsv1beta1.StatefulSet |
		*kruiseappsv1alpha1.CloneSet |
		*prometheusv1.ServiceMonitor
}

type baseTestCase struct {
	risingwave *risingwavev1alpha1.RisingWave
}

type testCase[T risingWaveComponentGroup, K kubeObjectsUpgradeStrategy] struct {
	baseTestCase
	component               string
	podTemplate             map[string]risingwavev1alpha1.RisingWavePodTemplate
	group                   T
	expectedUpgradeStrategy K
	restartAt               *metav1.Time
}

type deploymentTestCase testCase[risingwavev1alpha1.RisingWaveComponentGroup, *appsv1.DeploymentStrategy]

func deploymentTestCases() map[string]deploymentTestCase {
	return map[string]deploymentTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					NodeSelector: map[string]string{
						"a": "b",
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Tolerations: []corev1.Toleration{
						{
							Key:               "key1",
							Operator:          "Equal",
							Value:             "value1",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{3600}[0],
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:             rand.String(20),
					PriorityClassName: "high-priority",
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:           &[]int64{1000}[0],
						RunAsGroup:          &[]int64{3000}[0],
						FSGroup:             &[]int64{2000}[0],
						FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"1.2.3.4"},
						Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
						Options: []corev1.PodDNSConfigOption{
							{
								// spellchecker: disable
								Name:  "ndots",
								Value: &[]string{"2"}[0],
							},
							{
								// spellchecker: disable
								Name: "edns0",
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:                         rand.String(20),
					TerminationGracePeriodSeconds: pointer.Int64(5),
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:           rand.String(20),
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					ImagePullSecrets: []string{
						"a",
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
				RollingUpdate: &appsv1.RollingUpdateDeployment{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"resources-1c1g": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		"with-pod-template": {
			podTemplate: map[string]risingwavev1alpha1.RisingWavePodTemplate{
				"test": {
					Template: *newPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
						t.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						}
					}),
				},
			},
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:       rand.String(20),
					PodTemplate: pointer.String("test"),
				},
			},
		},
	}
}

type metaStatefulSetTestCase testCase[risingwavev1alpha1.RisingWaveComponentGroup, *appsv1.StatefulSetUpdateStrategy]

func metaStatefulSetTestCases() map[string]metaStatefulSetTestCase {
	return map[string]metaStatefulSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					NodeSelector: map[string]string{
						"a": "b",
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Tolerations: []corev1.Toleration{
						{
							Key:               "key1",
							Operator:          "Equal",
							Value:             "value1",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{3600}[0],
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:             rand.String(20),
					PriorityClassName: "high-priority",
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:           &[]int64{1000}[0],
						RunAsGroup:          &[]int64{3000}[0],
						FSGroup:             &[]int64{2000}[0],
						FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"1.2.3.4"},
						Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
						Options: []corev1.PodDNSConfigOption{
							{
								// spellchecker: disable
								Name:  "ndots",
								Value: &[]string{"2"}[0],
							},
							{
								// spellchecker: disable
								Name: "edns0",
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:                         rand.String(20),
					TerminationGracePeriodSeconds: pointer.Int64(5),
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:           rand.String(20),
					ImagePullPolicy: corev1.PullAlways,
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					ImagePullSecrets: []string{
						"a",
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					},
				},
			},
			expectedUpgradeStrategy: nil,
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"resources-1c1g": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		"with-pod-template": {
			podTemplate: map[string]risingwavev1alpha1.RisingWavePodTemplate{
				"test": {
					Template: *newPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
						t.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						}
					}),
				},
			},
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:       rand.String(20),
					PodTemplate: pointer.String("test"),
				},
			},
		},
	}
}

type computeStatefulSetTestCase testCase[risingwavev1alpha1.RisingWaveComputeGroup, *appsv1.StatefulSetUpdateStrategy]

func computeStatefulSetTestCases() map[string]computeStatefulSetTestCase {
	return map[string]computeStatefulSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Labels: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Tolerations: []corev1.Toleration{
							{
								Key:               "key1",
								Operator:          "Equal",
								Value:             "value1",
								Effect:            "NoExecute",
								TolerationSeconds: &[]int64{3600}[0],
							},
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:             rand.String(20),
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:           &[]int64{1000}[0],
							RunAsGroup:          &[]int64{3000}[0],
							FSGroup:             &[]int64{2000}[0],
							FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
						},
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						DNSConfig: &corev1.PodDNSConfig{
							Nameservers: []string{"1.2.3.4"},
							Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
							Options: []corev1.PodDNSConfigOption{
								{
									Name:  "ndots",
									Value: &[]string{"2"}[0],
								},
								{
									Name: "edns0",
								},
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:                         rand.String(20),
						TerminationGracePeriodSeconds: pointer.Int64(5),
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:           rand.String(20),
						ImagePullPolicy: corev1.PullAlways,
					},
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						},
					},
				},
			},
			expectedUpgradeStrategy: nil,
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
							RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
								MaxUnavailable: &intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "50%",
								},
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"resources-1c1g": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("1Gi"),
							},
						},
					},
				},
			},
		},
		"with-pod-template": {
			podTemplate: map[string]risingwavev1alpha1.RisingWavePodTemplate{
				"test": {
					Template: *newPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
						t.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						}
					}),
				},
			},
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:       rand.String(20),
						PodTemplate: pointer.String("test"),
					},
				},
			},
		},
		"with-pvc-volumes-mounts": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "test",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "t",
							MountPath: "/tt",
						},
					},
				},
			},
		},
	}
}

type cloneSetTestCase testCase[risingwavev1alpha1.RisingWaveComponentGroup, *kruiseappsv1alpha1.CloneSetUpdateStrategy]

func cloneSetTestCases() map[string]cloneSetTestCase {
	return map[string]cloneSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					NodeSelector: map[string]string{
						"a": "b",
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Tolerations: []corev1.Toleration{
						{
							Key:               "key1",
							Operator:          "Equal",
							Value:             "value1",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{3600}[0],
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:             rand.String(20),
					PriorityClassName: "high-priority",
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:           &[]int64{1000}[0],
						RunAsGroup:          &[]int64{3000}[0],
						FSGroup:             &[]int64{2000}[0],
						FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"1.2.3.4"},
						Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
						Options: []corev1.PodDNSConfigOption{
							{
								Name:  "ndots",
								Value: &[]string{"2"}[0],
							},
							{
								Name: "edns0",
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:                         rand.String(20),
					TerminationGracePeriodSeconds: pointer.Int64(5),
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-restart": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:           rand.String(20),
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					ImagePullSecrets: []string{
						"a",
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-Recreate-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-Recreate-max-surge-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxSurge: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
				MaxSurge: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-Recreate-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							Partition: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
				Partition: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-Recreate-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
							GracePeriodSeconds: 20,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
				InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
					GracePeriodSeconds: 20,
				},
			},
		},
		"upgrade-strategy-InPlaceOnly": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-InPlaceOnly-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-max-surge-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxSurge: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
				MaxSurge: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							Partition: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
				Partition: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		// HERE
		"upgrade-strategy-InPlaceOnly-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
							GracePeriodSeconds: 20,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
				InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
					GracePeriodSeconds: 20,
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-InPlaceIfPossible-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
				MaxUnavailable: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-max-surge-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxSurge: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
				MaxSurge: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							Partition: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
				Partition: &intstr.IntOrString{
					Type:   intstr.String,
					StrVal: "50%",
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
							GracePeriodSeconds: 20,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
				InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
					GracePeriodSeconds: 20,
				},
			},
		},
		"resources-1c1g": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		"with-pod-template": {
			podTemplate: map[string]risingwavev1alpha1.RisingWavePodTemplate{
				"test": {
					Template: *newPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
						t.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						}
					}),
				},
			},
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:       rand.String(20),
					PodTemplate: pointer.String("test"),
				},
			},
		},
	}
}

type metaAdvancedSTSTestCase testCase[risingwavev1alpha1.RisingWaveComponentGroup, *kruiseappsv1beta1.StatefulSetUpdateStrategy]

func metaAdvancedSTSTestCases() map[string]metaAdvancedSTSTestCase {
	return map[string]metaAdvancedSTSTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					NodeSelector: map[string]string{
						"a": "b",
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Tolerations: []corev1.Toleration{
						{
							Key:               "key1",
							Operator:          "Equal",
							Value:             "value1",
							Effect:            "NoExecute",
							TolerationSeconds: &[]int64{3600}[0],
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:             rand.String(20),
					PriorityClassName: "high-priority",
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:           &[]int64{1000}[0],
						RunAsGroup:          &[]int64{3000}[0],
						FSGroup:             &[]int64{2000}[0],
						FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					DNSConfig: &corev1.PodDNSConfig{
						Nameservers: []string{"1.2.3.4"},
						Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
						Options: []corev1.PodDNSConfigOption{
							{
								Name:  "ndots",
								Value: &[]string{"2"}[0],
							},
							{
								Name: "edns0",
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:                         rand.String(20),
					TerminationGracePeriodSeconds: pointer.Int64(5),
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
		},
		"with-restart": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:           rand.String(20),
					ImagePullPolicy: corev1.PullAlways,
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					ImagePullSecrets: []string{
						"a",
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.RecreatePodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-Recreate-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.RecreatePodUpdateStrategyType,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"upgrade-strategy-InPlaceOnly": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							Partition: &intstr.IntOrString{
								Type:   intstr.Int,
								IntVal: 50,
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
					Partition:       pointer.Int32(50),
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
							GracePeriodSeconds: 20,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
					},
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							MaxUnavailable: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
							Partition: &intstr.IntOrString{
								Type:   intstr.String,
								StrVal: "50%",
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					Partition:       pointer.Int32(50),
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
						Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
							GracePeriodSeconds: 20,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
					},
				},
			},
		},
		"resources-1c1g": {
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image: rand.String(20),
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
		},
		"with-pod-template": {
			podTemplate: map[string]risingwavev1alpha1.RisingWavePodTemplate{
				"test": {
					Template: *newPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
						t.Spec.Containers[0].SecurityContext = &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						}
					}),
				},
			},
			group: risingwavev1alpha1.RisingWaveComponentGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
					Image:       rand.String(20),
					PodTemplate: pointer.String("test"),
				},
			},
		},
	}
}

type computeAdvancedSTSTestCase testCase[risingwavev1alpha1.RisingWaveComputeGroup, *kruiseappsv1beta1.StatefulSetUpdateStrategy]

func computeAdvancedSTSTestCases() map[string]computeAdvancedSTSTestCase {
	return map[string]computeAdvancedSTSTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Labels: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Annotations: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						Tolerations: []corev1.Toleration{
							{
								Key:               "key1",
								Operator:          "Equal",
								Value:             "value1",
								Effect:            "NoExecute",
								TolerationSeconds: &[]int64{3600}[0],
							},
						},
					},
				},
			},
		},
		"priority-class-name": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:             rand.String(20),
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:           &[]int64{1000}[0],
							RunAsGroup:          &[]int64{3000}[0],
							FSGroup:             &[]int64{2000}[0],
							FSGroupChangePolicy: &[]corev1.PodFSGroupChangePolicy{"OnRootMismatch"}[0],
						},
					},
				},
			},
		},
		"dns-config": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						DNSConfig: &corev1.PodDNSConfig{
							Nameservers: []string{"1.2.3.4"},
							Searches:    []string{"ns1.svc.cluster-domain.example", "my.dns.search.suffix"},
							Options: []corev1.PodDNSConfigOption{
								{
									Name:  "ndots",
									Value: &[]string{"2"}[0],
								},
								{
									Name: "edns0",
								},
							},
						},
					},
				},
			},
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:                         rand.String(20),
						TerminationGracePeriodSeconds: pointer.Int64(5),
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image:           rand.String(20),
						ImagePullPolicy: corev1.PullAlways,
					},
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.RecreatePodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-InPlaceOnly": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
							RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
								MaxUnavailable: &intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "50%",
								},
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					MaxUnavailable: &intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "50%",
					},
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-partition-50%": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
							RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
								Partition: &intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: 50,
								},
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					Partition:       pointer.Int32(50),
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
							InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
								GracePeriodSeconds: 20,
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
					},
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-partition-50-string": {
			group: risingwavev1alpha1.RisingWaveComputeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Image: rand.String(20),
						ImagePullSecrets: []string{
							"a",
						},
						UpgradeStrategy: risingwavev1alpha1.RisingWaveUpgradeStrategy{
							Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
							RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
								Partition: &intstr.IntOrString{
									Type:   intstr.String,
									StrVal: "50%",
								},
							},
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					Partition:       pointer.Int32(50),
				},
			},
		},
	}
}

type servicesTestCase struct {
	baseTestCase
	component         string
	ports             map[string]int32
	globalServiceType corev1.ServiceType
	expectServiceType corev1.ServiceType
}

func servicesTestCases() map[string]servicesTestCase {
	return map[string]servicesTestCase{
		"random-meta-ports": {
			component:         consts.ComponentMeta,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService:   int32(rand.Int() & 0xffff),
				consts.PortMetrics:   int32(rand.Int() & 0xffff),
				consts.PortDashboard: int32(rand.Int() & 0xffff),
			},
		},
		"random-meta-ports-node-port": {
			component:         consts.ComponentMeta,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService:   int32(rand.Int() & 0xffff),
				consts.PortMetrics:   int32(rand.Int() & 0xffff),
				consts.PortDashboard: int32(rand.Int() & 0xffff),
			},
		},
		"random-frontend-ports": {
			component:         consts.ComponentFrontend,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-frontend-ports-node-port": {
			component:         consts.ComponentFrontend,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeNodePort,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-compute-ports": {
			component:         consts.ComponentCompute,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-compute-ports-node-port": {
			component:         consts.ComponentCompute,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-compactor-ports": {
			component:         consts.ComponentCompactor,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-compactor-ports-node-port": {
			component:         consts.ComponentCompactor,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-connector-ports": {
			component:         consts.ComponentConnector,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
		"random-connector-ports-node-port": {
			component:         consts.ComponentConnector,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: int32(rand.Int() & 0xffff),
				consts.PortMetrics: int32(rand.Int() & 0xffff),
			},
		},
	}

}

type serviceMetadataTestCase struct {
	baseTestCase
	component         string
	globalServiceMeta risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta
}

func serviceMetadataTestCases() map[string]serviceMetadataTestCase {
	return map[string]serviceMetadataTestCase{
		"random-meta-labels": {
			component: consts.ComponentMeta,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-meta-annotations": {
			component: consts.ComponentMeta,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-frontend-labels": {
			component: consts.ComponentFrontend,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-frontend-annotations": {
			component: consts.ComponentFrontend,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compute-labels": {
			component: consts.ComponentCompute,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compute-annotations": {
			component: consts.ComponentCompute,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compactor-labels": {
			component: consts.ComponentCompactor,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compactor-annotations": {
			component: consts.ComponentCompactor,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-connector-labels": {
			component: consts.ComponentConnector,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-connector-annotations": {
			component: consts.ComponentConnector,
			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
	}
}

type objectStoragesTestCase struct {
	baseTestCase
	objectStorage risingwavev1alpha1.RisingWaveObjectStorage
	envs          []corev1.EnvVar
}

func objectStorageTestCases() map[string]objectStoragesTestCase {
	return map[string]objectStoragesTestCase{
		"empty_data_directory": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				DataDirectory: "",
				Memory:        pointer.Bool(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_DATA_DIRECTORY",
					Value: "",
				},
			},
		},
		"some_data_directory": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				DataDirectory: "1234",
				Memory:        pointer.Bool(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_DATA_DIRECTORY",
					Value: "1234",
				},
			},
		},
		"memory": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				Memory: pointer.Bool(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+memory",
				},
			},
		},
		"minio": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{
					Secret:   "minio-creds",
					Endpoint: "minio-endpoint:1234",
					Bucket:   "minio-hummock01",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+minio://$(MINIO_USERNAME):$(MINIO_PASSWORD)@minio-endpoint:1234/minio-hummock01",
				},
				{
					Name: "MINIO_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "minio-creds",
							},
							Key: "username",
						},
					},
				},
				{
					Name: "MINIO_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "minio-creds",
							},
							Key: "password",
						},
					},
				},
			},
		},
		"s3": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Secret: "s3-creds",
					Bucket: "s3-hummock01",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3://s3-hummock01",
				},
				{
					Name:  "AWS_S3_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name: "AWS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "AWS_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name: "AWS_REGION",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3Region,
						},
					},
				},
			},
		},
		"gcs-workload": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				GCS: &risingwavev1alpha1.RisingWaveObjectStorageGCS{
					UseWorkloadIdentity: true,
					Bucket:              "gcs-bucket",
					Root:                "gcs-root",
				},
			},
			envs: []corev1.EnvVar{{
				Name:  "RW_STATE_STORE",
				Value: "hummock+gcs://gcs-bucket@gcs-root",
			}},
		},
		"gcs-secret": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				GCS: &risingwavev1alpha1.RisingWaveObjectStorageGCS{
					UseWorkloadIdentity: false,
					Secret:              "gcs-creds",
					Bucket:              "gcs-bucket",
					Root:                "gcs-root",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+gcs://gcs-bucket@gcs-root",
				},
				{
					Name: "GOOGLE_APPLICATION_CREDENTIALS",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "gcs-creds",
							},
							Key: consts.SecretKeyGCSServiceAccountCredentials,
						},
					},
				},
			},
		},
		"aliyun-oss": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
					Secret: "s3-creds",
					Bucket: "s3-hummock01",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION).aliyuncs.com",
				},
				{
					Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_REGION",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3Region,
						},
					},
				},
			},
		},
		"aliyun-oss-internal": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
					Secret:           "s3-creds",
					Bucket:           "s3-hummock01",
					InternalEndpoint: true,
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION)-internal.aliyuncs.com",
				},
				{
					Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_REGION",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3Region,
						},
					},
				},
			},
		},
		"aliyun-oss-with-region": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
					Secret: "s3-creds",
					Bucket: "s3-hummock01",
					Region: "cn-hangzhou",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION).aliyuncs.com",
				},
				{
					Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name:  "S3_COMPATIBLE_REGION",
					Value: "cn-hangzhou",
				},
			},
		},
		"s3-compatible": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Secret:   "s3-creds",
					Bucket:   "s3-hummock01",
					Endpoint: "oss-cn-hangzhou.aliyuncs.com",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://oss-cn-hangzhou.aliyuncs.com",
				},
				{
					Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_REGION",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3Region,
						},
					},
				},
			},
		},
		"s3-compatible-virtual-hosted-style": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Secret:             "s3-creds",
					Bucket:             "s3-hummock01",
					Endpoint:           "https://oss-cn-hangzhou.aliyuncs.com",
					VirtualHostedStyle: true,
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://$(S3_COMPATIBLE_BUCKET).oss-cn-hangzhou.aliyuncs.com",
				},
				{
					Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name: "S3_COMPATIBLE_REGION",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3Region,
						},
					},
				},
			},
		},
		"s3-with-region": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Secret: "s3-creds",
					Bucket: "s3-hummock01",
					Region: "ap-southeast-1",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3://s3-hummock01",
				},
				{
					Name:  "AWS_S3_BUCKET",
					Value: "s3-hummock01",
				},
				{
					Name: "AWS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3AccessKeyID,
						},
					},
				},
				{
					Name: "AWS_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s3-creds",
							},
							Key: consts.SecretKeyAWSS3SecretAccessKey,
						},
					},
				},
				{
					Name:  "AWS_REGION",
					Value: "ap-southeast-1",
				},
			},
		},
		"endpoint-with-region-variable": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Bucket:   "s3-hummock01",
					Endpoint: "s3.${REGION}.amazonaws.com",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://s3.$(S3_COMPATIBLE_REGION).amazonaws.com",
				},
			},
		},
		"endpoint-with-bucket-variable": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
					Bucket:   "s3-hummock01",
					Endpoint: "${BUCKET}.s3.${REGION}.amazonaws.com",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3-compatible://s3-hummock01",
				},
				{
					Name:  "S3_COMPATIBLE_ENDPOINT",
					Value: "https://$(S3_COMPATIBLE_BUCKET).s3.$(S3_COMPATIBLE_REGION).amazonaws.com",
				},
			},
		},
		"hdfs": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				HDFS: &risingwavev1alpha1.RisingWaveObjectStorageHDFS{
					NameNode: "name-node",
					Root:     "root",
				},
			},
			envs: []corev1.EnvVar{{
				Name:  "RW_STATE_STORE",
				Value: "hummock+hdfs://name-node@root",
			}},
		},
		"webhdfs": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				WebHDFS: &risingwavev1alpha1.RisingWaveObjectStorageHDFS{
					NameNode: "name-node",
					Root:     "root",
				},
			},
			hummockArg: "hummock+webhdfs://name-node@root",
			envs:       []corev1.EnvVar{},
		},
	}
}

type configMapTestCase struct {
	risingwave *risingwavev1alpha1.RisingWave
	configVal  string
}

func configMapTestCases() map[string]configMapTestCase {
	return map[string]configMapTestCase{
		"empty-val": {
			configVal: "",
		},
		"non-empty-val": {
			configVal: "a",
		},
	}
}

type computeArgsTestCase struct {
	cpuLimit int64
	memLimit int64
	envList  []corev1.EnvVar
}

func computeEnvsTestCases() map[string]computeArgsTestCase {
	return map[string]computeArgsTestCase{
		"empty-limits": {},
		"cpu-limit-4": {
			cpuLimit: 4,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_PARALLELISM",
					Value: "4",
				},
			},
		},
		"mem-limit-4g": {
			memLimit: 4 << 30,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(4 << 30),
				},
			},
		},
		"mem-limit-1g": {
			memLimit: 1 << 30,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(1 << 30),
				},
			},
		},
		"mem-limit-768m": {
			memLimit: 768 << 20,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(768 << 20),
				},
			},
		},
		"mem-limit-512m": {
			memLimit: 512 << 20,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(512 << 20),
				},
			},
		},
		"mem-limit-256m": {
			memLimit: 256 << 20,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(256 << 20),
				},
			},
		},
		"cpu-and-mem": {
			cpuLimit: 4,
			memLimit: 1 << 30,
			envList: []corev1.EnvVar{
				{
					Name:  "RW_PARALLELISM",
					Value: "4",
				},
				{
					Name:  "RW_TOTAL_MEMORY_BYTES",
					Value: strconv.Itoa(1 << 30),
				},
			},
		},
	}
}

type inheritedLabelsTestCase struct {
	labels             map[string]string
	inheritPrefixValue string
	inheritedLabels    map[string]string
}

func inheritedLabelsTestCases() map[string]inheritedLabelsTestCase {
	return map[string]inheritedLabelsTestCase{
		"no-inherit": {
			labels: map[string]string{
				"a":                               "b",
				"risingwave.risingwavelabs.com/a": "b",
			},
			inheritPrefixValue: "",
			inheritedLabels:    nil,
		},
		"inherit-with-one-prefix": {
			labels: map[string]string{
				"a":                               "b",
				"risingwave.risingwavelabs.com/a": "b",
			},
			inheritPrefixValue: "risingwave.risingwavelabs.com",
			inheritedLabels: map[string]string{
				"risingwave.risingwavelabs.com/a": "b",
			},
		},
		"inherit-with-two-prefixes": {
			labels: map[string]string{
				"a":                               "b",
				"risingwave.risingwavelabs.com/a": "b",
				"risingwave.cloud/c":              "d",
			},
			inheritPrefixValue: "risingwave.risingwavelabs.com,risingwave.cloud",
			inheritedLabels: map[string]string{
				"risingwave.risingwavelabs.com/a": "b",
				"risingwave.cloud/c":              "d",
			},
		},
		"won't-inherit-builtin-prefix": {
			labels: map[string]string{
				"a":                               "b",
				"risingwave/c":                    "d",
				"risingwave.risingwavelabs.com/a": "b",
			},
			inheritPrefixValue: "risingwave.risingwavelabs.com,risingwave",
			inheritedLabels: map[string]string{
				"risingwave.risingwavelabs.com/a": "b",
			},
		},
	}
}

type metaStorageTestCase struct {
	metaStorage risingwavev1alpha1.RisingWaveMetaStorage
	envs        []corev1.EnvVar
}

func metaStorageTestCases() map[string]metaStorageTestCase {
	return map[string]metaStorageTestCase{
		"memory": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Memory: pointer.Bool(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "mem",
				},
			},
		},
		"etcd-no-auth": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
					Endpoint: "etcd:1234",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "etcd",
				},
				{
					Name:  "RW_ETCD_ENDPOINTS",
					Value: "etcd:1234",
				},
			},
		},
		"etcd-auth": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
					Endpoint: "etcd:1234",
					Secret:   "etcd-credentials",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "etcd",
				},
				{
					Name:  "RW_ETCD_ENDPOINTS",
					Value: "etcd:1234",
				},
				{
					Name: "RW_ETCD_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "etcd-credentials",
							},
							Key: consts.SecretKeyEtcdUsername,
						},
					},
				},
				{
					Name: "RW_ETCD_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "etcd-credentials",
							},
							Key: consts.SecretKeyEtcdPassword,
						},
					},
				},
			},
		},
	}
}
