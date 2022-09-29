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

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	metrics "github.com/risingwavelabs/risingwave-operator/pkg/metrics"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
)

type RisingWaveMutatingWebhook struct{}

func setDefaultIfZero[T comparable](dst *T, defaultVal T) {
	var zero T
	if *dst == zero {
		*dst = defaultVal
	}
}

func (m *RisingWaveMutatingWebhook) GetType() metrics.WebhookType {
	return metrics.NewWebhookTypes(false)
}

func (m *RisingWaveMutatingWebhook) setDefault(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.ServicePort, consts.DefaultMetaServicePort)
	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.MetricsPort, consts.DefaultMetaMetricsPort)
	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.DashboardPort, consts.DefaultMetaDashboardPort)

	setDefaultIfZero(&obj.Spec.Components.Frontend.Ports.ServicePort, consts.DefaultFrontendServicePort)
	setDefaultIfZero(&obj.Spec.Components.Frontend.Ports.MetricsPort, consts.DefaultFrontendMetricsPort)

	setDefaultIfZero(&obj.Spec.Components.Compute.Ports.ServicePort, consts.DefaultComputeServicePort)
	setDefaultIfZero(&obj.Spec.Components.Compute.Ports.MetricsPort, consts.DefaultComputeMetricsPort)

	setDefaultIfZero(&obj.Spec.Components.Compactor.Ports.ServicePort, consts.DefaultCompactorServicePort)
	setDefaultIfZero(&obj.Spec.Components.Compactor.Ports.MetricsPort, consts.DefaultCompactorMetricsPort)

	return nil
}

// Default implements admission.CustomDefaulter.
func (m *RisingWaveMutatingWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return m.setDefault(ctx, obj.(*risingwavev1alpha1.RisingWave))
}

func NewRisingWaveMutatingWebhook() webhook.CustomDefaulter {
	return &MutWebhookMetricsRecorder{&RisingWaveMutatingWebhook{}}
}
