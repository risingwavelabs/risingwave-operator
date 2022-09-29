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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	metrics "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveMutatingWebhook_Default(t *testing.T) {
	risingwave := &risingwavev1alpha1.RisingWave{}

	NewRisingWaveMutatingWebhook().Default(context.Background(), risingwave)

	if !testutils.DeepEqual(risingwave, &risingwavev1alpha1.RisingWave{
		Spec: risingwavev1alpha1.RisingWaveSpec{
			Components: risingwavev1alpha1.RisingWaveComponentsSpec{
				Meta: risingwavev1alpha1.RisingWaveComponentMeta{
					Ports: risingwavev1alpha1.RisingWaveComponentMetaPorts{
						RisingWaveComponentCommonPorts: risingwavev1alpha1.RisingWaveComponentCommonPorts{
							ServicePort: consts.DefaultMetaServicePort,
							MetricsPort: consts.DefaultMetaMetricsPort,
						},
						DashboardPort: consts.DefaultMetaDashboardPort,
					},
				},
				Frontend: risingwavev1alpha1.RisingWaveComponentFrontend{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultFrontendServicePort,
						MetricsPort: consts.DefaultFrontendMetricsPort,
					},
				},
				Compute: risingwavev1alpha1.RisingWaveComponentCompute{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultComputeServicePort,
						MetricsPort: consts.DefaultComputeMetricsPort,
					},
				},
				Compactor: risingwavev1alpha1.RisingWaveComponentCompactor{
					Ports: risingwavev1alpha1.RisingWaveComponentCommonPorts{
						ServicePort: consts.DefaultCompactorServicePort,
						MetricsPort: consts.DefaultCompactorMetricsPort,
					},
				},
			},
		},
	}) {
		t.Fail()
	}
}

type panicMutWebhook struct{}

func (p *panicMutWebhook) Default(ctx context.Context, obj runtime.Object) error {
	panic("simulating a panic")
}

func (p *panicMutWebhook) GetType() metrics.WebhookType {
	return metrics.NewWebhookTypes(false)
}

func Test_MetricsMutatingWebhookPanic(t *testing.T) {
	metrics.ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	panicWebhook := &MutWebhookMetricsRecorder{&panicMutWebhook{}}
	panicWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 1, metrics.GetWebhookRequestPanicCountWith(panicWebhook.GetType(), risingwave), "Panic metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestRejectCount(panicWebhook.GetType(), risingwave), "Reject metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestCount(panicWebhook.GetType(), risingwave), "Count metric")
	assert.Equal(t, 0, metrics.GetWebhookRequestPassCount(panicWebhook.GetType(), risingwave), "Pass metric")
}

type successfulMutWebhook struct{}

func (s *successfulMutWebhook) Default(ctx context.Context, obj runtime.Object) error { return nil }

func (s *successfulMutWebhook) GetType() metrics.WebhookType {
	return metrics.NewWebhookTypes(false)
}

func Test_MetricsMutatingWebhookSuccess(t *testing.T) {
	metrics.ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	successWebhook := &MutWebhookMetricsRecorder{&successfulMutWebhook{}}
	successWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 0, metrics.GetWebhookRequestPanicCountWith(successWebhook.GetType(), risingwave), "Panic metric")
	assert.Equal(t, 0, metrics.GetWebhookRequestRejectCount(successWebhook.GetType(), risingwave), "Reject metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestCount(successWebhook.GetType(), risingwave), "Count metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestPassCount(successWebhook.GetType(), risingwave), "Request metric")
}

type errorMutWebhook struct{}

func (e *errorMutWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return fmt.Errorf("test error")
}

func (e *errorMutWebhook) GetType() metrics.WebhookType {
	return metrics.NewWebhookTypes(false)
}

func Test_MetricsMutatingWebhookError(t *testing.T) {
	metrics.ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	errorWebhook := &MutWebhookMetricsRecorder{&errorMutWebhook{}}
	errorWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 0, metrics.GetWebhookRequestPanicCountWith(errorWebhook.GetType(), risingwave), "Panic metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestRejectCount(errorWebhook.GetType(), risingwave), "Reject metric")
	assert.Equal(t, 1, metrics.GetWebhookRequestCount(errorWebhook.GetType(), risingwave), "Request metric")
	assert.Equal(t, 0, metrics.GetWebhookRequestPassCount(errorWebhook.GetType(), risingwave), "Pass metric")
}
