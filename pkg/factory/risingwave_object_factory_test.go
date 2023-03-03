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

func Test_RisingWaveObjectFactory_Services(t *testing.T) {
	predicates := servicesPredicates()

	for name, tc := range servicesTestCases() {
		tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServicesMeta(t *testing.T) {
	predicates := serviceMetadataPredicates()

	for name, tc := range serviceMetadataTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(svc, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ConfigMaps(t *testing.T) {
	predicates := configMapPredicates()

	for name, tc := range configMapTestCases() {
		tc.risingwave = newTestRisingwave()
		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_Deployments(t *testing.T) {
	predicates := compactorDeploymentPredicates()

	for name, tc := range deploymentTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(deploy, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Frontend_CloneSet(t *testing.T) {
	predicates := frontendCloneSetPredicates()

	for name, tc := range cloneSetTestCases() {
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

		tc.component = consts.ComponentFrontend

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		cloneSet := factory.NewFrontendCloneSet(tc.group.Name, tc.podTemplate)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compactor_CloneSet(t *testing.T) {
	predicates := compactorCloneSetPredicates()

	for name, tc := range cloneSetTestCases() {
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

		tc.component = consts.ComponentCompactor

		factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
		cloneSet := factory.NewCompactorCloneSet(tc.group.Name, tc.podTemplate)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(cloneSet, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_StatefulSets(t *testing.T) {
	predicates := metaStatefulSetPredicates()

	for name, tc := range metaStatefulSetTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compute_StatefulSets(t *testing.T) {
	predicates := computeStatefulSetPredicates()

	for name, tc := range computeStatefulSetTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Meta_AdvancedStatefulSets(t *testing.T) {
	predicates := metaAdvancedSTSPredicates()

	for name, tc := range metaAdvancedSTSTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_Compute_AdvancedStatefulSets(t *testing.T) {
	predicates := computeAdvancedSTSPredicates()

	for name, tc := range computeAdvancedSTSTestCases() {
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

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(asts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ObjectStorages(t *testing.T) {
	predicates := objectStorageStatefulsetPredicates()

	for name, tc := range objectStorageTestCases() {
		t.Run(name, func(t *testing.T) {
			tc.risingwave = newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages = risingwavev1alpha1.RisingWaveStoragesSpec{
					Meta:   risingwavev1alpha1.RisingWaveMetaStorage{Memory: pointer.Bool(true)},
					Object: tc.objectStorage,
				}
			})

			factory := NewRisingWaveObjectFactory(tc.risingwave, testutils.Scheme)
			sts := factory.NewMetaStatefulSet("", nil)

			t.Run(name, func(t *testing.T) {
				composeAssertions(predicates, t).assertTest(sts, tc)
			})
		})
	}
}

func Test_RisingWaveObjectFactory_MetaStorages(t *testing.T) {
	predicates := metaStoragePredicates()

	for name, tc := range metaStorageTestCases() {
		risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
			r.Spec.Storages = risingwavev1alpha1.RisingWaveStoragesSpec{
				Meta:   tc.metaStorage,
				Object: risingwavev1alpha1.RisingWaveObjectStorage{Memory: pointer.Bool(true)},
			}
		})

		factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
		sts := factory.NewMetaStatefulSet("", nil)

		t.Run(name, func(t *testing.T) {
			composeAssertions(predicates, t).assertTest(sts, tc)
		})
	}
}

func Test_RisingWaveObjectFactory_ServiceMonitor(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	predicates := serviceMonitorPredicates()

	factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
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
			}, nil)

			assert.Equal(t, tc.inheritedLabels, factory.getInheritedLabels(), "inherited labels not match")
		})
	}
}

func TestRisingWaveObjectFactory_ComputeArgs(t *testing.T) {
	for name, tc := range computeArgsTestCases() {
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
