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

package webhook

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveScaleViewValidatingWebhook_ValidateObject(t *testing.T) {
	testcases := map[string]struct {
		object    *risingwavev1alpha1.RisingWaveScaleView
		returnErr bool
	}{
		"good-0": {
			object: &risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
						Name:      "x",
						Component: consts.ComponentFrontend,
						UID:       "uid",
					},
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{Group: ""},
					},
				},
			},
			returnErr: false,
		},
		"good-1": {
			object: &risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
						Name:      "x",
						Component: consts.ComponentFrontend,
						UID:       "uid",
					},
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{Group: ""},
						{Group: "a", MaxReplicas: pointer.Int32(2)},
					},
				},
			},
			returnErr: false,
		},
		"bad-0": {
			object: &risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
						Component: consts.ComponentFrontend,
					},
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{Group: ""},
					},
				},
			},
			returnErr: true,
		},
		"bad-1": {
			object: &risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
						Name:      "x",
						Component: consts.ComponentFrontend,
						UID:       "uid",
					},
				},
			},
			returnErr: true,
		},
		"bad-2": {
			object: &risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					TargetRef: risingwavev1alpha1.RisingWaveScaleViewTargetRef{
						Name:      "x",
						Component: consts.ComponentFrontend,
						UID:       "uid",
					},
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{Group: "", MaxReplicas: pointer.Int32(1)},
					},
				},
			},
			returnErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			webhook := &RisingWaveScaleViewValidatingWebhook{}
			_, err := webhook.validateObject(context.Background(), tc.object)
			if tc.returnErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func risingwaveLockedBy(r *risingwavev1alpha1.RisingWave, v *risingwavev1alpha1.RisingWaveScaleView) *risingwavev1alpha1.RisingWave {
	err := object.NewScaleViewLockManager(r).GrabScaleViewLockFor(v)
	if err != nil {
		panic(err)
	}
	return r
}

func Test_RisingWaveScaleViewValidatingWebhook_ValidateCreate(t *testing.T) {
	goodScaleView := func() *risingwavev1alpha1.RisingWaveScaleView {
		return testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend,
			func(wave *risingwavev1alpha1.RisingWave, view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.UID = wave.UID
				view.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
					{Group: ""},
				}
			})
	}

	testcases := map[string]struct {
		initObjs  []client.Object
		scaleView *risingwavev1alpha1.RisingWaveScaleView
		returnErr bool
	}{
		"good": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			scaleView: goodScaleView(),
		},
		"bad-object": {
			scaleView: testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			returnErr: true,
		},
		"risingwave-not-found": {
			initObjs:  nil,
			scaleView: goodScaleView(),
			returnErr: true,
		},
		"risingwave-not-match": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			scaleView: testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend,
				func(wave *risingwavev1alpha1.RisingWave, view *risingwavev1alpha1.RisingWaveScaleView) {
					view.Spec.TargetRef.UID = types.UID(uuid.New().String())
					view.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{Group: ""},
					}
				}),
			returnErr: true,
		},
		"risingwave-already-locked": {
			initObjs: []client.Object{
				risingwaveLockedBy(testutils.FakeRisingWave(), goodScaleView()),
			},
			scaleView: goodScaleView(),
			returnErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			webhook := NewRisingWaveScaleViewValidatingWebhook(
				fake.NewClientBuilder().
					WithScheme(testutils.Scheme).
					WithStatusSubresource(&risingwavev1alpha1.RisingWave{}).
					WithObjects(tc.initObjs...).
					Build(),
			)
			_, err := webhook.ValidateCreate(context.Background(), tc.scaleView)
			if tc.returnErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func Test_RisingWaveScaleViewValidatingWebhook_ValidateUpdate(t *testing.T) {
	goodScaleView := func() *risingwavev1alpha1.RisingWaveScaleView {
		return testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend,
			func(wave *risingwavev1alpha1.RisingWave, view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.UID = wave.UID
				view.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
					{Group: ""},
				}
			})
	}

	testcases := map[string]struct {
		scaleView *risingwavev1alpha1.RisingWaveScaleView
		mutate    func(view *risingwavev1alpha1.RisingWaveScaleView)
		returnErr bool
	}{
		"update-target-ref-0": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.Name = rand.String(2)
			},
			returnErr: true,
		},
		"update-target-ref-1": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.Component = rand.String(2)
			},
			returnErr: true,
		},
		"update-target-ref-2": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.UID = types.UID(rand.String(2))
			},
			returnErr: true,
		},
		"update-label-selector": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.LabelSelector = rand.String(2)
			},
			returnErr: true,
		},
		"update-max-replicas": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.ScalePolicy[0].MaxReplicas = pointer.Int32(int32(rand.Int()))
			},
			returnErr: true,
		},
		"update-priority": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.ScalePolicy[0].Priority = int32(rand.Intn(10))
			},
		},
		"update-group-name": {
			scaleView: goodScaleView(),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.ScalePolicy[0].Group = rand.String(2)
			},
			returnErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			webhook := NewRisingWaveScaleViewValidatingWebhook(
				fake.NewClientBuilder().
					WithScheme(testutils.Scheme).
					WithStatusSubresource(&risingwavev1alpha1.RisingWave{}).
					Build(),
			)

			newObj := tc.scaleView.DeepCopy()
			tc.mutate(newObj)

			_, err := webhook.ValidateUpdate(context.Background(), tc.scaleView, newObj)
			if tc.returnErr {
				assert.NotNil(t, err, "error expected")
			} else {
				assert.Nil(t, err, "error unexpected")
			}
		})
	}
}

func Test_RisingWaveScaleViewValidatingWebhook_ValidateDelete(t *testing.T) {
	_, err := NewRisingWaveScaleViewValidatingWebhook(nil).ValidateDelete(context.Background(), &risingwavev1alpha1.RisingWaveScaleView{})
	assert.Nil(t, err)
}
