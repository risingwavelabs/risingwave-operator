// Copyright 2023 RisingWave Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package factory

import (
	"strconv"
	"testing"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

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

func mapContainsWith[K comparable, V any](a, b map[K]V, equals func(a, b V) bool) bool {
	if len(a) < len(b) {
		return false
	}

	for k, v := range b {
		va, ok := a[k]
		if !ok || !equals(va, v) {
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
	}

	for k, v := range a {
		vb, ok := b[k]
		if !ok || v != vb {
			return false
		}
	}

	return true
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

func serviceLabels(risingwave *risingwavev1alpha1.RisingWave, component string, sync bool) map[string]string {
	labels := componentLabels(risingwave, component, sync)
	switch component {
	case consts.ComponentFrontend:
		labels = mergeMap(labels, risingwave.Spec.AdditionalFrontendServiceMetadata.Labels)
	case consts.ComponentMeta:
		labels = mergeMap(labels, risingwave.Spec.AdditionalMetaServiceMetadata.Labels)
	}
	return labels
}

func componentLabels(risingwave *risingwavev1alpha1.RisingWave, component string, sync bool) map[string]string {
	labels := map[string]string{
		consts.LabelRisingWaveName:            risingwave.Name,
		consts.LabelRisingWaveComponent:       component,
		consts.LabelRisingWaveOperatorVersion: "",
	}
	if sync {
		labels[consts.LabelRisingWaveGeneration] = strconv.FormatInt(risingwave.Generation, 10)
	} else {
		labels[consts.LabelRisingWaveGeneration] = consts.NoSync
	}

	return labels
}

func componentGroupLabels(risingwave *risingwavev1alpha1.RisingWave, component string, group *string, sync bool) map[string]string {
	labels := map[string]string{
		consts.LabelRisingWaveName:            risingwave.Name,
		consts.LabelRisingWaveComponent:       component,
		consts.LabelRisingWaveOperatorVersion: "",
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

func componentAnnotations(risingwave *risingwavev1alpha1.RisingWave, component string) map[string]string {
	annotations := map[string]string{}
	if component == consts.ComponentFrontend {
		annotations = mergeMap(annotations, risingwave.Spec.AdditionalFrontendServiceMetadata.Annotations)
	}

	return annotations
}

func componentGroupAnnotations(risingwave *risingwavev1alpha1.RisingWave, group *string) map[string]string {
	annotations := map[string]string{}

	return annotations
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
	aKeys, bKeys := make(map[K]T), make(map[K]T)
	for i, x := range a {
		aKeys[key(&x)] = a[i]
	}

	for i, x := range b {
		bKeys[key(&x)] = b[i]
	}

	return mapContainsWith(aKeys, bKeys, equals)
}

//nolint:unused
func containsStringSlice(a, b []string) bool {
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

//nolint:unused
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

type composedAssertion[T kubeObject, testcase testCaseType] struct {
	t          *testing.T
	predicates []predicate[T, testcase]
}

func (a *composedAssertion[T, K]) assertTest(obj T, testcase K) {
	for _, pred := range a.predicates {
		if !pred.Fn(obj, testcase) {
			a.t.Errorf("Assertion %s failed", pred.Name)
		}
	}
}

func composeAssertions[T kubeObject, K testCaseType](predicates []predicate[T, K], t *testing.T) *composedAssertion[T, K] {
	return &composedAssertion[T, K]{
		predicates: predicates,
		t:          t,
	}
}
