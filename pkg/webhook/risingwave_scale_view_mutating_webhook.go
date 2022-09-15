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
	"bytes"
	"context"
	"errors"
	"math"
	"strings"

	"github.com/samber/lo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/scaleview"
)

type RisingWaveScaleViewMutatingWebhook struct {
	client client.Client
}

func GetScaleViewMinMaxConstraints(obj *risingwavev1alpha1.RisingWaveScaleView) (min, max int32) {
	min, max = int32(0), int32(0)
	for _, scalePolicy := range obj.Spec.ScalePolicy {
		min += scalePolicy.Constraints.Min
	}

	for _, scalePolicy := range obj.Spec.ScalePolicy {
		if scalePolicy.Constraints.Max == 0 {
			max = math.MaxInt32
			break
		}
		max += scalePolicy.Constraints.Max
	}

	return min, max
}

func (w *RisingWaveScaleViewMutatingWebhook) validateTheConstraints(obj *risingwavev1alpha1.RisingWaveScaleView) error {
	min, max := GetScaleViewMinMaxConstraints(obj)

	if obj.Spec.Replicas < min || obj.Spec.Replicas > max {
		return errors.New("replicas out of range")
	}

	return nil
}

func (w *RisingWaveScaleViewMutatingWebhook) setLabelSelector(obj *risingwavev1alpha1.RisingWaveScaleView) {
	labelBuilder := &bytes.Buffer{}

	labelBuilder.WriteString("risingwave/name=")
	labelBuilder.WriteString(obj.Spec.TargetRef.Name)
	labelBuilder.WriteRune(',')

	labelBuilder.WriteString("risingwave/component=")
	labelBuilder.WriteString(obj.Spec.TargetRef.Component)
	labelBuilder.WriteRune(',')

	labelBuilder.WriteString("risingwave/group in (")

	groups := lo.Map(obj.Spec.ScalePolicy, func(t risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy, _ int) string { return t.Group })
	labelBuilder.WriteString(strings.Join(groups, ","))

	labelBuilder.WriteRune(')')

	obj.Spec.LabelSelector = labelBuilder.String()
}

func (w *RisingWaveScaleViewMutatingWebhook) updateGroupReplicas(obj *risingwavev1alpha1.RisingWaveScaleView) error {
	if err := w.validateTheConstraints(obj); err != nil {
		if !ptrValueNotZero(obj.Spec.Strict) {
			return nil
		}
		return err
	}

	// TODO: actually split the replicas
	obj.Spec.ScalePolicy[0].Replicas = obj.Spec.Replicas

	return nil
}

func (w *RisingWaveScaleViewMutatingWebhook) readGroupReplicasFromRisingWave(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) error {
	var targetObj risingwavev1alpha1.RisingWave
	if err := w.client.Get(ctx, types.NamespacedName{
		Namespace: obj.Namespace,
		Name:      obj.Spec.TargetRef.Name,
	}, &targetObj); err != nil {
		if apierrors.IsNotFound(err) {
			gvk := obj.GroupVersionKind()
			return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, field.ErrorList{
				field.Invalid(field.NewPath("spec", "targetRef"), obj.Spec.TargetRef, "target risingwave not found"),
			})
		}
	}

	obj.Spec.TargetRef.UID = targetObj.UID

	fieldErrs := field.ErrorList{}
	replicas := int32(0)
	mgr := scaleview.NewComponentGroupReplicasManager(&targetObj, obj.Spec.TargetRef.Component)
	for i := range obj.Spec.ScalePolicy {
		scalePolicy := &obj.Spec.ScalePolicy[i]
		if r, ok := mgr.ReadReplicas(scalePolicy.Group); ok {
			if r < scalePolicy.Constraints.Min || (scalePolicy.Constraints.Max > 0 && r > scalePolicy.Constraints.Max) {
				fieldErrs = append(fieldErrs, field.Invalid(
					field.NewPath("spec", "scalePolicy").Index(i).Key("replicas"),
					r,
					"replicas of RisingWave out of range"),
				)
			}
			scalePolicy.Replicas = r
			replicas += r
		} else {
			fieldErrs = append(fieldErrs, field.Invalid(
				field.NewPath("spec", "scalePolicy").Index(i),
				*scalePolicy,
				"target group not found"),
			)
		}
	}
	obj.Spec.Replicas = replicas

	if len(fieldErrs) > 0 {
		gvk := obj.GroupVersionKind()
		return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, field.ErrorList{
			field.Invalid(field.NewPath("spec", "targetRef"), obj.Spec.TargetRef, "target risingwave not found"),
		})
	}

	// Set the finalizer if everything's ok.
	if !controllerutil.ContainsFinalizer(obj, consts.FinalizerScaleView) {
		controllerutil.AddFinalizer(obj, consts.FinalizerScaleView)
	}

	return nil
}

// Set default values.
// - Enforce strict mode if not specified.
// - Set the target group to "" if the .spec.scalePolicy is empty
// - Update the .spec.labelSelector
// - Read or update the .spec.replicas and group replicas under .spec.scalePolicy
//   - if it's a new created scale policy (with empty .spec.targetRef.uid), get the target RisingWave and read and copy the replicas.
//   - otherwise, update the group replicas according to the current .spec.replicas
func (w *RisingWaveScaleViewMutatingWebhook) setDefault(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) error {
	if obj.Spec.Strict == nil {
		obj.Spec.Strict = pointer.Bool(true)
	}

	if len(obj.Spec.ScalePolicy) == 0 {
		obj.Spec.ScalePolicy = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
			{Group: ""},
		}
	}

	w.setLabelSelector(obj)

	var err error
	if obj.Spec.TargetRef.UID == "" {
		err = w.readGroupReplicasFromRisingWave(ctx, obj)
	} else {
		err = w.updateGroupReplicas(obj)
	}
	if err != nil {
		return err
	}

	return nil
}

func (w *RisingWaveScaleViewMutatingWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return w.setDefault(ctx, obj.(*risingwavev1alpha1.RisingWaveScaleView))
}

func NewRisingWaveScaleViewMutatingWebhook(client client.Client) webhook.CustomDefaulter {
	return &RisingWaveScaleViewMutatingWebhook{
		client: client,
	}
}
