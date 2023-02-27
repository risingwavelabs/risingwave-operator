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
	"strconv"
	"strings"
	"testing"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

// func Test_RisingWaveObjectFactory_Services(t *testing.T) {
// 	testcases := map[string]struct {
// 		component         string
// 		ports             map[string]int32
// 		globalServiceType corev1.ServiceType
// 		expectServiceType corev1.ServiceType
// 	}{
// 		"random-meta-ports": {
// 			component:         consts.ComponentMeta,
// 			globalServiceType: corev1.ServiceTypeClusterIP,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService:   int32(rand.Int() & 0xffff),
// 				consts.PortMetrics:   int32(rand.Int() & 0xffff),
// 				consts.PortDashboard: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-meta-ports-node-port": {
// 			component:         consts.ComponentMeta,
// 			globalServiceType: corev1.ServiceTypeNodePort,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService:   int32(rand.Int() & 0xffff),
// 				consts.PortMetrics:   int32(rand.Int() & 0xffff),
// 				consts.PortDashboard: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-frontend-ports": {
// 			component:         consts.ComponentFrontend,
// 			globalServiceType: corev1.ServiceTypeClusterIP,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-frontend-ports-node-port": {
// 			component:         consts.ComponentFrontend,
// 			globalServiceType: corev1.ServiceTypeNodePort,
// 			expectServiceType: corev1.ServiceTypeNodePort,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-compute-ports": {
// 			component:         consts.ComponentCompute,
// 			globalServiceType: corev1.ServiceTypeClusterIP,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-compute-ports-node-port": {
// 			component:         consts.ComponentCompute,
// 			globalServiceType: corev1.ServiceTypeNodePort,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-compactor-ports": {
// 			component:         consts.ComponentCompactor,
// 			globalServiceType: corev1.ServiceTypeClusterIP,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-compactor-ports-node-port": {
// 			component:         consts.ComponentCompactor,
// 			globalServiceType: corev1.ServiceTypeNodePort,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-connector-ports": {
// 			component:         consts.ComponentConnector,
// 			globalServiceType: corev1.ServiceTypeClusterIP,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 		"random-connector-ports-node-port": {
// 			component:         consts.ComponentConnector,
// 			globalServiceType: corev1.ServiceTypeNodePort,
// 			expectServiceType: corev1.ServiceTypeClusterIP,
// 			ports: map[string]int32{
// 				consts.PortService: int32(rand.Int() & 0xffff),
// 				consts.PortMetrics: int32(rand.Int() & 0xffff),
// 			},
// 		},
// 	}

// 	for name, tc := range testcases {
// 		t.Run(name, func(t *testing.T) {
// 			risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
// 				r.Spec.Global.ServiceType = tc.globalServiceType

// 				switch tc.component {
// 				case consts.ComponentMeta:
// 					SetupMetaPorts(r, tc.ports)
// 				case consts.ComponentFrontend:
// 					SetupFrontendPorts(r, tc.ports)
// 				case consts.ComponentCompute:
// 					SetupComputePorts(r, tc.ports)
// 				case consts.ComponentCompactor:
// 					SetupCompactorPorts(r, tc.ports)
// 				case consts.ComponentConnector:
// 					SetupConnectorPorts(r, tc.ports)
// 				}
// 			})

// 			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

// 			var svc *corev1.Service
// 			switch tc.component {
// 			case consts.ComponentMeta:
// 				svc = factory.NewMetaService()
// 			case consts.ComponentFrontend:
// 				svc = factory.NewFrontendService()
// 			case consts.ComponentCompute:
// 				svc = factory.NewComputeService()
// 			case consts.ComponentCompactor:
// 				svc = factory.NewCompactorService()
// 			case consts.ComponentConnector:
// 				svc = factory.NewConnectorService()
// 			default:
// 				t.Fatal("bad test")
// 			}

