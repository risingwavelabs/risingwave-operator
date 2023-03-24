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

package manager

import (
	"context"
	"sort"
	"strconv"
	"testing"

	"github.com/go-logr/logr"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/strings/slices"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/risingwavelabs/risingwave-operator/pkg/event"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func newRisingWaveControllerManagerImplForTest(risingwave *risingwavev1alpha1.RisingWave, objects ...client.Object) *risingWaveControllerManagerImpl {
	fakeClient := fake.NewClientBuilder().
		WithObjects(append(objects, risingwave.DeepCopy())...).
		WithScheme(testutils.Scheme).
		Build()
	risingwaveManager := object.NewRisingWaveManager(fakeClient, risingwave.DeepCopy(), false)
	return newRisingWaveControllerManagerImpl(fakeClient, risingwaveManager, event.NewMessageStore(), "")
}

func newRisingWaveControllerManagerImplOpenKruiseAvailableForTest(risingwave *risingwavev1alpha1.RisingWave, objects ...client.Object) *risingWaveControllerManagerImpl {
	fakeClient := fake.NewClientBuilder().
		WithObjects(append(objects, risingwave.DeepCopy())...).
		WithScheme(testutils.Scheme).
		Build()
	risingwaveManager := object.NewRisingWaveManager(fakeClient, risingwave.DeepCopy(), true)
	return newRisingWaveControllerManagerImpl(fakeClient, risingwaveManager, event.NewMessageStore(), "")
}

func Test_IsObjectNil(t *testing.T) {
	testcases := map[string]struct {
		obj      client.Object
		expected bool
	}{
		"nil-interface": {
			obj:      nil,
			expected: true,
		},
		"nil-ptr-non-nil-interface": {
			obj:      (*corev1.Service)(nil),
			expected: true,
		},
		"non-nil-ptr": {
			obj:      &corev1.Service{},
			expected: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if tc.expected != isObjectNil(tc.obj) {
				t.Fail()
			}
		})
	}
}

func TestRisingWaveControllerManagerImpl_IsObjectSynced(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	testcases := map[string]struct {
		obj      client.Object
		expected bool
	}{
		"nil-not-synced": {
			obj:      nil,
			expected: false,
		},
		"nil-ptr-not-synced": {
			obj:      (*corev1.Service)(nil),
			expected: false,
		},
		"no-generation-label-not-synced": {
			obj:      &corev1.Service{},
			expected: false,
		},
		"generation-less-than-not-synced": {
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation-1, 10),
					},
				},
			},
			expected: false,
		},
		"generation-equal-synced": {
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation, 10),
					},
				},
			},
			expected: true,
		},
		"generation-greater-than-synced": {
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation+1, 10),
					},
				},
			},
			expected: true,
		},
		"generation-label-nosync-synced": {
			obj: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelRisingWaveGeneration: consts.NoSync,
					},
				},
			},
			expected: true,
		},
	}

	managerImpl := newRisingWaveControllerManagerImplForTest(testutils.FakeRisingWave())
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if tc.expected != managerImpl.isObjectSynced(tc.obj) {
				t.Fail()
			}
		})
	}
}

func Test_EnsureTheSameObject(t *testing.T) {
	testcases := map[string]struct {
		objA client.Object
		objB client.Object
		same bool
	}{
		"nils-not-the-same": {
			objA: nil,
			objB: nil,
			same: false,
		},
		"right-nil-not-the-same": {
			objA: &corev1.Service{},
			objB: nil,
			same: false,
		},
		"left-nil-not-the-same": {
			objA: nil,
			objB: &corev1.Service{},
			same: false,
		},
		"left-nil-ptr-the-same": {
			objA: (*corev1.Service)(nil),
			objB: &corev1.Service{},
			same: true,
		},
		"right-nil-ptr-not-the-same": {
			objA: &corev1.Service{},
			objB: (*corev1.Service)(nil),
			same: false,
		},
		"different-name-not-the-same": {
			objA: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "A"}},
			objB: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "B"}},
			same: false,
		},
		"different-namespace-not-the-same": {
			objA: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "A"}},
			objB: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "B"}},
			same: false,
		},
		"different-type-not-the-same": {
			objA: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "A"}},
			objB: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "A"}},
			same: false,
		},
		"different-spec-the-same": {
			objA: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Namespace: "A"},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeNodePort,
				},
			},
			objB: &corev1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "A"}},
			same: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r == nil) != tc.same {
					if tc.same {
						t.Fatal(r)
					} else {
						t.Fatal("expect not same")
					}
				}
			}()

			ensureTheSameObject(tc.objA, tc.objB)
		})
	}
}

