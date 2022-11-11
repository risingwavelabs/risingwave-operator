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

package factory

import (
	"math"
	"strconv"
	"testing"
	"time"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/pointer"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

type assertion interface {
	Assert(t *testing.T)
}

type assertFunc func(t *testing.T)

func (f assertFunc) Assert(t *testing.T) {
	f(t)
}

type composedAssertion struct {
	asserts []assertion
}

func (a *composedAssertion) Assert(t *testing.T) {
	for _, assert := range a.asserts {
		assert.Assert(t)
	}
}

func composeAssertions(asserts ...assertion) assertion {
	return &composedAssertion{asserts: asserts}
}

func newObjectAssert[T client.Object](obj T, desc string, pred func(obj T) bool) assertion {
	return assertFunc(func(t *testing.T) {
		if !pred(obj) {
			t.Fatalf("Assert %s failed", desc)
		}
	})
}

func mapContains[K, V comparable](a, b map[K]V) bool {
	if len(a) < len(b) {
		return false
	}

	for k, v := range b {
		va, ok := a[k]
		if !ok || va != v {
			return false
		}
	}

	return true
}

func mapEquals[K, V comparable](a, b map[K]V) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	} else if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return false
	} else {
		for k, v := range a {
			vb, ok := b[k]
			if !ok || v != vb {
				return false
			}
		}
		return true
	}
}

func hasLabels[T client.Object](obj T, labels map[string]string, exact bool) bool {
	for k, v := range labels {
		v1, ok := obj.GetLabels()[k]
		if !ok || v != v1 {
			return false
		}
	}
	if exact && len(obj.GetLabels()) != len(labels) {
		return false
	}
	return true
}

func hasAnnotations[T client.Object](obj T, annotations map[string]string, exact bool) bool {
	for k, v := range annotations {
		v1, ok := obj.GetAnnotations()[k]
		if !ok || v != v1 {
			return false
		}
	}
	if exact && len(obj.GetAnnotations()) != len(annotations) {
		return false
	}
	return true
}

func isServiceType(svc *corev1.Service, t corev1.ServiceType) bool {
	return svc.Spec.Type == t
}

func hasTCPServicePorts(svc *corev1.Service, ports map[string]int32) bool {
	svcPorts := make(map[string]corev1.ServicePort)
	for _, port := range svc.Spec.Ports {
		svcPorts[port.Name] = port
	}

	for name, port := range ports {
		svcPort, ok := svcPorts[name]
		if !ok || (svcPort.Protocol != corev1.ProtocolTCP && svcPort.Protocol != "") || svcPort.Port != port {
			return false
		}
	}

	return true
}

func hasServiceSelector(svc *corev1.Service, selector map[string]string) bool {
	return equality.Semantic.DeepEqual(svc.Spec.Selector, selector)
}

func componentLabels(risingwave *risingwavev1alpha1.RisingWave, component string, group *string, sync bool) map[string]string {
	labels := map[string]string{
		consts.LabelRisingWaveName:      risingwave.Name,
		consts.LabelRisingWaveComponent: component,
	}
	if sync {
		labels[consts.LabelRisingWaveGeneration] = strconv.FormatInt(risingwave.Generation, 10)
	} else {
		labels[consts.LabelRisingWaveGeneration] = consts.NoSync
	}
	if group != nil {
		labels[consts.LabelRisingWaveGroup] = *group
	}

	return labels
}

func podSelector(risingwave *risingwavev1alpha1.RisingWave, component string, group *string) map[string]string {
	labels := map[string]string{
		consts.LabelRisingWaveName:      risingwave.Name,
		consts.LabelRisingWaveComponent: component,
	}
	if group != nil {
		labels[consts.LabelRisingWaveGroup] = *group
	}
	return labels
}

func controlledBy(owner, ownee client.Object) bool {
	controllerRef, ok := lo.Find(ownee.GetOwnerReferences(), func(ref metav1.OwnerReference) bool {
		return ref.Controller != nil && *ref.Controller
	})
	if !ok {
		return false
	}
	return controllerRef.UID == owner.GetUID()
}

