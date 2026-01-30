// Copyright 2023 RisingWave Labs
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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type webhookMetricsRecorder struct {
	webhookType utils.WebhookType
}

func (r *webhookMetricsRecorder) beforeInvoke(ctx context.Context, obj runtime.Object) {
	IncWebhookRequestCount(r.webhookType, obj)
}

func (r *webhookMetricsRecorder) afterInvoke(ctx context.Context, obj runtime.Object, startTime time.Time, err *error) {
	if rec := recover(); rec != nil {
		IncWebhookRequestPanicCount(r.webhookType, obj)
		IncWebhookRequestRejectCount(r.webhookType, obj)
		*err = apierrors.NewInternalError(fmt.Errorf("panic in %s webhook: %v", r.webhookType, rec))

		return
	}

	if *err != nil {
		IncWebhookRequestRejectCount(r.webhookType, obj)
	} else {
		IncWebhookRequestPassCount(r.webhookType, obj)
	}
}

type mutatingWebhookMetricsRecorder[T runtime.Object] struct {
	webhookMetricsRecorder
	inner admission.Defaulter[T]
}

// Default implements the admission.Defaulter.
func (r *mutatingWebhookMetricsRecorder[T]) Default(ctx context.Context, obj T) (err error) {
	startTime := time.Now()

	r.beforeInvoke(ctx, obj)
	defer r.afterInvoke(ctx, obj, startTime, &err)

	return r.inner.Default(ctx, obj)
}

// NewMutatingWebhookMetricsRecorder creates a new metrics recorder for the mutating webhook.
func NewMutatingWebhookMetricsRecorder[T runtime.Object](inner admission.Defaulter[T]) admission.Defaulter[T] {
	return &mutatingWebhookMetricsRecorder[T]{
		webhookMetricsRecorder: webhookMetricsRecorder{webhookType: utils.MutatingWebhookType},
		inner:                  inner,
	}
}

type validatingWebhookMetricsRecorder[T runtime.Object] struct {
	webhookMetricsRecorder
	inner admission.Validator[T]
}

// ValidateCreate implements the admission.Validator.
func (r *validatingWebhookMetricsRecorder[T]) ValidateCreate(ctx context.Context, obj T) (warnings admission.Warnings, err error) {
	startTime := time.Now()

	r.beforeInvoke(ctx, obj)
	defer r.afterInvoke(ctx, obj, startTime, &err)

	return r.inner.ValidateCreate(ctx, obj)
}

// ValidateUpdate implements the admission.Validator.
func (r *validatingWebhookMetricsRecorder[T]) ValidateUpdate(ctx context.Context, oldObj, newObj T) (warnings admission.Warnings, err error) {
	startTime := time.Now()

	r.beforeInvoke(ctx, newObj)
	defer r.afterInvoke(ctx, newObj, startTime, &err)

	return r.inner.ValidateUpdate(ctx, oldObj, newObj)
}

// ValidateDelete implements the admission.Validator.
func (r *validatingWebhookMetricsRecorder[T]) ValidateDelete(ctx context.Context, obj T) (warnings admission.Warnings, err error) {
	startTime := time.Now()

	r.beforeInvoke(ctx, obj)
	defer r.afterInvoke(ctx, obj, startTime, &err)

	return r.inner.ValidateDelete(ctx, obj)
}

// NewValidatingWebhookMetricsRecorder creates a new webhook recorder for validating webhook.
func NewValidatingWebhookMetricsRecorder[T runtime.Object](inner admission.Validator[T]) admission.Validator[T] {
	return &validatingWebhookMetricsRecorder[T]{
		webhookMetricsRecorder: webhookMetricsRecorder{webhookType: utils.ValidatingWebhookType},
		inner:                  inner,
	}
}
