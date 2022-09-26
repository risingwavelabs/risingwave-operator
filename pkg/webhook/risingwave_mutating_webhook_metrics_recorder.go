package webhook

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"

	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
)

type mutatingWebhook interface {
	Default(context.Context, runtime.Object) error
}

// MutWebhookMetricsRecorder wrapping a mutating webhook to simplify metric calculation.
type MutWebhookMetricsRecorder struct {
	webhook mutatingWebhook
}

func (r *MutWebhookMetricsRecorder) recordAfter(err error, obj runtime.Object) error {
	if rec := recover(); rec != nil {
		m.IncWebhookRequestPanicCount(false, obj)
	}
	// TODO: Do we want to record the request reject/pass count if we panic?
	if err != nil {
		m.IncWebhookRequestRejectCount(false, obj)
	} else {
		m.IncWebhookRequestPassCount(false, obj)
	}
	return err
}

func (r *MutWebhookMetricsRecorder) recordBefore(obj runtime.Object) {
	m.IncWebhookRequestCount(true, obj)
}

func (r *MutWebhookMetricsRecorder) Default(ctx context.Context, obj runtime.Object) (err error) {
	// TODO: Check if call panicked
	r.recordBefore(obj)
	defer r.recordAfter(err, obj)

	return r.webhook.Default(ctx, obj)
}

// CustomDefault required to implement webhook.CustomDefaulter.
func (r *MutWebhookMetricsRecorder) CustomDefaulter(ctx context.Context, obj runtime.Object) (err error) {
	return r.Default(ctx, obj)
}

// TODO: Validating webhooks
