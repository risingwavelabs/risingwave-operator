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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

// RisingWaveMutatingWebhook is the mutating webhook for RisingWaves.
type RisingWaveMutatingWebhook struct{}

// Default implements admission.CustomDefaulter.
func (m *RisingWaveMutatingWebhook) Default(ctx context.Context, obj runtime.Object) error {
	ConvertToV1alpha2Features(obj.(*risingwavev1alpha1.RisingWave))
	risingwave := obj.(*risingwavev1alpha1.RisingWave)
	risingwave.Spec.StateStore.DataDirectory = strings.TrimRight(strings.TrimSpace(risingwave.Spec.StateStore.DataDirectory), "/")
	return nil
}

// NewRisingWaveMutatingWebhook returns a new mutating webhook for RisingWaves.
func NewRisingWaveMutatingWebhook() webhook.CustomDefaulter {
	return metrics.NewMutatingWebhookMetricsRecorder(&RisingWaveMutatingWebhook{})
}
