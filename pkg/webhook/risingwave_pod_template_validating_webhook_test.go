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

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

func Test_RisingWavePodTemplateValidatingWebhook_ValidateCreate(t *testing.T) {
	webhook := NewRisingWavePodTemplateValidatingWebhook()
	assert.Nil(t, webhook.ValidateCreate(context.Background(), nil))
}

func Test_RisingWavePodTemplateValidatingWebhook_ValidateDelete(t *testing.T) {
	webhook := NewRisingWavePodTemplateValidatingWebhook()
	assert.Nil(t, webhook.ValidateDelete(context.Background(), nil))
}

func Test_RisingWavePodTemplateValidatingWebhook_ValidateUpdate(t *testing.T) {
	template1 := &risingwavev1alpha1.RisingWavePodTemplate{
		Template: risingwavev1alpha1.RisingWavePodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "",
					},
				},
			},
		},
	}

	template2 := &risingwavev1alpha1.RisingWavePodTemplate{
		Template: risingwavev1alpha1.RisingWavePodTemplateSpec{
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name: "",
						SecurityContext: &corev1.SecurityContext{
							Privileged: pointer.Bool(true),
						},
					},
				},
			},
		},
	}

	webhook := NewRisingWavePodTemplateValidatingWebhook()
	if err := webhook.ValidateUpdate(context.Background(), template1, template1); err != nil {
		t.Fatal("same object should pass")
	}

	if err := webhook.ValidateUpdate(context.Background(), template1, template2); err == nil {
		t.Fatal("different object should fail")
	}
}
