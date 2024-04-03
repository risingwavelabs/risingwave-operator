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
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveObjectFactory_Services(t *testing.T) {
	predicates := servicesPredicates()

	for name, tc := range servicesTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableStandaloneMode = ptr.To(tc.enableStandaloneMode)
			r.Spec.FrontendServiceType = tc.globalServiceType
		})

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")

		var svc *corev1.Service
		switch tc.component {
		case consts.ComponentMeta:
			svc = factory.NewMetaService()
		case consts.ComponentFrontend:
			svc = factory.NewFrontendService()
		case consts.ComponentCompute:
			svc = factory.NewComputeService()
		case consts.ComponentCompactor:
			svc = factory.NewCompactorService()
		case consts.ComponentStandalone:
			svc = factory.NewStandaloneService()
		default:
			t.Fatal("bad test")
		}

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServicesMeta(t *testing.T) {
	predicates := serviceMetadataPredicates()

	for name, tc := range serviceMetadataTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.AdditionalFrontendServiceMetadata = tc.globalServiceMeta
		})

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")

		var svc *corev1.Service
		switch tc.component {
		case consts.ComponentMeta:
			svc = factory.NewMetaService()
		case consts.ComponentFrontend:
			svc = factory.NewFrontendService()
		case consts.ComponentCompute:
			svc = factory.NewComputeService()
		case consts.ComponentCompactor:
			svc = factory.NewCompactorService()
		case consts.ComponentStandalone:
			svc = factory.NewStandaloneService()
		default:
			t.Fatal("bad test")
		}

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ConfigMaps(t *testing.T) {
	predicates := configMapPredicates()

	for name, tc := range configMapTestCases() {
		tc.risingwave = newTestRisingwave()
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		cm := factory.NewConfigConfigMap(tc.configVal)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(cm, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Frontend_Deployments(t *testing.T) {
	predicates := frontendDeploymentPredicates()

	for name, tc := range deploymentTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Frontend.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentFrontend

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		deploy := factory.NewFrontendDeployment(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_Deployments(t *testing.T) {
	predicates := compactorDeploymentPredicates()

	for name, tc := range deploymentTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Compactor.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentCompactor

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		deploy := factory.NewCompactorDeployment(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Frontend_CloneSet(t *testing.T) {
	predicates := frontendCloneSetPredicates()

	for name, tc := range cloneSetTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = ptr.To(true)
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Frontend.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentFrontend

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		cloneSet := factory.NewFrontendCloneSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_CloneSet(t *testing.T) {
	predicates := compactorCloneSetPredicates()

	for name, tc := range cloneSetTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = ptr.To(true)
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Compactor.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentCompactor

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		cloneSet := factory.NewCompactorCloneSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_StatefulSets(t *testing.T) {
	predicates := metaStatefulSetPredicates()

	for name, tc := range metaStatefulSetTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Meta.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentMeta

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		sts := factory.NewMetaStatefulSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compute_StatefulSets(t *testing.T) {
	predicates := computeStatefulSetPredicates()

	for name, tc := range computeStatefulSetTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = ptr.To(true)
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Compute.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		sts := factory.NewComputeStatefulSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_AdvancedStatefulSets(t *testing.T) {
	predicates := metaAdvancedSTSPredicates()

	for name, tc := range metaAdvancedSTSTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Meta.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		tc.component = consts.ComponentMeta

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		sts := factory.NewMetaAdvancedStatefulSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compute_AdvancedStatefulSets(t *testing.T) {
	predicates := computeAdvancedSTSPredicates()

	for name, tc := range computeAdvancedSTSTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore.Memory = ptr.To(true)
			r.Spec.StateStore.Memory = ptr.To(true)
			tc.group.RestartAt = tc.restartAt
			r.Spec.Components.Compute.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
				tc.group,
			}
		})

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
		asts := factory.NewComputeAdvancedStatefulSet(tc.group.Name)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(asts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_StateStores(t *testing.T) {
	predicates := stateStoreStatefulsetPredicates()

	for name, tc := range stateStoreTestCases() {
		t.Run(name, func(t *testing.T) {
			tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{Memory: ptr.To(true)}
				r.Spec.StateStore = tc.stateStore
				r.Spec.Components = risingwavev1alpha1.RisingWaveComponentsSpec{
					Meta: risingwavev1alpha1.RisingWaveComponent{
						NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
							{
								Name:     "",
								Replicas: 1,
							},
						},
					},
				}
			})

			factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme, "")
			sts := factory.NewMetaStatefulSet("")

			t.Run(name, func(t *testing.T) {
				composeAssertions(predicates, t).assertTest(sts, tc)
			})
		})
	}
}

func Test_RisingWaveObjectFactory_MetaStores(t *testing.T) {
	predicates := metaStorePredicates()

	for name, tc := range metaStoreTestCases() {
		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.MetaStore = tc.metaStore
			r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{Memory: ptr.To(true)}
			r.Spec.Components = risingwavev1alpha1.RisingWaveComponentsSpec{
				Meta: risingwavev1alpha1.RisingWaveComponent{
					NodeGroups: []risingwavev1alpha1.RisingWaveNodeGroup{
						{
							Name:     "",
							Replicas: 1,
						},
					},
				},
			}
		})

		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme, "")
		sts := factory.NewMetaStatefulSet("")

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServiceMonitor(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	predicates := serviceMonitorPredicates()

	factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme, "")
	serviceMonitor := factory.NewServiceMonitor()

	composeAssertions(predicates, t).assertTest(serviceMonitor, baseTestCase{risingwave: risingwave})
}

func Test_RisingWaveObjectFactory_InheritLabels(t *testing.T) {
	for name, tc := range inheritedLabelsTestCases() {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(&risingwavev1alpha1.RisingWave{
				ObjectMeta: metav1.ObjectMeta{
					Labels: tc.labels,
					Annotations: map[string]string{
						consts.AnnotationInheritLabelPrefix: tc.inheritPrefixValue,
					},
				},
			}, nil, "")

			assert.Equal(t, tc.inheritedLabels, factory.getInheritedLabels(), "inherited labels not match")
		})
	}
}

func TestRisingWaveObjectFactory_ComputeArgs(t *testing.T) {
	for name, tc := range computeEnvsTestCases() {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(&risingwavev1alpha1.RisingWave{
				Spec: risingwavev1alpha1.RisingWaveSpec{
					StateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
						Memory: ptr.To(true),
					},
				},
			}, nil, "")
			envs := factory.envsForComputeArgs(tc.cpuLimit, tc.memLimit)

			for _, expectEnv := range tc.envList {
				if !listContainsByKey(envs, []corev1.EnvVar{expectEnv}, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar]) {
					t.Errorf("Env not found or value not expected, expects \"%s\"", testutils.JSONMustPrettyPrint(expectEnv))
				}
			}
		})
	}
}

