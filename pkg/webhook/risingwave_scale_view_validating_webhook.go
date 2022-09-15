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
	"math"
	"sort"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveScaleViewValidatingWebhook struct {
	client client.Client
}

func getScaleViewMaxConstraints(obj *risingwavev1alpha1.RisingWaveScaleView) int32 {
	max := int32(0)
	for _, scalePolicy := range obj.Spec.ScalePolicy {
		if scalePolicy.Constraints.Max == 0 {
			max = math.MaxInt32
			break
		}
		max += scalePolicy.Constraints.Max
	}

	return max
}

func (w *RisingWaveScaleViewValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) error {
	if getScaleViewMaxConstraints(obj) != math.MaxInt32 {
		gvk := obj.GroupVersionKind()
		return apierrors.NewInvalid(
			gvk.GroupKind(),
			obj.Name,
			field.ErrorList{
				field.Invalid(field.NewPath("spec", "scalePolicy"), obj.Spec.ScalePolicy, "at least one unlimited replicas"),
			},
		)
	}

	return nil
}

func (w *RisingWaveScaleViewValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return w.validateCreate(ctx, obj.(*risingwavev1alpha1.RisingWaveScaleView))
}

func (w *RisingWaveScaleViewValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWaveScaleView) error {
	gvk := oldObj.GroupVersionKind()
	fieldErrs := field.ErrorList{}

	if !equality.Semantic.DeepEqual(oldObj.Spec.TargetRef, newObj.Spec.TargetRef) {
		fieldErrs = append(fieldErrs, field.Forbidden(
			field.NewPath("spec", "targetRef"),
			"targetRefs should not be different",
		))
	}

	oldGroupList := lo.Map(oldObj.Spec.ScalePolicy, func(t risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy, _ int) string { return t.Group })
	newGroupList := lo.Map(newObj.Spec.ScalePolicy, func(t risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy, _ int) string { return t.Group })
	sort.Strings(oldGroupList)
	sort.Strings(newGroupList)
	if !slices.Equal(oldGroupList, newGroupList) {
		fieldErrs = append(fieldErrs, field.Forbidden(
			field.NewPath("spec", "scalePolicy"),
			"groups should no be changed",
		))
	}

	if !newObj.Status.Locked && newObj.Spec.Replicas != oldObj.Spec.Replicas {
		fieldErrs = append(fieldErrs, field.Forbidden(
			field.NewPath("spec", "replicas"),
			"update is forbidden before lock's grabbed",
		))
	}

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), oldObj.Name, fieldErrs)
	}

	return nil
}

func (w *RisingWaveScaleViewValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	// Validate the new object first.
	if err := w.ValidateCreate(ctx, newObj); err != nil {
		return err
	}

	return w.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWaveScaleView), newObj.(*risingwavev1alpha1.RisingWaveScaleView))
}

func (w *RisingWaveScaleViewValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func NewRisingWaveScaleViewValidatingWebhook(client client.Client) webhook.CustomValidator {
	return &RisingWaveScaleViewValidatingWebhook{
		client: client,
	}
}