// 			ComposeAssertions(
// 				NewObjectAssert(svc, "controlled-by-risingwave", func(obj *corev1.Service) bool {
// 					return ControlledBy(risingwave, obj)
// 				}),
// 				NewObjectAssert(svc, "namespace-equals", func(obj *corev1.Service) bool {
// 					return obj.Namespace == risingwave.Namespace
// 				}),
// 				NewObjectAssert(svc, "ports-equal", func(obj *corev1.Service) bool {
// 					return HasTCPServicePorts(obj, tc.ports)
// 				}),
// 				NewObjectAssert(svc, "service-type-match", func(obj *corev1.Service) bool {
// 					return IsServiceType(obj, tc.expectServiceType)
// 				}),
// 				NewObjectAssert(svc, "service-labels-match", func(obj *corev1.Service) bool {
// 					return HasLabels(obj, ComponentLabels(risingwave, tc.component, true), true)
// 				}),
// 				NewObjectAssert(svc, "selector-equals", func(obj *corev1.Service) bool {
// 					return HasServiceSelector(obj, PodSelector(risingwave, tc.component, nil))
// 				}),
// 			).Assert(t)
// 		})
// 	}
// }

// func Test_RisingWaveObjectFactory_ServicesMeta(t *testing.T) {
// 	testcases := map[string]struct {
// 		component         string
// 		globalServiceMeta risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta
// 	}{
// 		"random-meta-labels": {
// 			component: consts.ComponentMeta,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Labels: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-meta-annotations": {
// 			component: consts.ComponentMeta,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Annotations: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-frontend-labels": {
// 			component: consts.ComponentFrontend,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Labels: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-frontend-annotations": {
// 			component: consts.ComponentFrontend,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Annotations: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-compute-labels": {
// 			component: consts.ComponentCompute,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Labels: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-compute-annotations": {
// 			component: consts.ComponentCompute,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Annotations: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-compactor-labels": {
// 			component: consts.ComponentCompactor,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Labels: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-compactor-annotations": {
// 			component: consts.ComponentCompactor,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Annotations: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-connector-labels": {
// 			component: consts.ComponentConnector,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Labels: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 		"random-connector-annotations": {
// 			component: consts.ComponentConnector,
// 			globalServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
// 				Annotations: map[string]string{
// 					"key1": "value1",
// 					"key2": "value2",
// 				},
// 			},
// 		},
// 	}

// 	for name, tc := range testcases {
// 		t.Run(name, func(t *testing.T) {
// 			risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
// 				r.Spec.Global.ServiceMeta = tc.globalServiceMeta
// 			})

// 			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

// 			var svc *corev1.Service
// 			switch tc.component {
// 			case consts.ComponentMeta:
// 				svc = factory.NewMetaService()
// 			case consts.ComponentFrontend:
// 				svc = factory.NewFrontendService()
// 			case consts.ComponentCompute:
// 				svc = factory.NewComputeService()
// 			case consts.ComponentCompactor:
// 				svc = factory.NewCompactorService()
// 			case consts.ComponentConnector:
// 				svc = factory.NewConnectorService()
// 			default:
// 				t.Fatal("bad test")
// 			}

// 			ComposeAssertions(
// 				NewObjectAssert(svc, "service-labels-match", func(obj *corev1.Service) bool {
// 					return HasLabels(obj, ComponentLabels(risingwave, tc.component, true), true)
// 				}),
// 				NewObjectAssert(svc, "service-annotations-match", func(obj *corev1.Service) bool {
// 					return HasAnnotations(obj, ComponentAnnotations(risingwave, tc.component), true)
// 				}),
// 			).Assert(t)
// 		})
// 	}
// }

// func Test_RisingWaveObjectFactory_ConfigMaps(t *testing.T) {
// 	testcases := map[string]struct {
// 		configVal string
// 	}{
// 		"empty-val": {
// 			configVal: "",
// 		},
// 		"non-empty-val": {
// 			configVal: "a",
// 		},
// 	}

// 	for name, tc := range testcases {
// 		t.Run(name, func(t *testing.T) {
// 			risingwave := NewTestRisingwave()

// 			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

// 			cm := factory.NewConfigConfigMap(tc.configVal)

// 			ComposeAssertions(
// 				NewObjectAssert(cm, "controlled-by-risingwave", func(obj *corev1.ConfigMap) bool {
// 					return ControlledBy(risingwave, obj)
// 				}),
// 				NewObjectAssert(cm, "namespace-equals", func(obj *corev1.ConfigMap) bool {
// 					return obj.Namespace == risingwave.Namespace
// 				}),
// 				NewObjectAssert(cm, "configmap-labels-match", func(obj *corev1.ConfigMap) bool {
// 					return HasLabels(obj, ComponentLabels(risingwave, consts.ComponentConfig, false), true)
// 				}),
// 				NewObjectAssert(cm, "configmap-data-match", func(obj *corev1.ConfigMap) bool {
// 					return MapEquals(obj.Data, map[string]string{
// 						risingWaveConfigMapKey: lo.If(tc.configVal == "", "").Else(tc.configVal),
// 					})
// 				}),
// 			).Assert(t)
// 		})
// 	}
// }