func testSyncObject[T any, TP ptrAsObject[T]](t *testing.T, testcases map[string]struct {
	key     types.NamespacedName
	obj     TP
	factory func() TP
}) {
	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			initObjs := lo.Filter([]client.Object{tc.obj}, func(obj client.Object, _ int) bool {
				return !isObjectNil(obj)
			})
			managerImpl := newRisingWaveControllerManagerImplForTest(testutils.FakeRisingWave(), initObjs...)
			newObj := tc.factory()
			if err := managerImpl.syncObject(context.Background(), tc.obj, func() (client.Object, error) {
				return newObj, nil
			}, logr.Discard()); err != nil {
				t.Fatal(err)
			}

			var curObj T
			if err := managerImpl.client.Get(context.Background(), tc.key, TP(&curObj)); err != nil {
				t.Fatal(err)
			}

			if managerImpl.isObjectSynced(tc.obj) {
				if !equality.Semantic.DeepDerivative(tc.obj, TP(&curObj)) {
					t.Fatal("synced object shouldn't be updated")
				}
			} else {
				if !equality.Semantic.DeepDerivative(newObj, TP(&curObj)) {
					t.Fatal("un-synced object should be updated")
				}
			}
		})
	}
}

func newObjectFromKey[T any, TP ptrAsObject[T]](key types.NamespacedName, labels map[string]string) *T {
	var t T
	TP(&t).SetName(key.Name)
	TP(&t).SetNamespace(key.Namespace)
	TP(&t).SetLabels(labels)
	return &t
}

func TestRisingWaveControllerManagerImpl_SyncObject(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: "", Name: "t"}
	factory := func() *corev1.Service {
		return newObjectFromKey[corev1.Service](key, map[string]string{
			consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation, 10),
		})
	}
	testSyncObject(t, map[string]struct {
		key     types.NamespacedName
		obj     *corev1.Service
		factory func() *corev1.Service
	}{
		"current-nil": {
			key:     key,
			obj:     nil,
			factory: factory,
		},
		"current-nosync": {
			key: key,
			obj: newObjectFromKey[corev1.Service](key, map[string]string{
				consts.LabelRisingWaveGeneration: consts.NoSync,
			}),
			factory: factory,
		},
		"current-synced": {
			key: key,
			obj: newObjectFromKey[corev1.Service](key, map[string]string{
				consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation+1, 10),
			}),
			factory: factory,
		},
		"current-not-synced": {
			key: key,
			obj: newObjectFromKey[corev1.Service](key, map[string]string{
				consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation-1, 10),
			}),
			factory: factory,
		},
	})
}