func newTestRisingwave(patches ...func(r *risingwavev1alpha1.RisingWave)) *risingwavev1alpha1.RisingWave {
	r := &risingwavev1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "test",
			Namespace:  rand.String(10),
			Generation: int64(rand.Int()),
			UID:        uuid.NewUUID(),
		},
	}
	for _, patch := range patches {
		patch(r)
	}
	return r
}

func setupMetaPorts(r *risingwavev1alpha1.RisingWave, ports map[string]int32) {
	r.Spec.Components.Meta.Ports = risingwavev1alpha1.RisingWaveComponentMetaPorts{
		RisingWaveComponentCommonPorts: risingwavev1alpha1.RisingWaveComponentCommonPorts{
			ServicePort: ports[consts.PortService],
			MetricsPort: ports[consts.PortMetrics],
		},
		DashboardPort: ports[consts.PortDashboard],
	}
}

func setupFrontendPorts(r *risingwavev1alpha1.RisingWave, ports map[string]int32) {
	r.Spec.Components.Frontend.Ports = risingwavev1alpha1.RisingWaveComponentCommonPorts{
		ServicePort: ports[consts.PortService],
		MetricsPort: ports[consts.PortMetrics],
	}
}

func setupComputePorts(r *risingwavev1alpha1.RisingWave, ports map[string]int32) {
	r.Spec.Components.Compute.Ports = risingwavev1alpha1.RisingWaveComponentCommonPorts{
		ServicePort: ports[consts.PortService],
		MetricsPort: ports[consts.PortMetrics],
	}
}

func setupCompactorPorts(r *risingwavev1alpha1.RisingWave, ports map[string]int32) {
	r.Spec.Components.Compactor.Ports = risingwavev1alpha1.RisingWaveComponentCommonPorts{
		ServicePort: ports[consts.PortService],
		MetricsPort: ports[consts.PortMetrics],
	}
}

func Test_RisingWaveObjectFactory_Services(t *testing.T) {
	testcases := map[string]struct {
		component         string
		ports             map[string]int32
		globalServiceType corev1.ServiceType
		expectServiceType corev1.ServiceType
	}{
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
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
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
				}
			})

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
			default:
				t.Fatal("bad test")
			}

			composeAssertions(
				newObjectAssert(svc, "controlled-by-risingwave", func(obj *corev1.Service) bool {
					return controlledBy(risingwave, obj)
				}),
				newObjectAssert(svc, "namespace-equals", func(obj *corev1.Service) bool {
					return obj.Namespace == risingwave.Namespace
				}),
				newObjectAssert(svc, "ports-equal", func(obj *corev1.Service) bool {
					return hasTCPServicePorts(obj, tc.ports)
				}),
				newObjectAssert(svc, "service-type-match", func(obj *corev1.Service) bool {
					return isServiceType(obj, tc.expectServiceType)
				}),
				newObjectAssert(svc, "service-labels-match", func(obj *corev1.Service) bool {
					return hasLabels(obj, componentLabels(risingwave, tc.component, nil, true), true)
				}),
				newObjectAssert(svc, "selector-equals", func(obj *corev1.Service) bool {
					return hasServiceSelector(obj, podSelector(risingwave, tc.component, nil))
				}),
			).Assert(t)
		})
	}
}

func Test_RisingWaveObjectFactory_ConfigMaps(t *testing.T) {
	testcases := map[string]struct {
		configVal string
	}{
		"empty-val": {
			configVal: "",
		},
		"non-empty-val": {
			configVal: "a",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := newTestRisingwave()

			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

			cm := factory.NewConfigConfigMap(tc.configVal)

			composeAssertions(
				newObjectAssert(cm, "controlled-by-risingwave", func(obj *corev1.ConfigMap) bool {
					return controlledBy(risingwave, obj)
				}),
				newObjectAssert(cm, "namespace-equals", func(obj *corev1.ConfigMap) bool {
					return obj.Namespace == risingwave.Namespace
				}),
				newObjectAssert(cm, "configmap-labels-match", func(obj *corev1.ConfigMap) bool {
					return hasLabels(obj, componentLabels(risingwave, consts.ComponentConfig, nil, false), true)
				}),
				newObjectAssert(cm, "configmap-data-match", func(obj *corev1.ConfigMap) bool {
					return mapEquals(obj.Data, map[string]string{
						risingWaveConfigMapKey: lo.If(tc.configVal == "", "").Else(tc.configVal),
					})
				}),
			).Assert(t)
		})
	}
}