func Test_RisingWaveObjectFactory_Deployments(t *testing.T) {
	testcases := GetDeploymentTestcases()
	deploymentPreds := GetDeploymentPredicates()
	for _, component := range []string{consts.ComponentFrontend, consts.ComponentCompactor} {
		for name, tc := range testcases {
			risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta.Memory = pointer.Bool(true)
				r.Spec.Storages.Object.Memory = pointer.Bool(true)
				if tc.group.Name == "" {
					r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
					switch component {
					case consts.ComponentMeta:
						r.Spec.Global.Replicas.Meta = tc.group.Replicas
					case consts.ComponentFrontend:
						r.Spec.Global.Replicas.Frontend = tc.group.Replicas
					case consts.ComponentCompactor:
						r.Spec.Global.Replicas.Compactor = tc.group.Replicas
					}
				} else {
					switch component {
					case consts.ComponentMeta:
						r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
							tc.group,
						}
					case consts.ComponentFrontend:
						r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
							tc.group,
						}
						r.Spec.Components.Frontend.RestartAt = tc.restartAt
					case consts.ComponentCompactor:
						r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
							tc.group,
						}
						r.Spec.Components.Compactor.RestartAt = tc.restartAt
					}
				}

				switch component {
				case consts.ComponentMeta:
					r.Spec.Components.Meta.RestartAt = tc.restartAt
				case consts.ComponentFrontend:
					r.Spec.Components.Frontend.RestartAt = tc.restartAt
				case consts.ComponentCompactor:
					r.Spec.Components.Compactor.RestartAt = tc.restartAt
				}
			})

			tc.risingwave = risingwave
			tc.component = component
			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

			var deploy *appsv1.Deployment
			switch component {
			case consts.ComponentFrontend:
				deploy = factory.NewFrontendDeployment(tc.group.Name, tc.podTemplate)
			case consts.ComponentCompactor:
				deploy = factory.NewCompactorDeployment(tc.group.Name, tc.podTemplate)
			}
			t.Run(component+"-"+name+" testcase:", func(t *testing.T) {
				ComposeAssertions(deploymentPreds, t).assertTest(deploy, tc)
			})
		}
	}
}

func Test_RisingWaveObjectFactory_CloneSet(t *testing.T) {
	testcases := GetClonesetTestcases()
	clonesetPreds := GetClonesetPredicates()
	for _, component := range []string{consts.ComponentFrontend, consts.ComponentCompactor} {
		for name, tc := range testcases {
			t.Run(component+"-"+name, func(t *testing.T) {
				risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
					r.Spec.EnableOpenKruise = pointer.Bool(true)
					r.Spec.Storages.Meta.Memory = pointer.Bool(true)
					r.Spec.Storages.Object.Memory = pointer.Bool(true)
					if tc.group.Name == "" {
						r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
						switch component {
						case consts.ComponentMeta:
							r.Spec.Global.Replicas.Meta = tc.group.Replicas
						case consts.ComponentFrontend:
							r.Spec.Global.Replicas.Frontend = tc.group.Replicas
						case consts.ComponentCompactor:
							r.Spec.Global.Replicas.Compactor = tc.group.Replicas
						}
					} else {
						switch component {
						case consts.ComponentMeta:
							r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
								tc.group,
							}
						case consts.ComponentFrontend:
							r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
								tc.group,
							}
							r.Spec.Components.Frontend.RestartAt = tc.restartAt
						case consts.ComponentCompactor:
							r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
								tc.group,
							}
							r.Spec.Components.Compactor.RestartAt = tc.restartAt
						}

					}
					switch component {
					case consts.ComponentMeta:
						r.Spec.Components.Meta.RestartAt = tc.restartAt
					case consts.ComponentFrontend:
						r.Spec.Components.Frontend.RestartAt = tc.restartAt
					case consts.ComponentCompactor:
						r.Spec.Components.Compactor.RestartAt = tc.restartAt
					}
				})

				tc.risingwave = risingwave
				tc.component = component
				factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

				var cloneSet *kruiseappsv1alpha1.CloneSet
				switch component {
				case consts.ComponentFrontend:
					cloneSet = factory.NewFrontendCloneSet(tc.group.Name, tc.podTemplate)
				case consts.ComponentCompactor:
					cloneSet = factory.NewCompactorCloneSet(tc.group.Name, tc.podTemplate)
				}
				t.Run(component+"-"+tc.name+" testcase:", func(t *testing.T) {
					ComposeAssertions(clonesetPreds, t).assertTest(cloneSet, tc)
				})

			})
		}
	}
}

