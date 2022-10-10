// Copyright 2022 Singularity Data
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type panicValWebhook struct{}

func (p *panicValWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	panic("validateCreate panic")
}

func (p *panicValWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	panic("validateDelete update")
}

func (p *panicValWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	panic("validateUpdate panic")
}

func Test_MetricsValidatingWebhookPanic(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}
	panicWebhook := NewValidatingWebhookMetricsRecorder(&panicValWebhook{})

	_ = panicWebhook.ValidateCreate(context.Background(), risingwave)
	assert.Equal(t, 1, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()

	_ = panicWebhook.ValidateDelete(context.Background(), risingwave)
	assert.Equal(t, 1, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()

	_ = panicWebhook.ValidateUpdate(context.Background(), risingwave, risingwave)
	assert.Equal(t, 1, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()
}

type successfulValWebhook struct{}

func (s *successfulValWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	return nil
}

func (s *successfulValWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (s *successfulValWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	return nil
}

func Test_MetricsValidatingWebhookSuccess(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}
	successWebhook := NewValidatingWebhookMetricsRecorder(&successfulValWebhook{})

	_ = successWebhook.ValidateCreate(context.Background(), risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 0, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 1, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	ResetMetrics()

	_ = successWebhook.ValidateDelete(context.Background(), risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 0, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 1, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	ResetMetrics()

	_ = successWebhook.ValidateUpdate(context.Background(), risingwave, risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 0, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 1, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	ResetMetrics()
}

type errorValWebhook struct{}

func (e *errorValWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	return fmt.Errorf("validateCreate err")
}

func (e *errorValWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return fmt.Errorf("validateDelete err")
}

func (e *errorValWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	return fmt.Errorf("validateUpdate err")
}

func Test_MetricsValidatingWebhookError(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}
	errorWebhook := NewValidatingWebhookMetricsRecorder(&errorValWebhook{})

	_ = errorWebhook.ValidateCreate(context.Background(), risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()

	_ = errorWebhook.ValidateDelete(context.Background(), risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()

	_ = errorWebhook.ValidateUpdate(context.Background(), risingwave, risingwave)
	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.ValidatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.ValidatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.ValidatingWebhookType, risingwave), "Request metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.ValidatingWebhookType, risingwave), "Pass metric")
	ResetMetrics()
}

type panicMutWebhook struct{}

func (p *panicMutWebhook) Default(ctx context.Context, obj runtime.Object) error {
	panic("simulating a panic")
}

func Test_MetricsMutatingWebhookPanic(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	panicWebhook := NewMutatingWebhookMetricsRecorder(&panicMutWebhook{})
	_ = panicWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 1, GetWebhookRequestPanicCountWith(utils.MutatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.MutatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.MutatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.MutatingWebhookType, risingwave), "Pass metric")
}

type successfulMutWebhook struct{}

func (s *successfulMutWebhook) Default(ctx context.Context, obj runtime.Object) error { return nil }

func Test_MetricsMutatingWebhookSuccess(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	successWebhook := NewMutatingWebhookMetricsRecorder(&successfulMutWebhook{})
	_ = successWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.MutatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 0, GetWebhookRequestRejectCount(utils.MutatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.MutatingWebhookType, risingwave), "Count metric")
	assert.Equal(t, 1, GetWebhookRequestPassCount(utils.MutatingWebhookType, risingwave), "Request metric")
}

type errorMutWebhook struct{}

func (e *errorMutWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return fmt.Errorf("test error")
}

func Test_MetricsMutatingWebhookError(t *testing.T) {
	ResetMetrics()
	risingwave := &risingwavev1alpha1.RisingWave{}

	errorWebhook := NewMutatingWebhookMetricsRecorder(&errorMutWebhook{})
	_ = errorWebhook.Default(context.Background(), risingwave)

	assert.Equal(t, 0, GetWebhookRequestPanicCountWith(utils.MutatingWebhookType, risingwave), "Panic metric")
	assert.Equal(t, 1, GetWebhookRequestRejectCount(utils.MutatingWebhookType, risingwave), "Reject metric")
	assert.Equal(t, 1, GetWebhookRequestCount(utils.MutatingWebhookType, risingwave), "Request metric")
	assert.Equal(t, 0, GetWebhookRequestPassCount(utils.MutatingWebhookType, risingwave), "Pass metric")
}
