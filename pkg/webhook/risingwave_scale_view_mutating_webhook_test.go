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

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveScaleViewMutatingWebhook_Default(t *testing.T) {
	testcases := map[string]struct {
		initObjs  []client.Object
		origin    *risingwavev1alpha1.RisingWaveScaleView
		mutate    func(view *risingwavev1alpha1.RisingWaveScaleView)
		returnErr bool
	}{
		"risingwave-not-found": {
			initObjs:  nil,
			origin:    testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			returnErr: true,
		},
		"group-not-found": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			origin: testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
					{
						Group: "not-found",
					},
				}
			},
			returnErr: true,
		},
		"good-0": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			origin:    testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			returnErr: false,
		},
		"replicas-out-of-range": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			origin: testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
					{
						Group:       "",
						MaxReplicas: pointer.Int32(0),
					},
				}
			},
			returnErr: true,
		},
		"uid-not-empty": {
			initObjs: []client.Object{
				testutils.FakeRisingWave(),
			},
			origin: testutils.NewFakeRisingWaveScaleViewFor(testutils.FakeRisingWave(), consts.ComponentFrontend),
			mutate: func(view *risingwavev1alpha1.RisingWaveScaleView) {
				view.Spec.TargetRef.UID = "123"
			},
			returnErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			webhook := NewRisingWaveScaleViewMutatingWebhook(
				fake.NewClientBuilder().
					WithScheme(testutils.Scheme).
					WithObjects(tc.initObjs...).
					Build(),
			)

			if tc.mutate != nil {
				tc.mutate(tc.origin)
			}

			obj := tc.origin.DeepCopy()
			err := webhook.Default(context.Background(), obj)
			if tc.returnErr && err == nil {
				t.Fatal("error expected, but nil")
			} else if !tc.returnErr && err != nil {
				t.Fatal("unexpected err:", err)
			}

			if err == nil {
				// Run checks.
				spec := obj.Spec
				assert.NotEmpty(t, spec.TargetRef.UID, "uid should be set")
				assert.NotEmpty(t, spec.ScalePolicy, "scale policy should be set")
				assert.NotEmpty(t, spec.LabelSelector, "label selector should be set")
			}
		})
	}
}