func Test_RisingWaveObjectFactory_StatefulSets(t *testing.T) {
	testcases := GetSTSTestcases()
	stsPreds := GetSTSPredicates()
	for _, tc := range testcases {
		risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = pointer.Bool(true)
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.Replicas.Compute = tc.group.Replicas
				r.Spec.Global.RisingWaveComponentGroupTemplate = tc.group.RisingWaveComponentGroupTemplate
			} else {
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					tc.group,
				}
			}
			if tc.restartAt != nil {
				r.Spec.Components.Compute.RestartAt = tc.restartAt
			}
		})

		tc.risingwave = risingwave
		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

		sts := factory.NewComputeStatefulSet(tc.group.Name, tc.podTemplate)

		t.Run("compute-"+tc.name+" testcase:", func(t *testing.T) {
			ComposeAssertions(stsPreds, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_AdvancedStatefulSets(t *testing.T) {
	testcases := GetAdvancedSTSTestcases()
	astsPreds := GetAdvancedSTSPredicates()
	for name, tc := range testcases {
		risingwave := NewTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.Replicas.Compute = tc.group.Replicas
				r.Spec.Global.RisingWaveComponentGroupTemplate = tc.group.RisingWaveComponentGroupTemplate
			} else {
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					tc.group,
				}
			}
			if tc.restartAt != nil {
				r.Spec.Components.Compute.RestartAt = tc.restartAt
			}
		})

		tc.risingwave = risingwave
		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
		asts := factory.NewComputeAdvancedStatefulSet(tc.group.Name, tc.podTemplate)
		t.Run("compute-"+name+" testcase:", func(t *testing.T) {
			ComposeAssertions(astsPreds, t).assertTest(asts, tc)
		})
	}

}
func Test_RisingWaveObjectFactory_ObjectStorages(t *testing.T) {
	// testcases := map[string]struct {
	// 	objectStorage risingwavev1alpha1.RisingWaveObjectStorage
	// 	hummockArg    string
	// 	envs          []corev1.EnvVar
	// }{
	// 	"memory": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			Memory: pointer.Bool(true),
	// 		},
	// 		hummockArg: "hummock+memory",
	// 	},
	// 	"minio": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{
	// 				Secret:   "minio-creds",
	// 				Endpoint: "minio-endpoint:1234",
	// 				Bucket:   "minio-hummock01",
	// 			},
	// 		},
	// 		hummockArg: "hummock+minio://$(MINIO_USERNAME):$(MINIO_PASSWORD)@minio-endpoint:1234/minio-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name: "MINIO_USERNAME",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "minio-creds",
	// 						},
	// 						Key: "username",
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "MINIO_PASSWORD",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "minio-creds",
	// 						},
	// 						Key: "password",
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"s3": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Secret: "s3-creds",
	// 				Bucket: "s3-hummock01",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "AWS_S3_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name: "AWS_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "AWS_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "AWS_REGION",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3Region,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"aliyun-oss": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
	// 				Secret: "s3-creds",
	// 				Bucket: "s3-hummock01",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION).aliyuncs.com",
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_REGION",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3Region,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"aliyun-oss-internal": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
	// 				Secret:           "s3-creds",
	// 				Bucket:           "s3-hummock01",
	// 				InternalEndpoint: true,
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION)-internal.aliyuncs.com",
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_REGION",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3Region,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"aliyun-oss-with-region": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			AliyunOSS: &risingwavev1alpha1.RisingWaveObjectStorageAliyunOSS{
	// 				Secret: "s3-creds",
	// 				Bucket: "s3-hummock01",
	// 				Region: "cn-hangzhou",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://$(S3_COMPATIBLE_BUCKET).oss-$(S3_COMPATIBLE_REGION).aliyuncs.com",
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_REGION",
	// 				Value: "cn-hangzhou",
	// 			},
	// 		},
	// 	},
	// 	"s3-compatible": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Secret:   "s3-creds",
	// 				Bucket:   "s3-hummock01",
	// 				Endpoint: "oss-cn-hangzhou.aliyuncs.com",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://oss-cn-hangzhou.aliyuncs.com",
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_REGION",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3Region,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"s3-compatible-virtual-hosted-style": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Secret:             "s3-creds",
	// 				Bucket:             "s3-hummock01",
	// 				Endpoint:           "https://oss-cn-hangzhou.aliyuncs.com",
	// 				VirtualHostedStyle: true,
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://$(S3_COMPATIBLE_BUCKET).oss-cn-hangzhou.aliyuncs.com",
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "S3_COMPATIBLE_REGION",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3Region,
	// 					},
	// 				},
	// 			},
	// 		},
	// 	},
	// 	"s3-with-region": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Secret: "s3-creds",
	// 				Bucket: "s3-hummock01",
	// 				Region: "ap-southeast-1",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "AWS_S3_BUCKET",
	// 				Value: "s3-hummock01",
	// 			},
	// 			{
	// 				Name: "AWS_ACCESS_KEY_ID",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3AccessKeyID,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name: "AWS_SECRET_ACCESS_KEY",
	// 				ValueFrom: &corev1.EnvVarSource{
	// 					SecretKeyRef: &corev1.SecretKeySelector{
	// 						LocalObjectReference: corev1.LocalObjectReference{
	// 							Name: "s3-creds",
	// 						},
	// 						Key: consts.SecretKeyAWSS3SecretAccessKey,
	// 					},
	// 				},
	// 			},
	// 			{
	// 				Name:  "AWS_REGION",
	// 				Value: "ap-southeast-1",
	// 			},
	// 		},
	// 	},
	// 	"endpoint-with-region-variable": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Bucket:   "s3-hummock01",
	// 				Endpoint: "s3.${REGION}.amazonaws.com",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://s3.$(S3_COMPATIBLE_REGION).amazonaws.com",
	// 			},
	// 		},
	// 	},
	// 	"endpoint-with-bucket-variable": {
	// 		objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
	// 			S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
	// 				Bucket:   "s3-hummock01",
	// 				Endpoint: "${BUCKET}.s3.${REGION}.amazonaws.com",
	// 			},
	// 		},
	// 		hummockArg: "hummock+s3-compatible://s3-hummock01",
	// 		envs: []corev1.EnvVar{
	// 			{
	// 				Name:  "S3_COMPATIBLE_ENDPOINT",
	// 				Value: "https://$(S3_COMPATIBLE_BUCKET).s3.$(S3_COMPATIBLE_REGION).amazonaws.com",
	// 			},
	// 		},
	// 	},
	// }

	// for name, tc := range testcases {
	// 	t.Run(name, func(t *testing.T) {
	// 		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
	// 			r.Spec.Storages.Object = tc.objectStorage
	// 		})

	// 		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

	// 		deploy := factory.NewCompactorDeployment("", nil)

	// 		composeAssertions(
	// 			newObjectAssert(deploy, "hummock-args-match", func(obj *appsv1.Deployment) bool {
	// 				return lo.Contains(obj.Spec.Template.Spec.Containers[0].Args, tc.hummockArg)
	// 			}),
	// 			newObjectAssert(deploy, "env-vars-contains", func(obj *appsv1.Deployment) bool {
	// 				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
	// 			}),
	// 		).Assert(t)

	// 		sts := factory.NewComputeStatefulSet("", nil)

	// 		composeAssertions(
	// 			newObjectAssert(sts, "hummock-args-match", func(obj *appsv1.StatefulSet) bool {
	// 				return lo.Contains(obj.Spec.Template.Spec.Containers[0].Args, tc.hummockArg)
	// 			}),
	// 			newObjectAssert(sts, "env-vars-contains", func(obj *appsv1.StatefulSet) bool {
	// 				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
	// 			}),
	// 		).Assert(t)
	// 	})
	// }
}