func TestRisingWaveObjectFactory_TlsSupport(t *testing.T) {
	predicates := tlsPredicates()

	for name, tc := range tlsTestcases() {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore.Memory = ptr.To(true)
				r.Spec.StateStore.Memory = ptr.To(true)
				r.Spec.EnableStandaloneMode = ptr.To(tc.standalone)
				r.Spec.TLS = tc.tls
				r.Spec.Components.Frontend.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Name: "",
					},
				}
			}), testutils.Scheme, "")

			template := lo.If(tc.standalone, factory.NewStandaloneStatefulSet().Spec.Template).
				Else(factory.NewFrontendDeployment("").Spec.Template)
			composeAssertions(predicates, t).assertTest(&template, tc)
		})
	}
}

func TestRisingWaveObjectFactory_DataDirectory(t *testing.T) {
	testcases := map[string]struct {
		stateStore   risingwavev1alpha1.RisingWaveStateStoreBackend
		internalRoot string
		expect       string
	}{
		"gcs-without-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				GCS:           &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{},
			},
			internalRoot: "",
			expect:       "hummock",
		},
		"gcs-with-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
					Root: "root",
				},
			},
			internalRoot: "",
			expect:       "root/hummock",
		},
		"gcs-with-internal-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
					Root: "root",
				},
			},
			internalRoot: "root",
			expect:       "root/hummock",
		},
		"gcs-with-internal-root-only": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				GCS:           &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{},
			},
			internalRoot: "root",
			expect:       "root/hummock",
		},
		"azblob-without-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				AzureBlob:     &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{},
			},
			internalRoot: "",
			expect:       "hummock",
		},
		"azblob-with-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{
					Root: "root",
				},
			},
			internalRoot: "",
			expect:       "root/hummock",
		},
		"azblob-with-internal-root": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{
					Root: "root",
				},
			},
			internalRoot: "root",
			expect:       "root/hummock",
		},
		"azblob-with-internal-root-only": {
			stateStore: risingwavev1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
				AzureBlob:     &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{},
			},
			internalRoot: "root",
			expect:       "root/hummock",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			factory := NewRisingWaveObjectFactory(newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = tc.stateStore
				r.Status.Internal.StateStoreRootPath = tc.internalRoot
			}), testutils.Scheme, "")

			assert.Equal(t, tc.expect, factory.getDataDirectory())
		})
	}
}
