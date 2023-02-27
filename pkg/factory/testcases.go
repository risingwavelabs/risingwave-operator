package factory

import (
	"math"
	"time"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
)

func GetDeploymentTestcases() map[string]deploymentTestcase {
	return map[string]deploymentTestcase{
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
					Template: *NewPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
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

func GetSTSTestcases() map[string]stsTestcase {
	return map[string]stsTestcase{
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
					Template: *NewPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
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

func GetClonesetTestcases() map[string]clonesetTestcase {
	return map[string]clonesetTestcase{
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
					Template: *NewPodTemplate(func(t *risingwavev1alpha1.RisingWavePodTemplateSpec) {
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

func GetAdvancedSTSTestcases() map[string]advancedSTSTestcase {
	return map[string]advancedSTSTestcase{
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