func testRisingWaveControllerManagerImplSyncSingleObject[T any, TP ptrAsObject[T]](t *testing.T, key types.NamespacedName, sync func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *T) (ctrl.Result, error), hooks ...func(t *testing.T, obj *T)) {
	fakeRisingwave := testutils.FakeRisingWave()

	testcases := map[string]struct {
		origin TP
	}{
		"no-origin": {
			origin: nil,
		},
		"origin-not-synced": {
			origin: newObjectFromKey[T, TP](key, nil),
		},
		"origin-synced": {
			origin: newObjectFromKey[T, TP](key, map[string]string{
				consts.LabelRisingWaveGeneration: strconv.FormatInt(fakeRisingwave.Generation-1, 10),
			}),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			initObjs := lo.Filter([]client.Object{tc.origin}, func(obj client.Object, _ int) bool {
				return !isObjectNil(obj)
			})
			managerImpl := newRisingWaveControllerManagerImplForTest(testutils.FakeRisingWave(), initObjs...)
			r, err := sync(managerImpl, context.Background(), logr.Discard(), (*T)(tc.origin))
			if ctrlkit.NeedsRequeue(r, err) {
				t.Fatal("sync failed", r, err)
			}

			var current T
			if err := managerImpl.client.Get(context.Background(), key, TP(&current)); err != nil {
				t.Fatal(err)
			}

			if !managerImpl.isObjectSynced(TP(&current)) {
				t.Fatal("object not synced after sync")
			}

			for _, hook := range hooks {
				hook(t, &current)
			}
		})
	}
}

func TestRisingWaveControllerManagerImpl_SyncServiceMonitor(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()
	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: "risingwave-" + fakeRisingwave.Name}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *monitoringv1.ServiceMonitor) (ctrl.Result, error) {
			return managerImpl.SyncServiceMonitor(ctx, logger, obj)
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncMetaService(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: fakeRisingwave.Name + "-meta"}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *corev1.Service) (ctrl.Result, error) {
			return managerImpl.SyncMetaService(ctx, logger, obj)
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncFrontendService(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: fakeRisingwave.Name + "-frontend"}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *corev1.Service) (ctrl.Result, error) {
			return managerImpl.SyncFrontendService(ctx, logger, obj)
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncComputeService(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: fakeRisingwave.Name + "-compute"}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *corev1.Service) (ctrl.Result, error) {
			return managerImpl.SyncComputeService(ctx, logger, obj)
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncCompactorService(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: fakeRisingwave.Name + "-compactor"}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *corev1.Service) (ctrl.Result, error) {
			return managerImpl.SyncCompactorService(ctx, logger, obj)
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncConfigConfigMap(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()

	key := types.NamespacedName{Namespace: fakeRisingwave.Namespace, Name: fakeRisingwave.Name + "-config"}
	testRisingWaveControllerManagerImplSyncSingleObject(t, key,
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj *corev1.ConfigMap) (ctrl.Result, error) {
			return managerImpl.SyncConfigConfigMap(ctx, logger, obj)
		},
		func(t *testing.T, obj *corev1.ConfigMap) {
			generation := obj.Labels[consts.LabelRisingWaveGeneration]
			if generation != consts.NoSync {
				t.Fatal("must be nosync")
			}
		},
	)
}

type ptrAsObjectList[T any] interface {
	*T
	client.ObjectList
}

func newGroupObjectFromGroup[T any, TP ptrAsObject[T]](namespace, base, group string, labels map[string]string) T {
	var t T
	if group == "" {
		TP(&t).SetName(base)
	} else {
		TP(&t).SetName(base + "-" + group)
	}
	TP(&t).SetNamespace(namespace)
	labels[consts.LabelRisingWaveGroup] = group
	TP(&t).SetLabels(labels)
	return t
}

