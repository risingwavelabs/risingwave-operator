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
	"strings"

	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveScaleViewMutatingWebhook struct {
}

func GetScaleViewMinMaxConstraints(obj *risingwavev1alpha1.RisingWaveScaleView) (min, max int32) {
	min, max = int32(0), int32(0)
	for _, scalePolicy := range obj.Spec.ScalePolicy {
		min += scalePolicy.Constraints.Min
		if scalePolicy.Constraints.Max == 0 {
			max = 100000
		} else {
			max += scalePolicy.Constraints.Max
		}
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

func (w *RisingWaveScaleViewMutatingWebhook) splitReplicas(obj *risingwavev1alpha1.RisingWaveScaleView) error {
	if err := w.validateTheConstraints(obj); err != nil {
		if !ptrValueNotZero(obj.Spec.Strict) {
			return nil
		}
		return err
	}

	// TODO: actually split the replicas

	return nil
}

func (w *RisingWaveScaleViewMutatingWebhook) setDefault(ctx context.Context, obj *risingwavev1alpha1.RisingWaveScaleView) error {
	// Enforce the strict mode if not specified.
	if obj.Spec.Strict == nil {
		obj.Spec.Strict = pointer.Bool(true)
	}

	// Set the label selector.
	w.setLabelSelector(obj)

	// Split the total replicas.
	if err := w.splitReplicas(obj); err != nil {
		return err
	}

	return nil
}

func (w *RisingWaveScaleViewMutatingWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return w.setDefault(ctx, obj.(*risingwavev1alpha1.RisingWaveScaleView))
}

func NewRisingWaveScaleViewMutatingWebhook() webhook.CustomDefaulter {
	return &RisingWaveScaleViewMutatingWebhook{}
}
