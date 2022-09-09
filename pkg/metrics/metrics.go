package manager

import (
	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	Goobers = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "goobers_total",
			Help: "Number of goobers processed",
		},
	)
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
			Name: "mutated_something",
			Help: "Incremented if mutating webhooks mutates at least one attribute",
		},
	)
)

func init() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(Goobers, Reconciles, RequeueCount, DidMutate)
}