func deepEqual[T any](x, y T) bool {
	return equality.Semantic.DeepEqual(x, y)
}

func listContains[T comparable](a, b []T) bool {
	if len(b) > len(a) {
		return false
	}

	m := make(map[T]int)
	for _, x := range a {
		m[x]++
	}

	for _, x := range b {
		c := m[x]
		if c == 0 {
			return false
		}
		m[x]--
	}
	return true
}

func listContainsByKey[T any, K comparable](a, b []T, key func(*T) K, equals func(x, y T) bool) bool {
	bKeys := make(map[K]*T)
	for i, x := range b {
		bKeys[key(&x)] = &b[i]
	}
	for _, x := range a {
		if y, ok := bKeys[key(&x)]; ok {
			if !equals(x, *y) {
				return false
			}
		}
	}
	return true
}

func matchesPodTemplate(podSpec *corev1.PodTemplateSpec, podTemplate *risingwavev1alpha1.RisingWavePodTemplateSpec) bool {
	if podTemplate == nil {
		return true
	}

	if !(mapContains(podSpec.Labels, podTemplate.Labels) &&
		mapContains(podSpec.Annotations, podTemplate.Annotations)) {
		return false
	}

	pSpec, tSpec := podSpec.Spec, podTemplate.Spec
	pSpec.Containers = pSpec.Containers[1:]
	tSpec.Containers = tSpec.Containers[1:]

	// Check volumes first.
	if !listContainsByKey(pSpec.Volumes, tSpec.Volumes, func(x *corev1.Volume) string { return x.Name }, deepEqual[corev1.Volume]) {
		return false
	}
	pSpec.Volumes, tSpec.Volumes = nil, nil

	// Set default enable service links to false.
	if tSpec.EnableServiceLinks == nil {
		tSpec.EnableServiceLinks = pointer.Bool(false)
	}

	if !equality.Semantic.DeepEqual(pSpec, tSpec) {
		return false
	}

	pContainer, tContainer := podSpec.Spec.Containers[0], podTemplate.Spec.Containers[0]

	// Only check the
	//   * SecurityContext
	//   * Env
	//   * EnvFrom
	//   * VolumeDevices
	return deepEqual(pContainer.SecurityContext, tContainer.SecurityContext) &&
		listContainsByKey(pContainer.VolumeMounts, tContainer.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount]) &&
		listContainsByKey(pContainer.Env, tContainer.Env, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar]) &&
		listContainsByKey(pContainer.EnvFrom, tContainer.EnvFrom, func(t *corev1.EnvFromSource) string { return t.Prefix }, deepEqual[corev1.EnvFromSource])

}

func newPodTemplate(patches ...func(t *risingwavev1alpha1.RisingWavePodTemplateSpec)) *risingwavev1alpha1.RisingWavePodTemplateSpec {
	t := &risingwavev1alpha1.RisingWavePodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "",
				},
			},
		},
	}

	for _, patch := range patches {
		patch(t)
	}

	return t
}

