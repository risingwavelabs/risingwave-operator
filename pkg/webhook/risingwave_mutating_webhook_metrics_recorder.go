package webhook

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type mutatingWebhook interface {
	Default(context.Context, runtime.Object) error
}

// WebhookMetricsRecorder wrapping a mutating webhook to simplify metric calculation.
type WebhookMetricsRecorder struct {
	webhook mutatingWebhook
}

func (r *WebhookMetricsRecorder) recordAfter(err error, obj runtime.Object) error {
	if err != nil {
		m.WebhookRequestRejectCount.Inc()
	} else {
		m.IncWebhookRequestPassCount(false, obj)
	}
	return err
}

func (r *WebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	m.IncWebhookRequestCount(true, obj)
}

func (r *WebhookMetricsRecorder) Default(ctx context.Context, obj runtime.Object) (err error) {
	r.recordBefore(obj)
	defer r.recordAfter(err, obj)

	return r.webhook.Default(ctx, obj)
}

// CustomDefault required to implement webhook.CustomDefaulter.
func (r *WebhookMetricsRecorder) CustomDefaulter(ctx context.Context, obj runtime.Object) (err error) {
	return r.Default(ctx, obj)
}

// TODO: Validating webhooks
