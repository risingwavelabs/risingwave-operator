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

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type RisingWaveMutatingWebhook struct{}

func setDefaultIfZero[T comparable](dst *T, defaultVal T, didMutate *bool) {
	var zero T
	if *dst == zero {
		*dst = defaultVal
		*didMutate = true
	}
}

func (h *RisingWaveMutatingWebhook) setDefault(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	didMutate := false

	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.ServicePort, consts.DefaultMetaServicePort, &didMutate)
	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.MetricsPort, consts.DefaultMetaMetricsPort, &didMutate)
	setDefaultIfZero(&obj.Spec.Components.Meta.Ports.DashboardPort, consts.DefaultMetaDashboardPort, &didMutate)

	setDefaultIfZero(&obj.Spec.Components.Frontend.Ports.ServicePort, consts.DefaultFrontendServicePort, &didMutate)
	setDefaultIfZero(&obj.Spec.Components.Frontend.Ports.MetricsPort, consts.DefaultFrontendMetricsPort, &didMutate)

	setDefaultIfZero(&obj.Spec.Components.Compute.Ports.ServicePort, consts.DefaultComputeServicePort, &didMutate)
	setDefaultIfZero(&obj.Spec.Components.Compute.Ports.MetricsPort, consts.DefaultComputeMetricsPort, &didMutate)

	setDefaultIfZero(&obj.Spec.Components.Compactor.Ports.ServicePort, consts.DefaultCompactorServicePort, &didMutate)
	setDefaultIfZero(&obj.Spec.Components.Compactor.Ports.MetricsPort, consts.DefaultCompactorMetricsPort, &didMutate)

	if didMutate {
		m.DidMutate.Inc()
	}

	return nil
}

// Default implements admission.CustomDefaulter.
func (h *RisingWaveMutatingWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return h.setDefault(ctx, obj.(*risingwavev1alpha1.RisingWave))
}

func NewRisingWaveMutatingWebhook() webhook.CustomDefaulter {
	return &MutWebhookMetricsRecorder{&RisingWaveMutatingWebhook{}}
}