func Test_RisingWaveObjectFactory_Deployments(t *testing.T) {
	testcases := map[string]struct {
		podTemplate           map[string]risingwavev1alpha1.RisingWavePodTemplate
		group                 risingwavev1alpha1.RisingWaveComponentGroup
		restartAt             *metav1.Time
		expectUpgradeStrategy *appsv1.DeploymentStrategy
	}{
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
			expectUpgradeStrategy: &appsv1.DeploymentStrategy{
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
			expectUpgradeStrategy: &appsv1.DeploymentStrategy{
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

	for _, component := range []string{consts.ComponentMeta, consts.ComponentFrontend, consts.ComponentCompactor} {
		for name, tc := range testcases {
			t.Run(component+"-"+name, func(t *testing.T) {
				risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

				group := &tc.group.Name

				factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

				var deploy *appsv1.Deployment
				switch component {
				case consts.ComponentMeta:
					deploy = factory.NewMetaDeployment(tc.group.Name, tc.podTemplate)
				case consts.ComponentFrontend:
					deploy = factory.NewFrontendDeployment(tc.group.Name, tc.podTemplate)
				case consts.ComponentCompactor:
					deploy = factory.NewCompactorDeployment(tc.group.Name, tc.podTemplate)
				}

				composeAssertions(
					newObjectAssert(deploy, "namespace-equals", func(obj *appsv1.Deployment) bool {
						return obj.Namespace == risingwave.Namespace
					}),
					newObjectAssert(deploy, "labels-equal", func(obj *appsv1.Deployment) bool {
						return hasLabels(obj, componentLabels(risingwave, component, group, true), true)
					}),
					newObjectAssert(deploy, "replicas-equal", func(obj *appsv1.Deployment) bool {
						return *obj.Spec.Replicas == tc.group.Replicas
					}),
					newObjectAssert(deploy, "pod-template-labels-match", func(obj *appsv1.Deployment) bool {
						return mapContains(obj.Spec.Template.Labels, podSelector(risingwave, component, group))
					}),
					newObjectAssert(deploy, "pod-template-annotations-match", func(obj *appsv1.Deployment) bool {
						if tc.restartAt != nil {
							return mapContains(obj.Spec.Template.Annotations, map[string]string{
								consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
							})
						} else {
							_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
							return !ok
						}
					}),
					newObjectAssert(deploy, "pod-template-works", func(obj *appsv1.Deployment) bool {
						if tc.group.PodTemplate != nil {
							temp := tc.podTemplate[*tc.group.PodTemplate].Template
							return matchesPodTemplate(&obj.Spec.Template, &temp)
						} else {
							return true
						}
					}),
					newObjectAssert(deploy, "image-match", func(obj *appsv1.Deployment) bool {
						return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
					}),
					newObjectAssert(deploy, "image-pull-policy-match", func(obj *appsv1.Deployment) bool {
						return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
					}),
					newObjectAssert(deploy, "image-pull-secrets-match", func(obj *appsv1.Deployment) bool {
						for _, s := range tc.group.ImagePullSecrets {
							if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
								return false
							}
						}
						return true
					}),
					newObjectAssert(deploy, "resources-match", func(obj *appsv1.Deployment) bool {
						return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
					}),
					newObjectAssert(deploy, "node-selector-match", func(obj *appsv1.Deployment) bool {
						return mapContains(obj.Spec.Template.Spec.NodeSelector, tc.group.NodeSelector)
					}),
					newObjectAssert(deploy, "upgrade-strategy-match", func(obj *appsv1.Deployment) bool {
						if tc.expectUpgradeStrategy == nil {
							return equality.Semantic.DeepEqual(obj.Spec.Strategy, appsv1.DeploymentStrategy{})
						} else {
							return equality.Semantic.DeepEqual(obj.Spec.Strategy, *tc.expectUpgradeStrategy)
						}
					}),
					newObjectAssert(deploy, "first-container-must-have-probes", func(obj *appsv1.Deployment) bool {
						container := &obj.Spec.Template.Spec.Containers[0]
						return container.LivenessProbe != nil && container.ReadinessProbe != nil
					}),
				).Assert(t)
			})
		}
	}
}

func Test_RisingWaveObjectFactory_CloneSet(t *testing.T) {
	testcases := map[string]struct {
		podTemplate             map[string]risingwavev1alpha1.RisingWavePodTemplate
		group                   risingwavev1alpha1.RisingWaveComponentGroup
		restartAt               *metav1.Time
		expectedUpgradeStrategy *kruiseappsv1alpha1.CloneSetUpdateStrategy
	}{
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
		"upgrade-strategy-InplaceOnly": {
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
		"upgrade-strategy-InplaceOnly-max-unavailable-50%": {
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
		"upgrade-strategy-InplaceOnly-max-surge-50%": {
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
		"upgrade-strategy-InplaceOnly-partition-50%": {
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
		"upgrade-strategy-InplaceOnly-Grace-Period-20seconds": {
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
		"upgrade-strategy-InplaceIfPossible": {
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
		"upgrade-strategy-InplaceIfPossible-max-unavailable-50%": {
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
		"upgrade-strategy-InplaceIfPossible-max-surge-50%": {
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
		"upgrade-strategy-InplaceIfPossible-partition-50%": {
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
		"upgrade-strategy-InplaceIfPossible-Grace-Period-20seconds": {
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

	for _, component := range []string{consts.ComponentMeta} {
		for name, tc := range testcases {
			t.Run(component+"-"+name, func(t *testing.T) {
				risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

				group := &tc.group.Name
				factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

				var cloneSet *kruiseappsv1alpha1.CloneSet
				switch component {
				case consts.ComponentMeta:
					cloneSet = factory.NewMetaCloneSet(tc.group.Name, tc.podTemplate)
				case consts.ComponentFrontend:
					cloneSet = factory.NewFrontEndCloneSet(tc.group.Name, tc.podTemplate)
				case consts.ComponentCompactor:
					cloneSet = factory.NewCompactorCloneSet(tc.group.Name, tc.podTemplate)
				}

				composeAssertions(
					newObjectAssert(cloneSet, "namespace-equals", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return obj.Namespace == risingwave.Namespace
					}),
					newObjectAssert(cloneSet, "labels-equal", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return hasLabels(obj, componentLabels(risingwave, component, group, true), true)
					}),
					newObjectAssert(cloneSet, "replicas-equal", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return *obj.Spec.Replicas == tc.group.Replicas
					}),
					newObjectAssert(cloneSet, "pod-template-labels-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return mapContains(obj.Spec.Template.Labels, podSelector(risingwave, component, group))
					}),
					newObjectAssert(cloneSet, "pod-template-annotations-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						if tc.restartAt != nil {
							return mapContains(obj.Spec.Template.Annotations, map[string]string{
								consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
							})
						} else {
							_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
							return !ok
						}
					}),
					newObjectAssert(cloneSet, "pod-template-works", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						if tc.group.PodTemplate != nil {
							temp := tc.podTemplate[*tc.group.PodTemplate].Template
							return matchesPodTemplate(&obj.Spec.Template, &temp)
						} else {
							return true
						}
					}),
					newObjectAssert(cloneSet, "image-pull-secrets-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						for _, s := range tc.group.ImagePullSecrets {
							if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
								return false
							}
						}
						return true
					}),
					newObjectAssert(cloneSet, "resources-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
					}),
					newObjectAssert(cloneSet, "node-selector-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						return mapContains(obj.Spec.Template.Spec.NodeSelector, tc.group.NodeSelector)
					}),
					newObjectAssert(cloneSet, "upgrade-strategy-match", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						if tc.expectedUpgradeStrategy == nil {
							return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1alpha1.CloneSetUpdateStrategy{})
						} else {
							return equality.Semantic.DeepEqual(&obj.Spec.UpdateStrategy, tc.expectedUpgradeStrategy)
						}
					}),
					newObjectAssert(cloneSet, "first-container-must-have-probes", func(obj *kruiseappsv1alpha1.CloneSet) bool {
						container := &obj.Spec.Template.Spec.Containers[0]
						return container.LivenessProbe != nil && container.ReadinessProbe != nil
					}),
				).Assert(t)
			})
		}
	}
}

