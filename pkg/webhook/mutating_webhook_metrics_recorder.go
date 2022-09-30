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

	metrics "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	utils "github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type mutatingWebhook interface {
	webhook.CustomDefaulter
	getType() utils.WebhookType
}

// MutWebhookMetricsRecorder wraps a mutating webhook to simplify metric calculation.
type MutWebhookMetricsRecorder struct {
	webhook mutatingWebhook
}

func (r *MutWebhookMetricsRecorder) GetType() utils.WebhookType {
	return r.webhook.getType()
}

func (r *MutWebhookMetricsRecorder) recordAfter(err *error, obj runtime.Object, reconcileStartTS time.Time) error {
	if rec := recover(); rec != nil {
		metrics.IncWebhookRequestPanicCount(r.webhook.getType(), obj)
		metrics.IncWebhookRequestRejectCount(r.webhook.getType(), obj)
		return apierrors.NewInternalError(fmt.Errorf("panic in mutating webhook: %v", rec))
	}
	if *err != nil {
		metrics.IncWebhookRequestRejectCount(r.webhook.getType(), obj)
	} else {
		metrics.IncWebhookRequestPassCount(r.webhook.getType(), obj)
	}
	return *err
}

func (r *MutWebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	metrics.IncWebhookRequestCount(r.webhook.getType(), obj)
}

func (r *MutWebhookMetricsRecorder) Default(ctx context.Context, obj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	defer r.recordAfter(&err, obj, reconcileStartTS)
	r.recordBefore(obj)
	return r.webhook.Default(ctx, obj)
}

// CustomDefault required to implement webhook.CustomDefaulter.
func (r *MutWebhookMetricsRecorder) CustomDefaulter(ctx context.Context, obj runtime.Object) (err error) {
	return r.Default(ctx, obj)
}
