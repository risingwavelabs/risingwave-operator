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
	"k8s.io/utils/ptr"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type risingWaveComponentGroup interface {
	risingwavev1alpha1.RisingWaveNodeGroup
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
		stateStoresTestCase |
		configMapTestCase |
		computeArgsTestCase |
		metaStoreTestCase |
		metaStatefulSetTestCase |
		metaAdvancedSTSTestCase |
		tlsTestcase
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
		*prometheusv1.ServiceMonitor |
		*corev1.PodTemplateSpec
}

type baseTestCase struct {
	risingwave *risingwavev1alpha1.RisingWave
}

type testCase[T risingWaveComponentGroup, K kubeObjectsUpgradeStrategy] struct {
	baseTestCase
	component               string
	group                   T
	expectedUpgradeStrategy K
	restartAt               *metav1.Time
}

type deploymentTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *appsv1.DeploymentStrategy]

func deploymentTestCases() map[string]deploymentTestCase {
	return map[string]deploymentTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				},
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type metaStatefulSetTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *appsv1.StatefulSetUpdateStrategy]

func metaStatefulSetTestCases() map[string]metaStatefulSetTestCase {
	return map[string]metaStatefulSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				},
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "100%"},
				},
			},
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type computeStatefulSetTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *appsv1.StatefulSetUpdateStrategy]

func computeStatefulSetTestCases() map[string]computeStatefulSetTestCase {
	return map[string]computeStatefulSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
		},
		"termination-grace-period-seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart-at": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				},
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			expectedUpgradeStrategy: &appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
					MaxUnavailable: &intstr.IntOrString{Type: intstr.String, StrVal: "100%"},
				},
			},
		},
		"upgrade-strategy-rolling-update-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type cloneSetTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *kruiseappsv1alpha1.CloneSetUpdateStrategy]

func cloneSetTestCases() map[string]cloneSetTestCase {
	return map[string]cloneSetTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-Recreate-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxSurge: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-InPlaceOnly-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxSurge: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1alpha1.CloneSetUpdateStrategy{
				Type: kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType,
			},
		},
		"upgrade-strategy-InPlaceIfPossible-max-unavailable-50%": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxSurge: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type metaAdvancedSTSTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *kruiseappsv1beta1.StatefulSetUpdateStrategy]

func metaAdvancedSTSTestCases() map[string]metaAdvancedSTSTestCase {
	return map[string]metaAdvancedSTSTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
					Partition:       ptr.To(int32(50)),
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					Partition:       ptr.To(int32(50)),
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type computeAdvancedSTSTestCase testCase[risingwavev1alpha1.RisingWaveNodeGroup, *kruiseappsv1beta1.StatefulSetUpdateStrategy]

func computeAdvancedSTSTestCases() map[string]computeAdvancedSTSTestCase {
	return map[string]computeAdvancedSTSTestCase{
		"pods-meta-labels": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"pods-meta-annotations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Annotations: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"default-group": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"node-selectors": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					ObjectMeta: risingwavev1alpha1.PartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						NodeSelector: map[string]string{
							"a": "b",
						},
					},
				},
			},
		},
		"tolerations": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						PriorityClassName: "high-priority",
					},
				},
			},
		},
		"security-context": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						TerminationGracePeriodSeconds: ptr.To(int64(5)),
					},
				},
			},
		},
		"with-group-name": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
		},
		"with-restart": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-policy-always": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image:           rand.String(20),
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			restartAt: &metav1.Time{Time: time.Now()},
		},
		"image-pull-secrets": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "test-group",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{Name: "a"},
						},
					},
				},
			},
		},
		"upgrade-strategy-Recreate": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 50,
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType,
					Partition:       ptr.To(int32(50)),
				},
			},
		},
		"upgrade-strategy-InPlaceOnly-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						MaxUnavailable: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
						},
					},
				},
			},
			expectedUpgradeStrategy: &kruiseappsv1beta1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
				RollingUpdate: &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
					PodUpdatePolicy: kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType,
					Partition:       ptr.To(int32(50)),
				},
			},
		},
		"upgrade-strategy-InPlaceIfPossible-Grace-Period-20seconds": {
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
							Image: rand.String(20),
						},
					},
				},
				UpgradeStrategy: risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{
						GracePeriodSeconds: 20,
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
			group: risingwavev1alpha1.RisingWaveNodeGroup{
				Name:     "",
				Replicas: int32(rand.Intn(math.MaxInt32)),
				Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
					Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
						RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
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
		},
	}
}