func testRisingWaveControllerManagerImplSyncObjectGroups[T any, TL any, TP ptrAsObject[T], TLP ptrAsObjectList[TL]](
	t *testing.T,
	risingwave *risingwavev1alpha1.RisingWave,
	openKruiseAvailable bool,
	initObjs []client.Object,
	groups []string,
	labelSelector map[string]string,
	syncGroups func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []T) (ctrl.Result, error),
	getItems func(*TL) []T,
	hooks ...func(t *testing.T, obj *T)) {
	testcases := map[string]struct {
		origin []T
	}{
		"no-objects": {
			origin: nil,
		},
		"some-groups": {
			origin: []T{
				newGroupObjectFromGroup[T, TP](risingwave.Namespace, risingwave.Name, "", labelSelector),
				newGroupObjectFromGroup[T, TP](risingwave.Namespace, risingwave.Name, testutils.GetGroupName(0), labelSelector),
				newGroupObjectFromGroup[T, TP](risingwave.Namespace, risingwave.Name, "group3", labelSelector),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			initObjs := lo.Filter(append(initObjs, lo.Map[T](tc.origin, func(t T, _ int) client.Object { return TP(&t) })...),
				func(obj client.Object, _ int) bool {
					return !isObjectNil(obj)
				},
			)
			var managerImpl *risingWaveControllerManagerImpl
			if !openKruiseAvailable {
				managerImpl = newRisingWaveControllerManagerImplForTest(risingwave, initObjs...)
			} else {
				managerImpl = newRisingWaveControllerManagerImplOpenKruiseAvailableForTest(risingwave, initObjs...)
			}
			r, err := syncGroups(managerImpl, context.Background(), logr.Discard(), tc.origin)
			if ctrlkit.NeedsRequeue(r, err) {
				t.Fatal("sync failed", r, err)
			}

			var currentList TL
			if err := managerImpl.client.List(context.Background(), TLP(&currentList)); err != nil {
				t.Fatal(err)
			}

			currentGroups := make([]string, 0)
			for _, obj := range getItems(&currentList) {
				if !managerImpl.isObjectSynced(TP(&obj)) {
					t.Fatal("object not synced after sync")
				}
				for _, hook := range hooks {
					hook(t, &obj)
				}
				currentGroups = append(currentGroups, TP(&obj).GetLabels()[consts.LabelRisingWaveGroup])
			}

			sort.Strings(groups)
			sort.Strings(currentGroups)
			if !slices.Equal(groups, currentGroups) {
				t.Fatal("groups not equal", groups, currentGroups)
			}
		})
	}
}

func TestRisingWaveControllerManagerImpl_SyncMetaDeployments(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnly()

	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, false, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, false, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)
	// enable openKruise and check for tear down
	fakeRisingwave = testutils.FakeRisingWaveOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncMetaCloneSets(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWaveOpenKruiseEnabled()

	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnlyOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, true, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)

	// Disable Open Kruise and test for teardown of objects
	fakeRisingwave = testutils.FakeRisingWave()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentMeta,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncMetaAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentMeta {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncFrontendDeployments(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnly()

	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, false, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncFrontendDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, false, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncFrontendDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)

	// enable openKruise and check for tear down
	fakeRisingwave = testutils.FakeRisingWaveOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncFrontendDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncFrontendCloneSets(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWaveOpenKruiseEnabled()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnlyOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncFrontendCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, true, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncFrontendCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)

	// Disable Open Kruise and test for teardown of objects
	fakeRisingwave = testutils.FakeRisingWave()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentFrontend,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncFrontendCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentFrontend {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncCompactorDeployments(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnly()

	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, false, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncCompactorDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, false, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncCompactorDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)

	// enable openKruise and check for tear down
	fakeRisingwave = testutils.FakeRisingWaveOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.Deployment) (ctrl.Result, error) {
			return managerImpl.SyncCompactorDeployments(ctx, logger, obj)
		},
		func(tl *appsv1.DeploymentList) []appsv1.Deployment { return tl.Items },
		func(t *testing.T, obj *appsv1.Deployment) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncCompactorCloneSets(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWaveOpenKruiseEnabled()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnlyOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncCompactorCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, true, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncCompactorCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)

	// Disable Open Kruise and test for teardown of objects
	fakeRisingwave = testutils.FakeRisingWave()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompactor,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1alpha1.CloneSet) (ctrl.Result, error) {
			return managerImpl.SyncCompactorCloneSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1alpha1.CloneSetList) []kruiseappsv1alpha1.CloneSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1alpha1.CloneSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompactor {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncComputeStatefulSets(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWave()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnly()

	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, false, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, false, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)

	// enable openKruise and check for tear down
	fakeRisingwave = testutils.FakeRisingWaveOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []appsv1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeStatefulSets(ctx, logger, obj)
		},
		func(tl *appsv1.StatefulSetList) []appsv1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *appsv1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)
}

