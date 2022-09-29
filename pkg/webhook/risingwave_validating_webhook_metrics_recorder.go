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

package webhook

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	utils "github.com/risingwavelabs/risingwave-operator/pkg/utils"

	metrics "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type validatingWebhook interface {
	webhook.CustomValidator
	getType() utils.WebhookType
}

// ValWebhookMetricsRecorder wraps a validating webhook to simplify metric calculation.
type ValWebhookMetricsRecorder struct {
	webhook validatingWebhook
}

func (v *ValWebhookMetricsRecorder) GetType() utils.WebhookType {
	return v.webhook.getType()
}

func (v *ValWebhookMetricsRecorder) recordAfter(err *error, obj runtime.Object, reconcileStartTS time.Time) error {
	if rec := recover(); rec != nil {
		metrics.IncWebhookRequestPanicCount(v.GetType(), obj)
		metrics.IncWebhookRequestRejectCount(v.GetType(), obj)
		return apierrors.NewInternalError(fmt.Errorf("panic in validating webhook: %v", rec))
	}
	if *err != nil {
		metrics.IncWebhookRequestRejectCount(v.GetType(), obj)
	} else {
		metrics.IncWebhookRequestPassCount(v.GetType(), obj)
	}
	return *err
}

func (v *ValWebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	metrics.IncWebhookRequestCount(v.GetType(), obj)
}

func (v *ValWebhookMetricsRecorder) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	defer v.recordAfter(&err, obj, reconcileStartTS)
	v.recordBefore(obj)
	return v.webhook.ValidateCreate(ctx, obj)
}

func (v *ValWebhookMetricsRecorder) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	defer v.recordAfter(&err, newObj, reconcileStartTS)
	v.recordBefore(newObj)
	return v.webhook.ValidateUpdate(ctx, oldObj, newObj)
}

func (v *ValWebhookMetricsRecorder) ValidateDelete(ctx context.Context, obj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	defer v.recordAfter(&err, obj, reconcileStartTS)
	v.recordBefore(obj)
	return v.webhook.ValidateDelete(ctx, obj)
}