type servicesTestCase struct {
	baseTestCase
	component            string
	selectorComponent    string // Empty means equals to component.
	ports                map[string]int32
	enableStandaloneMode bool
	globalServiceType    corev1.ServiceType
	expectServiceType    corev1.ServiceType
}

func servicesTestCases() map[string]servicesTestCase {
	return map[string]servicesTestCase{
		"meta-ports": {
			component:         consts.ComponentMeta,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService:   consts.MetaServicePort,
				consts.PortMetrics:   consts.MetaMetricsPort,
				consts.PortDashboard: consts.MetaDashboardPort,
			},
		},
		"meta-ports-node-port": {
			component:         consts.ComponentMeta,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
		},
		"frontend-ports": {
			component:         consts.ComponentFrontend,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: consts.FrontendServicePort,
				consts.PortMetrics: consts.FrontendMetricsPort,
			},
		},
		"frontend-ports-node-port": {
			component:         consts.ComponentFrontend,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeNodePort,
		},
		"compute-ports": {
			component:         consts.ComponentCompute,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: consts.ComputeServicePort,
				consts.PortMetrics: consts.ComputeMetricsPort,
			},
		},
		"compute-ports-node-port": {
			component:         consts.ComponentCompute,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
		},
		"compactor-ports": {
			component:         consts.ComponentCompactor,
			globalServiceType: corev1.ServiceTypeClusterIP,
			expectServiceType: corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService: consts.CompactorServicePort,
				consts.PortMetrics: consts.CompactorMetricsPort,
			},
		},
		"compactor-ports-node-port": {
			component:         consts.ComponentCompactor,
			globalServiceType: corev1.ServiceTypeNodePort,
			expectServiceType: corev1.ServiceTypeClusterIP,
		},
		"standalone-ports": {
			component:            consts.ComponentStandalone,
			enableStandaloneMode: true,
			globalServiceType:    corev1.ServiceTypeNodePort,
			expectServiceType:    corev1.ServiceTypeClusterIP,
			ports: map[string]int32{
				consts.PortService:   consts.FrontendServicePort,
				consts.PortMetrics:   consts.MetaMetricsPort,
				consts.PortDashboard: consts.MetaDashboardPort,
			},
		},
		"standalone-frontend-ports": {
			component:            consts.ComponentFrontend,
			selectorComponent:    consts.ComponentStandalone,
			enableStandaloneMode: true,
			globalServiceType:    corev1.ServiceTypeNodePort,
			expectServiceType:    corev1.ServiceTypeNodePort,
			ports: map[string]int32{
				consts.PortService: consts.FrontendServicePort,
			},
		},
	}

}

type serviceMetadataTestCase struct {
	baseTestCase
	component         string
	globalServiceMeta risingwavev1alpha1.PartialObjectMeta
}