func slicesContains(a, b []string) bool {
	if len(a) < len(b) {
		return false
	}
	if len(b) == 0 {
		return true
	}
	for i := 0; i <= len(a)-len(b); i++ {
		if a[i] == b[0] && slices.Equal(a[i:i+len(b)], b) {
			return true
		}
	}
	return false
}

// func Test_RisingWaveObjectFactory_MetaStorages(t *testing.T) {
// 	// testcases := map[string]struct {
// 	// 	metaStorage risingwavev1alpha1.RisingWaveMetaStorage
// 	// 	args        []string
// 	// 	envs        []corev1.EnvVar
// 	// }{
// 	// 	"memory": {
// 	// 		metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
// 	// 			Memory: pointer.Bool(true),
// 	// 		},
// 	// 		args: []string{
// 	// 			"--backend", "mem",
// 	// 		},
// 	// 	},
// 	// 	"etcd-no-auth": {
// 	// 		metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
// 	// 			Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
// 	// 				Endpoint: "etcd:1234",
// 	// 			},
// 	// 		},
// 	// 		args: []string{
// 	// 			"--backend", "etcd", "--etcd-endpoints", "etcd:1234",
// 	// 		},
// 	// 	},
// 	// 	"etcd-auth": {
// 	// 		metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
// 	// 			Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
// 	// 				Endpoint: "etcd:1234",
// 	// 				Secret:   "etcd-credentials",
// 	// 			},
// 	// 		},
// 	// 		args: []string{
// 	// 			"--backend", "etcd", "--etcd-endpoints", "etcd:1234", "--etcd-auth",
// 	// 		},
// 	// 		envs: []corev1.EnvVar{
// 	// 			{
// 	// 				Name: "RW_ETCD_USERNAME",
// 	// 				ValueFrom: &corev1.EnvVarSource{
// 	// 					SecretKeyRef: &corev1.SecretKeySelector{
// 	// 						LocalObjectReference: corev1.LocalObjectReference{
// 	// 							Name: "etcd-credentials",
// 	// 						},
// 	// 						Key: consts.SecretKeyEtcdUsername,
// 	// 					},
// 	// 				},
// 	// 			},
// 	// 			{
// 	// 				Name: "RW_ETCD_PASSWORD",
// 	// 				ValueFrom: &corev1.EnvVarSource{
// 	// 					SecretKeyRef: &corev1.SecretKeySelector{
// 	// 						LocalObjectReference: corev1.LocalObjectReference{
// 	// 							Name: "etcd-credentials",
// 	// 						},
// 	// 						Key: consts.SecretKeyEtcdPassword,
// 	// 					},
// 	// 				},
// 	// 			},
// 	// 		},
// 	// 	},
// 	// }

