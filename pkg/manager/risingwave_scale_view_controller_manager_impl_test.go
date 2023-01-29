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
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func TestRisingWaveScaleViewControllerManagerImpl_HandleScaleViewFinalizer(t *testing.T) {
	scaleView := testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend)
	scaleView.ResourceVersion = "1234"
	scaleView.Finalizers = []string{consts.FinalizerScaleView}
	scaleView.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
		{
			Group: "",
		},
	}

	risingwave := testutils.FakeRisingWave()
	risingwave.ResourceVersion = "1234"
	scaleView.Spec.TargetRef.UID = risingwave.UID
	err := object.NewScaleViewLockManager(risingwave).GrabScaleViewLockFor(scaleView)
	assert.Nil(t, err, "must grab the lock")

	client := fake.NewClientBuilder().
		WithScheme(testutils.Scheme).
		WithObjects(risingwave.DeepCopy(), scaleView.DeepCopy()).
		Build()

	impl := NewRisingWaveScaleViewControllerManagerImpl(client, scaleView)

	// Handle the finalizer.
	r, err := impl.HandleScaleViewFinalizer(context.Background(), logr.Discard(), risingwave)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, r, ctrl.Result{}, "should be empty")

	// Checks RisingWave and RisingWaveScaleView
	var remoteRisingWave risingwavev1alpha1.RisingWave
	_ = client.Get(context.Background(), types.NamespacedName{Namespace: risingwave.Namespace, Name: risingwave.Name}, &remoteRisingWave)

	locked := object.NewScaleViewLockManager(&remoteRisingWave).IsScaleViewLocked(scaleView)
	assert.False(t, locked, "must be unlocked")

	var remoteScaleView risingwavev1alpha1.RisingWaveScaleView
	_ = client.Get(context.Background(), types.NamespacedName{Namespace: scaleView.Namespace, Name: scaleView.Name}, &remoteScaleView)
	assert.NotContains(t, remoteScaleView.Finalizers, consts.FinalizerScaleView, "finalizer should be removed")
}

func TestRisingWaveScaleViewControllerManagerImpl_GrabOrUpdateScaleViewLock(t *testing.T) {
	scaleView := testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend)
	scaleView.ResourceVersion = "1234"
	scaleView.Finalizers = []string{consts.FinalizerScaleView}
	scaleView.Spec.Replicas = pointer.Int32(1)
	scaleView.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
		{
			Group: "",
		},
	}

	risingwave := testutils.FakeRisingWave()
	risingwave.ResourceVersion = "1234"
	scaleView.Spec.TargetRef.UID = risingwave.UID

	client := fake.NewClientBuilder().
		WithScheme(testutils.Scheme).
		WithObjects(risingwave.DeepCopy(), scaleView.DeepCopy()).
		Build()

	impl := NewRisingWaveScaleViewControllerManagerImpl(client, scaleView)

	// Grab the lock.
	r, err := impl.GrabOrUpdateScaleViewLock(context.Background(), logr.Discard(), risingwave)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, r, ctrl.Result{Requeue: true}, "should requeue immediately")

	// Checks RisingWave and RisingWaveScaleView
	_ = client.Get(context.Background(), types.NamespacedName{Namespace: risingwave.Namespace, Name: risingwave.Name}, risingwave)

	lock := object.NewScaleViewLockManager(risingwave).GetScaleViewLock(scaleView)
	assert.NotNil(t, lock, "must be locked")
	assert.Equal(t, lock.Generation, scaleView.Generation, "lock generation equals")

	// Run again. Nothing happens.
	r, err = impl.GrabOrUpdateScaleViewLock(context.Background(), logr.Discard(), risingwave)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, r, ctrl.Result{}, "should be empty")

	// Change the scale view, try updating lock.
	scaleView.Generation = 2
	scaleView.Spec.Replicas = pointer.Int32(2)

	r, err = impl.GrabOrUpdateScaleViewLock(context.Background(), logr.Discard(), risingwave)
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, r, ctrl.Result{Requeue: true}, "should requeue immediately")

	_ = client.Get(context.Background(), types.NamespacedName{Namespace: risingwave.Namespace, Name: risingwave.Name}, risingwave)

	lock = object.NewScaleViewLockManager(risingwave).GetScaleViewLock(scaleView)
	assert.NotNil(t, lock, "must be locked")
	assert.Equal(t, lock.Generation, scaleView.Generation, "lock generation equals")
}

func TestRisingWaveScaleViewControllerManagerImpl_UpdateScaleViewStatus(t *testing.T) {
	scaleView := testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend)
	scaleView.ResourceVersion = "1234"

	client := fake.NewClientBuilder().
		WithScheme(testutils.Scheme).
		WithObjects(scaleView.DeepCopy()).
		Build()

	impl := NewRisingWaveScaleViewControllerManagerImpl(client, scaleView)

	// Mutate the status outside the impl, but it should be recognized since currently it is directly referenced.
	scaleView.Status.Locked = !scaleView.Status.Locked

	r, err := impl.UpdateScaleViewStatus(context.Background(), logr.Discard())
	assert.Nil(t, err, "should be nil")
	assert.Equal(t, r, ctrl.Result{}, "should be empty")

	// Checks on the status.
	var remoteScaleView risingwavev1alpha1.RisingWaveScaleView
	_ = client.Get(context.Background(), types.NamespacedName{Namespace: scaleView.Namespace, Name: scaleView.Name}, &remoteScaleView)

	assert.True(t, equality.Semantic.DeepEqual(scaleView.Status, remoteScaleView.Status))
}
