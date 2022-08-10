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

package webhook

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

func Test_RisingWavePodTemplateMutatingWebhook_Default(t *testing.T) {
	testcases := map[string]struct {
		obj *risingwavev1alpha1.RisingWavePodTemplate
	}{
		"empty-obj": {
			obj: &risingwavev1alpha1.RisingWavePodTemplate{},
		},
		"one-container": {
			obj: &risingwavev1alpha1.RisingWavePodTemplate{
				Template: risingwavev1alpha1.RisingWavePodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{}},
					},
				},
			},
		},
		"two-containers": {
			obj: &risingwavev1alpha1.RisingWavePodTemplate{
				Template: risingwavev1alpha1.RisingWavePodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{}, {}},
					},
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			hook := NewRisingWavePodTemplateMutatingWebhook()
			containersLen := len(tc.obj.Template.Spec.Containers)
			err := hook.Default(context.Background(), tc.obj)
			if err != nil {
				t.Fatal(err)
			}
			if containersLen > 0 && len(tc.obj.Template.Spec.Containers) != containersLen {
				t.Fatal("containers should be untouched")
			}
		})
	}
}