func Test_RisingWaveObjectFactory_StatefulSets(t *testing.T) {
	testcases := map[string]struct {
		podTemplate           map[string]risingwavev1alpha1.RisingWavePodTemplate
		group                 risingwavev1alpha1.RisingWaveComputeGroup
		restartAt             *metav1.Time
		expectUpgradeStrategy *appsv1.StatefulSetUpdateStrategy
	}{
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
			expectUpgradeStrategy: nil,
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
			expectUpgradeStrategy: &appsv1.StatefulSetUpdateStrategy{
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

	for name, tc := range testcases {
		t.Run("compute-"+name, func(t *testing.T) {
			risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

			group := &tc.group.Name

			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

			sts := factory.NewComputeStatefulSet(tc.group.Name, tc.podTemplate)

			composeAssertions(
				newObjectAssert(sts, "namespace-equals", func(obj *appsv1.StatefulSet) bool {
					return obj.Namespace == risingwave.Namespace
				}),
				newObjectAssert(sts, "labels-equal", func(obj *appsv1.StatefulSet) bool {
					return hasLabels(obj, componentLabels(risingwave, consts.ComponentCompute, group, true), true)
				}),
				newObjectAssert(sts, "replicas-equal", func(obj *appsv1.StatefulSet) bool {
					return *obj.Spec.Replicas == tc.group.Replicas
				}),
				newObjectAssert(sts, "pod-template-labels-match", func(obj *appsv1.StatefulSet) bool {
					return mapContains(obj.Spec.Template.Labels, podSelector(risingwave, consts.ComponentCompute, group))
				}),
				newObjectAssert(sts, "pod-template-annotations-match", func(obj *appsv1.StatefulSet) bool {
					if tc.restartAt != nil {
						return mapContains(obj.Spec.Template.Annotations, map[string]string{
							consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
						})
					} else {
						_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
						return !ok
					}
				}),
				newObjectAssert(sts, "pod-template-works", func(obj *appsv1.StatefulSet) bool {
					if tc.group.PodTemplate != nil {
						temp := tc.podTemplate[*tc.group.PodTemplate].Template
						return matchesPodTemplate(&obj.Spec.Template, &temp)
					} else {
						return true
					}
				}),
				newObjectAssert(sts, "image-match", func(obj *appsv1.StatefulSet) bool {
					return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
				}),
				newObjectAssert(sts, "image-pull-policy-match", func(obj *appsv1.StatefulSet) bool {
					return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
				}),
				newObjectAssert(sts, "image-pull-secrets-match", func(obj *appsv1.StatefulSet) bool {
					for _, s := range tc.group.ImagePullSecrets {
						if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
							return false
						}
					}
					return true
				}),
				newObjectAssert(sts, "resources-match", func(obj *appsv1.StatefulSet) bool {
					return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
				}),
				newObjectAssert(sts, "node-selector-match", func(obj *appsv1.StatefulSet) bool {
					return mapContains(obj.Spec.Template.Spec.NodeSelector, tc.group.NodeSelector)
				}),
				newObjectAssert(sts, "upgrade-strategy-match", func(obj *appsv1.StatefulSet) bool {
					if tc.expectUpgradeStrategy == nil {
						return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, appsv1.StatefulSetUpdateStrategy{})
					} else {
						return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectUpgradeStrategy)
					}
				}),
				newObjectAssert(sts, "check-volume-mounts", func(obj *appsv1.StatefulSet) bool {
					return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
				}),
				newObjectAssert(sts, "first-container-must-have-probes", func(obj *appsv1.StatefulSet) bool {
					container := &obj.Spec.Template.Spec.Containers[0]
					return container.LivenessProbe != nil && container.ReadinessProbe != nil
				}),
			).Assert(t)
		})
	}
}

