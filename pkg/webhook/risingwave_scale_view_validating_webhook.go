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
	"fmt"
	"math"
	"sort"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/strings/slices"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
)

// RisingWaveScaleViewValidatingWebhook is the validating webhook for RisingWaveScaleView.
type RisingWaveScaleViewValidatingWebhook struct {
	client client.Reader
}

func getScaleViewMaxConstraints(obj *risingwavev1alpha1.RisingWaveScaleView) int32 {
	maxValue := int32(0)

	for _, scalePolicy := range obj.Spec.ScalePolicy {
		if scalePolicy.MaxReplicas == nil {
			maxValue = math.MaxInt32

			break
		}

		maxValue += *scalePolicy.MaxReplicas
	}

	return maxValue
}

func (w *RisingWaveScaleViewValidatingWebhook) validateObject(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) (warnings admission.Warnings, err error) {
	fieldErrs := field.ErrorList{}

	targetRefPath := field.NewPath("spec", "targetRef")
	if obj.Spec.TargetRef.Name == "" {
		fieldErrs = append(fieldErrs, field.Required(targetRefPath.Child("name"), "target name must be provided"))
	}

	if obj.Spec.TargetRef.UID == "" {
		fieldErrs = append(fieldErrs, field.Required(targetRefPath.Child("uid"), "uid should be set by mutating webhook"))
	}

	scalePolicyPath := field.NewPath("spec", "scalePolicy")
	if len(obj.Spec.ScalePolicy) == 0 {
		fieldErrs = append(fieldErrs, field.Required(scalePolicyPath, "must not be empty"))
	}

	if getScaleViewMaxConstraints(obj) != math.MaxInt32 {
		fieldErrs = append(fieldErrs, field.Invalid(scalePolicyPath, obj.Spec.ScalePolicy, "at least one unlimited replicas"))
	}

	if len(fieldErrs) > 0 {
		gvk := obj.GroupVersionKind()

		return nil, apierrors.NewInvalid(
			gvk.GroupKind(),
			obj.Name,
			fieldErrs,
		)
	}

	return nil, nil
}

func (w *RisingWaveScaleViewValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) (warnings admission.Warnings, err error) {
	if warnings, err := w.validateObject(ctx, obj); err != nil {
		return warnings, err
	}

	var risingwave risingwavev1alpha1.RisingWave

	err = w.client.Get(ctx, types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Spec.TargetRef.Name,
	}, &risingwave)
	if err != nil {
		return nil, fmt.Errorf("unable to get the risingwave: %w", err)
	}

	if obj.Spec.TargetRef.UID != risingwave.UID {
		return nil, fmt.Errorf("risingwave not match, expect uid: %s, but is: %s", obj.Spec.TargetRef.UID, risingwave.UID)
	}

	// Try grab the lock to see if there are conflicts.
	if err := object.NewScaleViewLockManager(&risingwave).GrabScaleViewLockFor(obj); err != nil {
		gvk := obj.GroupVersionKind()

		return nil, apierrors.NewConflict(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			obj.Name,
			fmt.Errorf("conflict detected: %w", err),
		)
	}

	return nil, nil
}

// ValidateCreate implements the webhook.CustomValidator.
func (w *RisingWaveScaleViewValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
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

	if oldObj.Spec.LabelSelector != newObj.Spec.LabelSelector {
		fieldErrs = append(fieldErrs, field.Forbidden(
			field.NewPath("spec", "labelSelector"),
			"labelSelector should not be changed",
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

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), oldObj.Name, fieldErrs)
	}

	return nil
}

// ValidateUpdate implements the webhook.CustomValidator.
func (w *RisingWaveScaleViewValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (warnings admission.Warnings, err error) {
	// Validate the new object first.
	if warnings, err := w.validateObject(ctx, newObj.(*risingwavev1alpha1.RisingWaveScaleView)); err != nil {
		return warnings, err
	}

	err = w.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWaveScaleView), newObj.(*risingwavev1alpha1.RisingWaveScaleView))

	return nil, err
}

// ValidateDelete implements the webhook.CustomValidator.
func (w *RisingWaveScaleViewValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

// NewRisingWaveScaleViewValidatingWebhook returns a new validator for RisingWaveScaleViews.
func NewRisingWaveScaleViewValidatingWebhook(client client.Reader) webhook.CustomValidator {
	return metrics.NewValidatingWebhookMetricsRecorder(&RisingWaveScaleViewValidatingWebhook{client: client})
}
