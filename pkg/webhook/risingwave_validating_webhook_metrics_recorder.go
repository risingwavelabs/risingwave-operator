package webhook

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"

	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type validatingWebhook interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateDelete(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error
	GetName() string
}

// ValWebhookMetricsRecorder wrapping a mutating webhook to simplify metric calculation.
type ValWebhookMetricsRecorder struct {
	webhook validatingWebhook
}

func (v *ValWebhookMetricsRecorder) recordAfter(err *error, obj runtime.Object, reconcileStartTS time.Time) error {
	if rec := recover(); rec != nil {
		m.IncWebhookRequestPanicCount(true, obj)
		m.IncWebhookRequestRejectCount(true, obj)
		return apierrors.NewInternalError(fmt.Errorf("panic in validating webhook: %v", rec))
	}
	if *err != nil {
		m.IncWebhookRequestRejectCount(true, obj)
	} else {
		m.IncWebhookRequestPassCount(true, obj)
	}
	m.UpdateControllerReconcileDuration(time.Since(reconcileStartTS).Milliseconds(), obj, v.webhook.GetName())
	return *err
}

func (v *ValWebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	m.IncWebhookRequestCount(true, obj)
}

func (v *ValWebhookMetricsRecorder) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	v.recordBefore(obj)
	defer v.recordAfter(&err, obj, reconcileStartTS)
	return v.webhook.ValidateCreate(ctx, obj)
}

func (v *ValWebhookMetricsRecorder) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	v.recordBefore(newObj)
	defer v.recordAfter(&err, newObj, reconcileStartTS)
	return v.webhook.ValidateUpdate(ctx, oldObj, newObj)
}

func (v *ValWebhookMetricsRecorder) ValidateDelete(ctx context.Context, obj runtime.Object) (err error) {
	reconcileStartTS := time.Now()
	v.recordBefore(obj)
	defer v.recordAfter(&err, obj, reconcileStartTS)
	return v.webhook.ValidateDelete(ctx, obj)
}