func serviceMetadataTestCases() map[string]serviceMetadataTestCase {
	return map[string]serviceMetadataTestCase{
		"random-meta-labels": {
			component: consts.ComponentMeta,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-meta-annotations": {
			component: consts.ComponentMeta,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-frontend-labels": {
			component: consts.ComponentFrontend,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-frontend-annotations": {
			component: consts.ComponentFrontend,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compute-labels": {
			component: consts.ComponentCompute,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compute-annotations": {
			component: consts.ComponentCompute,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compactor-labels": {
			component: consts.ComponentCompactor,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Labels: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
		"random-compactor-annotations": {
			component: consts.ComponentCompactor,
			globalServiceMeta: risingwavev1alpha1.PartialObjectMeta{
				Annotations: map[string]string{
					"key1": "value1",
					"key2": "value2",
				},
			},
		},
	}
}

type stateStoresTestCase struct {
	baseTestCase
	stateStore risingwavev1alpha1.RisingWaveStateStoreBackend
	envs       []corev1.EnvVar
}

func stateStoreTestCases() map[string]stateStoresTestCase {
	return map[string]stateStoresTestCase{
		"empty_data_directory": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "",
				Memory:        ptr.To(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_DATA_DIRECTORY",
					Value: "",
				},
			},
		},
		"some_data_directory": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "1234",
				Memory:        ptr.To(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_DATA_DIRECTORY",
					Value: "1234",
				},
			},
		},
		"memory": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				Memory: ptr.To(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+memory",
				},
			},
		},
		"minio": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{
					Endpoint: "minio-endpoint:1234",
					Bucket:   "minio-hummock01",
					RisingWaveMinIOCredentials: risingwavev1alpha1.RisingWaveMinIOCredentials{
						SecretName:     "minio-creds",
						UsernameKeyRef: consts.SecretKeyMinIOUsername,
						PasswordKeyRef: consts.SecretKeyMinIOPassword,
					},
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
							Key: consts.SecretKeyMinIOUsername,
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
							Key: consts.SecretKeyMinIOPassword,
						},
					},
				},
			},
		},
		"s3": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket: "s3-hummock01",
					Region: "ap-southeast-1",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "s3-creds",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
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
		"gcs-workload": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
					RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
						UseWorkloadIdentity: ptr.To(true),
					},
					Bucket: "gcs-bucket",
					Root:   "gcs-root",
				},
			},
			envs: []corev1.EnvVar{{
				Name:  "RW_STATE_STORE",
				Value: "hummock+gcs://gcs-bucket@gcs-root",
			}},
		},
		"gcs-secret": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
					Bucket: "gcs-bucket",
					Root:   "gcs-root",
					RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
						SecretName:                      "gcs-creds",
						ServiceAccountCredentialsKeyRef: consts.SecretKeyGCSServiceAccountCredentials,
					},
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
		"aliyun-oss-not-internal": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				AliyunOSS: &risingwavev1alpha1.RisingWaveStateStoreBackendAliyunOSS{
					Bucket:           "aliyun-oss-hummock01",
					Root:             "aliyun-oss-root",
					Region:           "cn-hangzhou",
					InternalEndpoint: false,
					RisingWaveAliyunOSSCredentials: risingwavev1alpha1.RisingWaveAliyunOSSCredentials{
						SecretName:         "aliyun-oss-creds",
						AccessKeyIDRef:     consts.SecretKeyAliyunOSSAccessKeyID,
						AccessKeySecretRef: consts.SecretKeyAliyunOSSAccessKeySecret,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+oss://aliyun-oss-hummock01@aliyun-oss-root",
				},
				{
					Name: "OSS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "aliyun-oss-creds",
							},
							Key: consts.SecretKeyAliyunOSSAccessKeyID,
						},
					},
				},
				{
					Name: "OSS_ACCESS_KEY_SECRET",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "aliyun-oss-creds",
							},
							Key: consts.SecretKeyAliyunOSSAccessKeySecret,
						},
					},
				},
				{
					Name:  "OSS_ENDPOINT",
					Value: "https://oss-$(OSS_REGION).aliyuncs.com",
				},
			},
		},
		"aliyun-oss-internal": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				AliyunOSS: &risingwavev1alpha1.RisingWaveStateStoreBackendAliyunOSS{
					Bucket:           "aliyun-oss-hummock01",
					Root:             "aliyun-oss-root",
					Region:           "cn-hangzhou",
					InternalEndpoint: true,
					RisingWaveAliyunOSSCredentials: risingwavev1alpha1.RisingWaveAliyunOSSCredentials{
						SecretName:         "aliyun-oss-creds",
						AccessKeyIDRef:     consts.SecretKeyAliyunOSSAccessKeyID,
						AccessKeySecretRef: consts.SecretKeyAliyunOSSAccessKeySecret,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+oss://aliyun-oss-hummock01@aliyun-oss-root",
				},
				{
					Name: "OSS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "aliyun-oss-creds",
							},
							Key: consts.SecretKeyAliyunOSSAccessKeyID,
						},
					},
				},
				{
					Name: "OSS_ACCESS_KEY_SECRET",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "aliyun-oss-creds",
							},
							Key: consts.SecretKeyAliyunOSSAccessKeySecret,
						},
					},
				},
				{
					Name:  "OSS_ENDPOINT",
					Value: "https://oss-$(OSS_REGION)-internal.aliyuncs.com",
				},
			},
		},
		"s3-compatible": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket:   "s3-hummock01",
					Endpoint: "oss-cn-hangzhou.aliyuncs.com",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "s3-creds",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
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
					Name:  "RW_S3_ENDPOINT",
					Value: "https://oss-cn-hangzhou.aliyuncs.com",
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
		"s3-compatible-virtual-hosted-style": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket:   "s3-hummock01",
					Endpoint: "https://${BUCKET}.oss-cn-hangzhou.aliyuncs.com",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "s3-creds",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
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
					Name:  "RW_S3_ENDPOINT",
					Value: "https://$(AWS_S3_BUCKET).oss-cn-hangzhou.aliyuncs.com",
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
		"s3-with-region": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket: "s3-hummock01",
					Region: "ap-southeast-1",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "s3-creds",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
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
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket:   "s3-hummock01",
					Endpoint: "s3.${REGION}.amazonaws.com",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3://s3-hummock01",
				},
				{
					Name:  "RW_S3_ENDPOINT",
					Value: "https://s3.$(AWS_REGION).amazonaws.com",
				},
			},
		},
		"endpoint-with-bucket-variable": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
					Bucket:   "s3-hummock01",
					Endpoint: "${BUCKET}.s3.${REGION}.amazonaws.com",
					RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
						SecretName:         "",
						AccessKeyRef:       consts.SecretKeyAWSS3AccessKeyID,
						SecretAccessKeyRef: consts.SecretKeyAWSS3SecretAccessKey,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+s3://s3-hummock01",
				},
				{
					Name:  "RW_S3_ENDPOINT",
					Value: "https://$(AWS_S3_BUCKET).s3.$(AWS_REGION).amazonaws.com",
				},
			},
		},
		"azure-blob": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{
					Container: "azure-blob-hummock01",
					Root:      "/azure-blob-root",
					Endpoint:  "https://accountName.blob.core.windows.net",
					RisingWaveAzureBlobCredentials: risingwavev1alpha1.RisingWaveAzureBlobCredentials{
						SecretName:     "azure-blob-creds",
						AccountNameRef: consts.SecretKeyAzureBlobAccountName,
						AccountKeyRef:  consts.SecretKeyAzureBlobAccountKey,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+azblob://azure-blob-hummock01@/azure-blob-root",
				},
				{
					Name: "AZBLOB_ACCOUNT_NAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "azure-blob-creds",
							},
							Key: consts.SecretKeyAzureBlobAccountName,
						},
					},
				},
				{
					Name: "AZBLOB_ACCOUNT_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "azure-blob-creds",
							},
							Key: consts.SecretKeyAzureBlobAccountKey,
						},
					},
				},
				{
					Name:  "AZBLOB_ENDPOINT",
					Value: "https://accountName.blob.core.windows.net",
				},
			},
		},
		"azure-blob-use-sa": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{
					Container: "azure-blob-hummock01",
					Root:      "/azure-blob-root",
					Endpoint:  "https://accountName.blob.core.windows.net",
					RisingWaveAzureBlobCredentials: risingwavev1alpha1.RisingWaveAzureBlobCredentials{
						SecretName:        "azure-blob-creds",
						UseServiceAccount: ptr.To(true),
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+azblob://azure-blob-hummock01@/azure-blob-root",
				},
				{
					Name:  "AZBLOB_ENDPOINT",
					Value: "https://accountName.blob.core.windows.net",
				},
			},
		},
		"hdfs": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				HDFS: &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{
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
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				WebHDFS: &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{
					NameNode: "name-node",
					Root:     "root",
				},
			},
			envs: []corev1.EnvVar{{
				Name:  "RW_STATE_STORE",
				Value: "hummock+webhdfs://name-node@root",
			}},
		},
		"local-disk": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				LocalDisk: &risingwavev1alpha1.RisingWaveStateStoreBackendLocalDisk{
					Root: "root",
				},
			},
			envs: []corev1.EnvVar{{
				Name:  "RW_STATE_STORE",
				Value: "hummock+fs://root",
			}},
		},
		"huawei-cloud-obs": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				HuaweiCloudOBS: &risingwavev1alpha1.RisingWaveStateStoreBackendHuaweiCloudOBS{
					Bucket: "obs-hummock01",
					Region: "ap-southeast-2",
					RisingWaveHuaweiCloudOBSCredentials: risingwavev1alpha1.RisingWaveHuaweiCloudOBSCredentials{
						SecretName:         "obs-creds",
						AccessKeyIDRef:     consts.SecretKeyHuaweiCloudOBSAccessKeyID,
						AccessKeySecretRef: consts.SecretKeyHuaweiCloudOBSAccessKeySecret,
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_STATE_STORE",
					Value: "hummock+obs://obs-hummock01",
				},
				{
					Name:  "OBS_REGION",
					Value: "ap-southeast-2",
				},
				{
					Name: "OBS_ACCESS_KEY_ID",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "obs-creds",
							},
							Key: consts.SecretKeyHuaweiCloudOBSAccessKeyID,
						},
					},
				},
				{
					Name: "OBS_SECRET_ACCESS_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "obs-creds",
							},
							Key: consts.SecretKeyHuaweiCloudOBSAccessKeySecret,
						},
					},
				},
				{
					Name:  "OBS_ENDPOINT",
					Value: "https://obs.$(OBS_REGION).myhuaweicloud.com",
				},
			},
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

