package manager

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	Reconciles = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "reconciles_total",
			Help: "Number of reconciles",
		},
	)
	RequeueCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "requeue_total",
			Help: "Number of requeue. Incremented if any kind of requeue is needed",
		},
	)
	DidMutate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_mutated_request_total",
			Help: "Incremented if mutating webhooks mutates at least one attribute",
		},
	)
	ValidatingWebhookCalls = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_validated_called_total",
			Help: "Incremented if the validating webhook is called on Create/Delete/Update",
		},
	)
	ValidatingWebhookErr = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_validated_err_total",
			Help: "Incremented if the validating webhook ran into error Create/Delete/Update",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(Reconciles, RequeueCount, DidMutate, ValidatingWebhookCalls, ValidatingWebhookErr)
}