func Test_RisingWaveObjectFactory_AdvancedStatefulSets(t *testing.T) {
	testcases := map[string]struct {
		podTemplate             map[string]risingwavev1alpha1.RisingWavePodTemplate
		group                   risingwavev1alpha1.RisingWaveComputeGroup
		restartAt               *metav1.Time
		expectedUpgradeStrategy *kruiseappsv1beta1.StatefulSetUpdateStrategy
	}{
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

	for name, tc := range testcases {
		t.Run("compute-"+name, func(t *testing.T) {
			risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
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

			group := &tc.group.Name

			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

			asts := factory.NewComputeAdvancedStatefulSet(tc.group.Name, tc.podTemplate)

			composeAssertions(
				newObjectAssert(asts, "namespace-equals", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return obj.Namespace == risingwave.Namespace
				}),
				newObjectAssert(asts, "labels-equal", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return hasLabels(obj, componentLabels(risingwave, consts.ComponentCompute, group, true), true)
				}),
				newObjectAssert(asts, "replicas-equal", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return *obj.Spec.Replicas == tc.group.Replicas
				}),
				newObjectAssert(asts, "pod-template-labels-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return mapContains(obj.Spec.Template.Labels, podSelector(risingwave, consts.ComponentCompute, group))
				}),
				newObjectAssert(asts, "pod-template-annotations-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					if tc.restartAt != nil {
						return mapContains(obj.Spec.Template.Annotations, map[string]string{
							consts.AnnotationRestartAt: tc.restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z"),
						})
					} else {
						_, ok := obj.Spec.Template.Annotations[consts.AnnotationRestartAt]
						return !ok
					}
				}),
				newObjectAssert(asts, "pod-template-works", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					if tc.group.PodTemplate != nil {
						temp := tc.podTemplate[*tc.group.PodTemplate].Template
						return matchesPodTemplate(&obj.Spec.Template, &temp)
					} else {
						return true
					}
				}),
				newObjectAssert(asts, "image-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return obj.Spec.Template.Spec.Containers[0].Image == tc.group.Image
				}),
				newObjectAssert(asts, "image-pull-policy-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return obj.Spec.Template.Spec.Containers[0].ImagePullPolicy == tc.group.ImagePullPolicy
				}),
				newObjectAssert(asts, "image-pull-secrets-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					for _, s := range tc.group.ImagePullSecrets {
						if !lo.Contains(obj.Spec.Template.Spec.ImagePullSecrets, corev1.LocalObjectReference{Name: s}) {
							return false
						}
					}
					return true
				}),
				newObjectAssert(asts, "resources-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return equality.Semantic.DeepEqual(obj.Spec.Template.Spec.Containers[0].Resources, tc.group.Resources)
				}),
				newObjectAssert(asts, "node-selector-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return mapContains(obj.Spec.Template.Spec.NodeSelector, tc.group.NodeSelector)
				}),
				newObjectAssert(asts, "upgrade-strategy-match", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					if tc.expectedUpgradeStrategy == nil {
						return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, kruiseappsv1beta1.StatefulSetUpdateStrategy{})
					} else {
						return equality.Semantic.DeepEqual(obj.Spec.UpdateStrategy, *tc.expectedUpgradeStrategy)
					}
				}),
				newObjectAssert(asts, "check-volume-mounts", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					return listContainsByKey(obj.Spec.Template.Spec.Containers[0].VolumeMounts, tc.group.VolumeMounts, func(t *corev1.VolumeMount) string { return t.MountPath }, deepEqual[corev1.VolumeMount])
				}),
				newObjectAssert(asts, "first-container-must-have-probes", func(obj *kruiseappsv1beta1.StatefulSet) bool {
					container := &obj.Spec.Template.Spec.Containers[0]
					return container.LivenessProbe != nil && container.ReadinessProbe != nil
				}),
			).Assert(t)
		})
	}
}
func Test_RisingWaveObjectFactory_ObjectStorages(t *testing.T) {
	testcases := map[string]struct {
		objectStorage risingwavev1alpha1.RisingWaveObjectStorage
		hummockArg    string
		envs          []corev1.EnvVar
	}{
		"memory": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				Memory: pointer.Bool(true),
			},
			hummockArg: "hummock+memory",
		},
		"minio": {
			objectStorage: risingwavev1alpha1.RisingWaveObjectStorage{
				MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{
					Secret:   "minio-creds",
					Endpoint: "minio-endpoint:1234",
					Bucket:   "minio-hummock01",
				},
			},
			hummockArg: "hummock+minio://$(MINIO_USERNAME):$(MINIO_PASSWORD)@minio-endpoint:1234/minio-hummock01",
			envs: []corev1.EnvVar{
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
			hummockArg: "hummock+s3://s3-hummock01",
			envs: []corev1.EnvVar{
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
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = tc.objectStorage
			})

			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)

			deploy := factory.NewCompactorDeployment("", nil)

			composeAssertions(
				newObjectAssert(deploy, "hummock-args-match", func(obj *appsv1.Deployment) bool {
					return lo.Contains(obj.Spec.Template.Spec.Containers[0].Args, tc.hummockArg)
				}),
				newObjectAssert(deploy, "env-vars-contains", func(obj *appsv1.Deployment) bool {
					return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
				}),
			).Assert(t)

			sts := factory.NewComputeStatefulSet("", nil)

			composeAssertions(
				newObjectAssert(sts, "hummock-args-match", func(obj *appsv1.StatefulSet) bool {
					return lo.Contains(obj.Spec.Template.Spec.Containers[0].Args, tc.hummockArg)
				}),
				newObjectAssert(sts, "env-vars-contains", func(obj *appsv1.StatefulSet) bool {
					return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
				}),
			).Assert(t)
		})
	}
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