type metaStoreTestCase struct {
	metaStore risingwavev1alpha1.RisingWaveMetaStoreBackend
	envs      []corev1.EnvVar
}

func metaStoreTestCases() map[string]metaStoreTestCase {
	return map[string]metaStoreTestCase{
		"memory": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				Memory: ptr.To(true),
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "mem",
				},
			},
		},
		"etcd-no-auth": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
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
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
					Endpoint: "etcd:1234",
					RisingWaveEtcdCredentials: &risingwavev1alpha1.RisingWaveEtcdCredentials{
						SecretName:     "etcd-credentials",
						UsernameKeyRef: consts.SecretKeyEtcdUsername,
						PasswordKeyRef: consts.SecretKeyEtcdPassword,
					},
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
		"sqlite": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				SQLite: &risingwavev1alpha1.RisingWaveMetaStoreBackendSQLite{
					Path: "/data/risingwave.db",
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "sql",
				},
				{
					Name:  "RW_SQL_ENDPOINT",
					Value: "sqlite:///data/risingwave.db?mode=rwc",
				},
			},
		},
		"mysql-no-options": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				MySQL: &risingwavev1alpha1.RisingWaveMetaStoreBackendMySQL{
					Host:     "mysql",
					Port:     3307,
					Database: "risingwave",
					RisingWaveDBCredentials: risingwavev1alpha1.RisingWaveDBCredentials{
						SecretName:     "s",
						UsernameKeyRef: "username",
						PasswordKeyRef: "password",
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "sql",
				},
				{
					Name:  "RW_SQL_ENDPOINT",
					Value: "mysql://$(RW_MYSQL_USERNAME):$(RW_MYSQL_PASSWORD)@mysql:3307/risingwave",
				},
				{
					Name: "RW_MYSQL_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
				{
					Name: "RW_MYSQL_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
			},
		},
		"mysql-with-options": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				MySQL: &risingwavev1alpha1.RisingWaveMetaStoreBackendMySQL{
					Host:     "mysql",
					Port:     3307,
					Database: "risingwave",
					Options: map[string]string{
						"a": "b",
						"c": "d=e",
					},
					RisingWaveDBCredentials: risingwavev1alpha1.RisingWaveDBCredentials{
						SecretName:     "s",
						UsernameKeyRef: "username",
						PasswordKeyRef: "password",
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "sql",
				},
				{
					Name:  "RW_SQL_ENDPOINT",
					Value: "mysql://$(RW_MYSQL_USERNAME):$(RW_MYSQL_PASSWORD)@mysql:3307/risingwave?a=b&c=d%3De",
				},
				{
					Name: "RW_MYSQL_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
				{
					Name: "RW_MYSQL_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
			},
		},
		"pg-no-options": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				PostgreSQL: &risingwavev1alpha1.RisingWaveMetaStoreBackendPostgreSQL{
					Host:     "postgresql",
					Port:     3307,
					Database: "risingwave",
					RisingWaveDBCredentials: risingwavev1alpha1.RisingWaveDBCredentials{
						SecretName:     "s",
						UsernameKeyRef: "username",
						PasswordKeyRef: "password",
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "sql",
				},
				{
					Name:  "RW_SQL_ENDPOINT",
					Value: "postgres://$(RW_POSTGRES_USERNAME):$(RW_POSTGRES_PASSWORD)@postgresql:3307/risingwave",
				},
				{
					Name: "RW_POSTGRES_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
				{
					Name: "RW_POSTGRES_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
			},
		},
		"pg-with-options": {
			metaStore: risingwavev1alpha1.RisingWaveMetaStoreBackend{
				PostgreSQL: &risingwavev1alpha1.RisingWaveMetaStoreBackendPostgreSQL{
					Host:     "postgresql",
					Port:     3307,
					Database: "risingwave",
					Options: map[string]string{
						"a": "b",
						"c": "d=e",
					},
					RisingWaveDBCredentials: risingwavev1alpha1.RisingWaveDBCredentials{
						SecretName:     "s",
						UsernameKeyRef: "username",
						PasswordKeyRef: "password",
					},
				},
			},
			envs: []corev1.EnvVar{
				{
					Name:  "RW_BACKEND",
					Value: "sql",
				},
				{
					Name:  "RW_SQL_ENDPOINT",
					Value: "postgres://$(RW_POSTGRES_USERNAME):$(RW_POSTGRES_PASSWORD)@postgresql:3307/risingwave?a=b&c=d%3De",
				},
				{
					Name: "RW_POSTGRES_USERNAME",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "username",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
				{
					Name: "RW_POSTGRES_PASSWORD",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							Key: "password",
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "s",
							},
						},
					},
				},
			},
		},
	}
}