// 	// for name, tc := range testcases {
// 	// 	t.Run(name, func(t *testing.T) {
// 	// 		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
// 	// 			r.Spec.Storages.Meta = tc.metaStorage
// 	// 		})

// 	// 		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
// 	// 		deploy := factory.NewMetaStatefulSet("", nil)

// 	// 		composeAssertions(
// 	// 			newObjectAssert(deploy, "args-match", func(obj *appsv1.StatefulSet) bool {
// 	// 				return slicesContains(obj.Spec.Template.Spec.Containers[0].Args, tc.args)
// 	// 			}),
// 	// 			newObjectAssert(deploy, "env-vars-contains", func(obj *appsv1.StatefulSet) bool {
// 	// 				return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
// 	// 			}),
// 	// 		).Assert(t)
// 	// 	})
// 	// }
// }

func Test_RisingWaveObjectFactory_ServiceMonitor(t *testing.T) {
	// risingwave := testutils.FakeRisingWave()

	// factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
	// serviceMonitor := factory.NewServiceMonitor()

	// composeAssertions(
	// 	newObjectAssert(serviceMonitor, "owned", func(obj *prometheusv1.ServiceMonitor) bool {
	// 		return controlledBy(risingwave, obj)
	// 	}),
	// 	newObjectAssert(serviceMonitor, "has-labels", func(obj *prometheusv1.ServiceMonitor) bool {
	// 		return hasLabels(obj, map[string]string{
	// 			consts.LabelRisingWaveName:       risingwave.Name,
	// 			consts.LabelRisingWaveGeneration: strconv.FormatInt(risingwave.Generation, 10),
	// 		}, true)
	// 	}),
	// 	newObjectAssert(serviceMonitor, "select-risingwave-services", func(obj *prometheusv1.ServiceMonitor) bool {
	// 		return mapEquals(obj.Spec.Selector.MatchLabels, map[string]string{
	// 			consts.LabelRisingWaveName: risingwave.Name,
	// 		})
	// 	}),
	// 	newObjectAssert(serviceMonitor, "target-labels", func(obj *prometheusv1.ServiceMonitor) bool {
	// 		return listContains(obj.Spec.TargetLabels, []string{
	// 			consts.LabelRisingWaveName,
	// 			consts.LabelRisingWaveComponent,
	// 			consts.LabelRisingWaveGroup,
	// 		})
	// 	}),
	// 	newObjectAssert(serviceMonitor, "scrape-port-metrics", func(obj *prometheusv1.ServiceMonitor) bool {
	// 		return len(obj.Spec.Endpoints) > 0 && obj.Spec.Endpoints[0].Port == consts.PortMetrics
	// 	}),
	// ).Assert(t)
}