func Test_RisingWaveObjectFactory_MetaStorages(t *testing.T) {
	testcases := map[string]struct {
		metaStorage risingwavev1alpha1.RisingWaveMetaStorage
		args        []string
		envs        []corev1.EnvVar
	}{
		"memory": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Memory: pointer.Bool(true),
			},
			args: []string{
				"--backend", "mem",
			},
		},
		"etcd-no-auth": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
					Endpoint: "etcd:1234",
				},
			},
			args: []string{
				"--backend", "etcd", "--etcd-endpoints", "etcd:1234",
			},
		},
		"etcd-auth": {
			metaStorage: risingwavev1alpha1.RisingWaveMetaStorage{
				Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
					Endpoint: "etcd:1234",
					Secret:   "etcd-credentials",
				},
			},
			args: []string{
				"--backend", "etcd", "--etcd-endpoints", "etcd:1234", "--etcd-auth",
			},
			envs: []corev1.EnvVar{
				{
					Name: "ETCD_USERNAME",
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
					Name: "ETCD_PASSWORD",
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

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := newTestRisingwave(func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = tc.metaStorage
			})

			factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
			deploy := factory.NewMetaDeployment("", nil)

			composeAssertions(
				newObjectAssert(deploy, "args-match", func(obj *appsv1.Deployment) bool {
					return slicesContains(obj.Spec.Template.Spec.Containers[0].Args, tc.args)
				}),
				newObjectAssert(deploy, "env-vars-contains", func(obj *appsv1.Deployment) bool {
					return listContainsByKey(obj.Spec.Template.Spec.Containers[0].Env, tc.envs, func(t *corev1.EnvVar) string { return t.Name }, deepEqual[corev1.EnvVar])
				}),
			).Assert(t)
		})
	}
}