func TestRisingWaveControllerManagerImpl_SyncComputeAdvancedStatefulSets(t *testing.T) {
	fakeRisingwave := testutils.FakeRisingWaveOpenKruiseEnabled()
	fakeRisingwaveComponentOnly := testutils.FakeRisingWaveComponentOnlyOpenKruiseEnabled()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{""},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwaveComponentOnly, true, nil, []string{testutils.GetGroupName(0)},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwaveComponentOnly.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)
	// Disable Open Kruise and test for teardown of objects
	fakeRisingwave = testutils.FakeRisingWave()
	testRisingWaveControllerManagerImplSyncObjectGroups(
		t, fakeRisingwave, true, nil, []string{},
		map[string]string{
			consts.LabelRisingWaveName:      fakeRisingwave.Name,
			consts.LabelRisingWaveComponent: consts.ComponentCompute,
		},
		func(managerImpl *risingWaveControllerManagerImpl, ctx context.Context, logger logr.Logger, obj []kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
			return managerImpl.SyncComputeAdvancedStatefulSets(ctx, logger, obj)
		},
		func(tl *kruiseappsv1beta1.StatefulSetList) []kruiseappsv1beta1.StatefulSet { return tl.Items },
		func(t *testing.T, obj *kruiseappsv1beta1.StatefulSet) {
			if obj.Labels[consts.LabelRisingWaveComponent] != consts.ComponentCompute {
				t.Fatal("component labels not match")
			}
		},
	)
}

func Test_WaitComponentGroupWorkloadsReady(t *testing.T) {
	testcases := map[string]struct {
		groups  map[string]int
		objects []appsv1.Deployment
		ready   bool
	}{
		"objects-too-few": {
			groups: map[string]int{
				"": 1,
			},
			objects: nil,
			ready:   false,
		},
		"objects-not-ready": {
			groups: map[string]int{
				"": 1,
			},
			objects: []appsv1.Deployment{
				newGroupObjectFromGroup[appsv1.Deployment]("", "", "", map[string]string{}),
			},
			ready: false,
		},
		"objects-ready": {
			groups: map[string]int{
				"": 1,
			},
			objects: []appsv1.Deployment{
				newGroupObjectFromGroup[appsv1.Deployment]("", "", "", map[string]string{
					"ready": "1",
				}),
			},
			ready: true,
		},
		"objects-some-not-ready": {
			groups: map[string]int{
				"":                        1,
				testutils.GetGroupName(0): 1,
			},
			objects: []appsv1.Deployment{
				newGroupObjectFromGroup[appsv1.Deployment]("", "", "", map[string]string{
					"ready": "1",
				}),
				newGroupObjectFromGroup[appsv1.Deployment]("", "", testutils.GetGroupName(0), map[string]string{}),
			},
			ready: false,
		},
		"unexpected-groups": {
			groups: map[string]int{
				"": 1,
			},
			objects: []appsv1.Deployment{
				newGroupObjectFromGroup[appsv1.Deployment]("", "", "", map[string]string{
					"ready": "1",
				}),
				newGroupObjectFromGroup[appsv1.Deployment]("", "", testutils.GetGroupName(0), map[string]string{
					"ready": "1",
				}),
			},
			ready: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := waitComponentGroupWorkloadsReady(
				context.Background(), logr.Discard(), "", tc.groups, tc.objects,
				func(obj *appsv1.Deployment) bool {
					return obj.Labels["ready"] == "1"
				},
			)
			if ctrlkit.NeedsRequeue(r, err) == tc.ready {
				t.Fatal("mismatch, expect ", tc.ready)
			}
		})
	}
}