type tlsTestcase struct {
	standalone     bool
	tls            *risingwavev1alpha1.RisingWaveTLSConfiguration
	expectedEnvs   []corev1.EnvVar
	unexpectedEnvs []string
}

func tlsTestcases() map[string]tlsTestcase {
	return map[string]tlsTestcase{
		"tls-disabled-nil": {
			tls: nil,
			unexpectedEnvs: []string{
				"RW_SSL_KEY",
				"RW_SSL_CERT",
			},
		},
		"tls-disabled-empty": {
			tls: &risingwavev1alpha1.RisingWaveTLSConfiguration{
				SecretName: "",
			},
			unexpectedEnvs: []string{
				"RW_SSL_KEY",
				"RW_SSL_CERT",
			},
		},
		"tls-enabled": {
			tls: &risingwavev1alpha1.RisingWaveTLSConfiguration{
				SecretName: "tls",
			},
			expectedEnvs: []corev1.EnvVar{
				{
					Name: "RW_SSL_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "tls",
							},
							Key: "tls.key",
						},
					},
				},
				{
					Name: "RW_SSL_CERT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "tls",
							},
							Key: "tls.crt",
						},
					},
				},
			},
		},
		"tls-disabled-nil-standalone": {
			standalone: true,
			tls:        nil,
			unexpectedEnvs: []string{
				"RW_SSL_KEY",
				"RW_SSL_CERT",
			},
		},
		"tls-disabled-empty-standalone": {
			standalone: true,
			tls: &risingwavev1alpha1.RisingWaveTLSConfiguration{
				SecretName: "",
			},
			unexpectedEnvs: []string{
				"RW_SSL_KEY",
				"RW_SSL_CERT",
			},
		},
		"tls-enabled-standalone": {
			standalone: true,
			tls: &risingwavev1alpha1.RisingWaveTLSConfiguration{
				SecretName: "tls",
			},
			expectedEnvs: []corev1.EnvVar{
				{
					Name: "RW_SSL_KEY",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "tls",
							},
							Key: "tls.key",
						},
					},
				},
				{
					Name: "RW_SSL_CERT",
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{
								Name: "tls",
							},
							Key: "tls.crt",
						},
					},
				},
			},
		},
	}
}
