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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

// This file contains the test cases for the object factory. The tests compromises of composing assertions based on a set of
// predicates that can be defined in predicates.go. Predicates are functions that take in an obj and a testcase both constrained
// by the kubeObject contraint and the testcaseType constraint defined in test_common.go.
func Test_RisingWaveObjectFactory_Services(t *testing.T) {
	testcases := getServicesTestcases()
	servicesPreds := getServicesPredicate()
	for name, tc := range testcases {
		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Global.ServiceType = tc.globalServiceType

			switch tc.component {
			case consts.ComponentMeta:
				setupMetaPorts(r, tc.ports)
			case consts.ComponentFrontend:
				setupFrontendPorts(r, tc.ports)
			case consts.ComponentCompute:
				setupComputePorts(r, tc.ports)
			case consts.ComponentCompactor:
				setupCompactorPorts(r, tc.ports)
			case consts.ComponentConnector:
				setupConnectorPorts(r, tc.ports)
			}
		})

		tc.risingwave = risingwave
		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

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
		case consts.ComponentConnector:
			svc = factory.NewConnectorService()
		default:
			t.Fatal("bad test")
		}
		t.Run(name+" testcase:", func(t *testing.T) {
			composeAssertions(servicesPreds, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServicesMeta(t *testing.T) {
	testcases := getServicesMetaTestCases()
	servicesPreds := getServicesMetaPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Global.ServiceMeta = tc.globalServiceMeta
		})

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)

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
		case consts.ComponentConnector:
			svc = factory.NewConnectorService()
		default:
			t.Fatal("bad test")
		}
		t.Run(name+" testcase:", func(t *testing.T) {
			composeAssertions(servicesPreds, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ConfigMaps(t *testing.T) {
	testcases := getConfigmapTestCases()
	configmapPreds := getConfigmapPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave()
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		cm := factory.NewConfigConfigMap(tc.configVal)

		t.Run(name, func(t *testing.T) {
			composeAssertions(configmapPreds, t).assertTest(cm, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Frontend_Deployments(t *testing.T) {
	testcases := getDeploymentTestcases()
	deploymentPreds := getFrontendDeploymentPredicates()
	component := consts.ComponentFrontend
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Frontend = tc.group.Replicas
			} else {
				r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Frontend.RestartAt = tc.restartAt
			}
			r.Spec.Components.Frontend.RestartAt = tc.restartAt
		})

		tc.component = consts.ComponentFrontend
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		deploy := factory.NewFrontendDeployment(tc.group.Name, tc.podTemplate)

		t.Run(component+"-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(deploymentPreds, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_Deployments(t *testing.T) {
	testcases := getDeploymentTestcases()
	deploymentPreds := getCompactorDeploymentPredicates()
	component := consts.ComponentCompactor
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Compactor = tc.group.Replicas
			} else {
				r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Compactor.RestartAt = tc.restartAt
			}
			r.Spec.Components.Compactor.RestartAt = tc.restartAt
		})

		tc.component = consts.ComponentCompactor
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		deploy := factory.NewCompactorDeployment(tc.group.Name, tc.podTemplate)

		t.Run(component+"-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(deploymentPreds, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Frontend_CloneSet(t *testing.T) {
	testcases := getClonesetTestcases()
	clonesetPreds := getFrontendClonesetPredicates()
	component := consts.ComponentFrontend
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = pointer.Bool(true)
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Frontend = tc.group.Replicas
			} else {
				r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Frontend.RestartAt = tc.restartAt
			}
			r.Spec.Components.Frontend.RestartAt = tc.restartAt
		})

		tc.component = component
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		cloneSet := factory.NewFrontendCloneSet(tc.group.Name, tc.podTemplate)

		t.Run(component+"-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(clonesetPreds, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_CloneSet(t *testing.T) {
	testcases := getClonesetTestcases()
	clonesetPreds := getCompactorClonesetPredicates()
	component := consts.ComponentCompactor
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.EnableOpenKruise = pointer.Bool(true)
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Compactor = tc.group.Replicas
			} else {
				r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Compactor.RestartAt = tc.restartAt
			}
			r.Spec.Components.Compactor.RestartAt = tc.restartAt
		})

		tc.component = component
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		cloneSet := factory.NewCompactorCloneSet(tc.group.Name, tc.podTemplate)

		t.Run(component+"-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(clonesetPreds, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_StatefulSets(t *testing.T) {
	testcases := getMetaSTSTestcases()
	stsPreds := getMetaSTSPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Meta = tc.group.Replicas
			} else {
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Meta.RestartAt = tc.restartAt
			}
			if tc.restartAt != nil {
				r.Spec.Components.Meta.RestartAt = tc.restartAt
			}
		})

		tc.component = consts.ComponentMeta
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		sts := factory.NewMetaStatefulSet(tc.group.Name, tc.podTemplate)

		t.Run("Meta-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(stsPreds, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compute_StatefulSets(t *testing.T) {
	testcases := getSTSTestcases()
	stsPreds := getComputeSTSPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)

		sts := factory.NewComputeStatefulSet(tc.group.Name, tc.podTemplate)

		t.Run("compute-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(stsPreds, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_AdvancedStatefulSets(t *testing.T) {
	testcases := getMetaAdvancedSTSTestcases()
	stsPreds := getMetaAdvancedSTSPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta.Memory = pointer.Bool(true)
			r.Spec.Storages.Object.Memory = pointer.Bool(true)
			if tc.group.Name == "" {
				r.Spec.Global.RisingWaveComponentGroupTemplate = *tc.group.RisingWaveComponentGroupTemplate
				r.Spec.Global.Replicas.Meta = tc.group.Replicas
			} else {
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					tc.group,
				}
				r.Spec.Components.Meta.RestartAt = tc.restartAt
			}
			if tc.restartAt != nil {
				r.Spec.Components.Meta.RestartAt = tc.restartAt
			}
		})

		tc.component = consts.ComponentMeta
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		sts := factory.NewMetaAdvancedStatefulSet(tc.group.Name, tc.podTemplate)

		t.Run("Meta-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(stsPreds, t).assertTest(sts, tc)
		})
	}
}
func Test_RisingWaveObjectFactory_AdvancedStatefulSets(t *testing.T) {
	testcases := getAdvancedSTSTestcases()
	astsPreds := getComputeAdvancedSTSPredicates()
	for name, tc := range testcases {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		asts := factory.NewComputeAdvancedStatefulSet(tc.group.Name, tc.podTemplate)

		t.Run("compute-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(astsPreds, t).assertTest(asts, tc)
		})
	}
}
func Test_RisingWaveObjectFactory_ObjectStorages(t *testing.T) {
	testcases := getObjectStoragesTestcase()
	deployObjectStoragePreds := getObjectStoratesDeploymentPredicates()
	stsObjectStoragePreds := getObjectStoratesStatefulsetPredicates()
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = tc.objectStorage
			})

			factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)

			deploy := factory.NewCompactorDeployment("", nil)
			t.Run(name+" testcase:", func(t *testing.T) {
				composeAssertions(deployObjectStoragePreds, t).assertTest(deploy, tc)
			})

			sts := factory.NewComputeStatefulSet("", nil)

			t.Run(name+" testcase:", func(t *testing.T) {
				composeAssertions(stsObjectStoragePreds, t).assertTest(sts, tc)
			})
		})
	}
}
func Test_RisingWaveObjectFactory_MetaStorages(t *testing.T) {
	testcases := getMetaStoragesTestCases()
	metaStoragePreds := getMetaStoragePredicates()
	for name, tc := range testcases {
		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages.Meta = tc.metaStorage
		})

		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
		sts := factory.NewMetaStatefulSet("", nil)

		t.Run("compute-"+name+" testcase:", func(t *testing.T) {
			composeAssertions(metaStoragePreds, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServiceMonitor(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	serviceMonitorPreds := getServiceMonitorPredicates()
	factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
	serviceMonitor := factory.NewServiceMonitor()
	composeAssertions(serviceMonitorPreds, t).assertTest(serviceMonitor, baseTestCase{risingwave: risingwave})
}

func Test_RisingWaveObjectFactory_InheritLabels(t *testing.T) {
	testcases := getInheritedLabelsTestCaes()
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
	testcases := getComputeArgsTestCases()
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
					t.Errorf("Args not expected, expects \"%s\", but is \"%s\"",
						strings.Join(expectArgs, " "), strings.Join(args, " "))
				}
			}
		})
	}
}
