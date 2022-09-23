package webhook

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type validatingWebhook interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateDelete(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error
}

// ValWebhookMetricsRecorder wrapping a mutating webhook to simplify metric calculation.
type ValWebhookMetricsRecorder struct {
	webhook validatingWebhook
}

func (v *ValWebhookMetricsRecorder) recordAfter(err error, obj runtime.Object) error {
	if rec := recover(); rec != nil {
		m.WebhookRequestPanicCount.Inc()
	}
	// TODO: Do we want to record the request reject/pass count if we panic?
	if err != nil {
		m.WebhookRequestRejectCount.Inc()
	} else {
		m.IncWebhookRequestPassCount(true, obj)
	}
	return err
}

func (v *ValWebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	m.IncWebhookRequestCount(true, obj)
}

func (v *ValWebhookMetricsRecorder) ValidateCreate(ctx context.Context, obj runtime.Object) (err error) {
	defer v.recordAfter(err, obj)
	v.recordBefore(obj)
	return v.webhook.ValidateCreate(ctx, obj)
}

func (v *ValWebhookMetricsRecorder) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (err error) {
	defer v.recordAfter(err, newObj)
	v.recordBefore(newObj)
	return v.webhook.ValidateUpdate(ctx, oldObj, newObj)
}

func (v *ValWebhookMetricsRecorder) ValidateDelete(ctx context.Context, obj runtime.Object) (err error) {
	defer v.recordAfter(err, obj)
	v.recordBefore(obj)
	return v.webhook.ValidateDelete(ctx, obj)
}