func Test_RisingWaveObjectFactory_ServiceMonitor(t *testing.T) {
	risingwave := testutils.FakeRisingWave()

	factory := NewRisingWaveObjectFactory(risingwave, testutils.Scheme)
	serviceMonitor := factory.NewServiceMonitor()

	composeAssertions(
		newObjectAssert(serviceMonitor, "owned", func(obj *prometheusv1.ServiceMonitor) bool {
			return controlledBy(risingwave, obj)
		}),
		newObjectAssert(serviceMonitor, "has-labels", func(obj *prometheusv1.ServiceMonitor) bool {
			return hasLabels(obj, map[string]string{
				consts.LabelRisingWaveName:       risingwave.Name,
				consts.LabelRisingWaveGeneration: strconv.FormatInt(risingwave.Generation, 10),
			}, true)
		}),
		newObjectAssert(serviceMonitor, "select-risingwave-services", func(obj *prometheusv1.ServiceMonitor) bool {
			return mapEquals(obj.Spec.Selector.MatchLabels, map[string]string{
				consts.LabelRisingWaveName: risingwave.Name,
			})
		}),
		newObjectAssert(serviceMonitor, "target-labels", func(obj *prometheusv1.ServiceMonitor) bool {
			return listContains(obj.Spec.TargetLabels, []string{
				consts.LabelRisingWaveName,
				consts.LabelRisingWaveComponent,
				consts.LabelRisingWaveGroup,
			})
		}),
		newObjectAssert(serviceMonitor, "scrape-port-metrics", func(obj *prometheusv1.ServiceMonitor) bool {
			return len(obj.Spec.Endpoints) > 0 && obj.Spec.Endpoints[0].Port == consts.PortMetrics
		}),
	).Assert(t)
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