func Test_RisingWaveObjectFactory_InheritLabels(t *testing.T) {
	testcases := map[string]struct {
		labels             map[string]string
		inheritPrefixValue string
		inheritedLabels    map[string]string
	}{
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

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(&risingwavev1alpha1.RisingWave{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tc.labels,
					Annotations: map[string]string{
						consts.AnnotationInheritLabelPrefix: tc.inheritPrefixValue,
					},
				},
			}, nil)

			assert.Equal(t, tc.inheritedLabels, factory.getInheritedLabels(), "inherited labels not match")
		})
	}
}

func containsSlice[T comparable](a, b []T) bool {
	for i := 0; i <= len(a)-len(b); i++ {
		match := true
		for j, element := range b {
			if a[i+j] != element {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}

	return false
}

func TestRisingWaveObjectFactory_ComputeArgs(t *testing.T) {
	testcases := map[string]struct {
		cpuLimit int64
		memLimit int64
		argsList [][]string
	}{
		"empty-limits": {},
		"cpu-limit-4": {
			cpuLimit: 4,
			argsList: [][]string{
				{"--parallelism", "4"},
			},
		},
		"mem-limit-4g": {
			memLimit: 4 << 30,
			argsList: [][]string{
				{"--total-memory-bytes", strconv.Itoa((4 << 30) - (512 << 20))},
			},
		},
		"mem-limit-1g": {
			memLimit: 1 << 30,
			argsList: [][]string{
				{"--total-memory-bytes", strconv.Itoa((1 << 30) - (512 << 20))},
			},
		},
		"mem-limit-768m": {
			memLimit: 768 << 20,
			argsList: [][]string{
				{"--total-memory-bytes", strconv.Itoa(512 << 20)},
			},
		},
		"mem-limit-512m": {
			memLimit: 512 << 20,
			argsList: [][]string{
				{"--total-memory-bytes", strconv.Itoa(512 << 20)},
			},
		},
		"mem-limit-256m": {
			memLimit: 256 << 20,
			argsList: [][]string{
				{"--total-memory-bytes", strconv.Itoa(256 << 20)},
			},
		},
		"cpu-and-mem": {
			cpuLimit: 4,
			memLimit: 1 << 30,
			argsList: [][]string{
				{"--parallelism", "4"},
				{"--total-memory-bytes", strconv.Itoa((1 << 30) - (512 << 20))},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(&risingwavev1alpha1.RisingWave{
				Spec: risingwavev1alpha1.RisingWaveSpec{
					Storages: risingwavev1alpha1.RisingWaveStoragesSpec{
						Object: risingwavev1alpha1.RisingWaveObjectStorage{
							Memory: pointer.Bool(true),
						},
					},
				},
			}, nil)
			args := factory.argsForCompute(tc.cpuLimit, tc.memLimit)

			for _, expectArgs := range tc.argsList {
				if !containsSlice(args, expectArgs) {
					t.Fatalf("Args not expected, expects \"%s\", but is \"%s\"",
						strings.Join(expectArgs, " "), strings.Join(args, " "))
				}
			}
		})
	}
}
