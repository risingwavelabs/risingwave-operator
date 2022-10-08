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

	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type RisingWavePodTemplateValidatingWebhook struct{}

// ValidateCreate implements admission.CustomValidator.
func (pt *RisingWavePodTemplateValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return nil
}

// ValidateDelete implements admission.CustomValidator.
func (pt *RisingWavePodTemplateValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (pt *RisingWavePodTemplateValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWavePodTemplate) error {
	gvk := oldObj.GroupVersionKind()

	if !equality.Semantic.DeepEqual(&oldObj.Template, &newObj.Template) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("template"), "template is immutable"),
		)
	}

	return nil
}

// ValidateUpdate implements admission.CustomValidator.
func (pt *RisingWavePodTemplateValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error {
	return pt.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWavePodTemplate), newObj.(*risingwavev1alpha1.RisingWavePodTemplate))
}

func NewRisingWavePodTemplateValidatingWebhook() webhook.CustomValidator {
	return metrics.NewValidatingWebhookMetricsRecorder(&RisingWavePodTemplateValidatingWebhook{})
}
