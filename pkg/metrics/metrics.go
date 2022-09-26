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

package manager

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// use constants to avoid string literals when using the counters.
const (
	mutatingWebhook   = "mutate"
	validatingWebhook = "validate"
)

// TODO: Do I need additional tests for this?
// TODO: Do some basic unit tests
// e.g manually trigger a panic and see if if panic got incremented
// hijack the result and see if requeue on err is called and so on. Err counter increased?

var (
	// Metric is used to test if metric collection works.
	ReceivingMetricsFromOperator = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receiving_metrics_from_operator",
			Help: "Value is 1, if you are able to receive custom metrics from the operator",
		},
	)

	// Webhook metrics vectors have the following attributes:
	// type: The value should be mutating or validating
	// group: The target resource group of the webhook, e.g., risingwave.risingwavelabs.com
	// version: The target API version, e.g., v1alpha1
	// kind: The target API kind, e.g., risingwave, risingwavepodtemplate
	// namespace: The namespace of the object, e.g., default
	// name: The name of the object
	// TODO: verb: The verb (action) on the object which triggers the webhook, the value should be one of "create", "update", and "delete".
	webhookRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_request_count",
			Help: "Total number of validating and mutating webhook calls",
		},
		[]string{"type", "group", "version", "kind", "namespace", "name"},
	)
	webhookRequestPassCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_request_pass_count",
			Help: "Total number of accepted validating and mutating webhook calls",
		},
		[]string{"type", "group", "version", "kind", "namespace", "name"},
	)
	webhookRequestRejectCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_request_reject_count",
			Help: "Total number of rejected validating and mutating webhook calls",
		},
		[]string{"type", "group", "version", "kind", "namespace", "name"},
	)
	webhookRequestPanicCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "webhook_request_panic_count",
			Help: "Total number of panics during validating and mutating webhook calls",
		},
		[]string{"type", "group", "version", "kind", "namespace", "name"},
	)

	// Controller metrics vectors have the following attributes
	// group: The target resource group of the webhook, e.g., risingwave.risingwavelabs.com
	// version: The target API version, e.g., v1alpha1
	// kind: The target API kind, e.g., risingwave, risingwavepodtemplate
	// namespace: The namespace of the object, e.g., default
	// name: The name of the object
	controllerReconcileCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_reconcile_count",
			Help: "Total number of reconciles",
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)
	controllerReconcileRequeueCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_reconcile_requeue_count",
			Help: "Total number of requeue",
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)
	controllerReconcileRequeueErrorCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_reconcile_error_count",
			Help: "Total number of requeue errors",
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)

	controllerReconcileRequeueAfter = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "controller_reconcile_requeue_after",
			Help:    "Delay of last delayed requeue in ms",
			Buckets: []float64{1, 10, 50, 100, 500, 1000}, // wait time in ms
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)
	controllerReconcilePanicCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_reconcile_panic_count",
			Help: "Total number of reconcile panics",
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)
)

// rwReqData hold specified fields of the RisingWave request.
type rwReqData struct {
	Namespace string
	Name      string
}

// getRWData returns the relevant data about the RisingWave request.
func getRWData(obj runtime.Object) rwReqData {
	rw, ok := obj.(*risingwavev1alpha1.RisingWave)
	if ok {
		return rwReqData{rw.Namespace, rw.Name}
	}
	rw2, _ := obj.(*risingwavev1alpha1.RisingWavePodTemplate)
	return rwReqData{rw2.Namespace, rw2.Name}
}

// incWebhooksWithLabelValues increments the webhooks metric counter 'metric' by one.
func incWebhooksWithLabelValues(metric prometheus.CounterVec, isValidating bool, obj runtime.Object) {
	type_ := mutatingWebhook
	if isValidating {
		type_ = validatingWebhook
	}
	gvk := obj.GetObjectKind().GroupVersionKind()
	reqData := getRWData(obj)
	metric.WithLabelValues(type_, gvk.Group, gvk.Version, gvk.Kind, reqData.Namespace, reqData.Name).Inc()
}

// IncWebhookRequestCount increments the WebhookRequestCount
// isValidating: If true then increments validating webhook, else mutating webhook.
func IncWebhookRequestCount(isValidating bool, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestCount, true, obj)
}

func IncWebhookRequestPassCount(isValidating bool, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestPassCount, isValidating, obj)
}

func IncWebhookRequestRejectCount(isValidating bool, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestRejectCount, isValidating, obj)
}

func IncWebhookRequestPanicCount(isValidating bool, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestPanicCount, isValidating, obj)
}

// incControllersWithLabelValues increments the controller metric counter 'metric' by one.
func incControllersWithLabelValues(metric prometheus.CounterVec, req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	gvk := rw.GroupVersionKind()
	metric.WithLabelValues(gvk.Group, gvk.Version, gvk.Kind, req.Namespace, req.Name).Inc()
}

func IncControllerReconcileCount(req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	incControllersWithLabelValues(*controllerReconcileCount, req, rw)
}

func IncControllerReconcilePanicCount(req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	incControllersWithLabelValues(*controllerReconcilePanicCount, req, rw)
}

func IncControllerReconcileRequeueCount(req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	incControllersWithLabelValues(*controllerReconcileRequeueCount, req, rw)
}

func IncControllerReconcileRequeueAfter(time_ms int64, req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	gvk := rw.GroupVersionKind()
	controllerReconcileRequeueAfter.WithLabelValues(gvk.Group, gvk.Version,
		gvk.Kind, req.Namespace, req.Name).Observe(float64(time_ms))
}

func IncControllerReconcileRequeueErrorCount(req reconcile.Request, rw risingwavev1alpha1.RisingWave) {
	incControllersWithLabelValues(*controllerReconcileRequeueErrorCount, req, rw)
}

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(controllerReconcileCount)
	metrics.Registry.MustRegister(controllerReconcilePanicCount)
	metrics.Registry.MustRegister(controllerReconcileRequeueAfter)
	metrics.Registry.MustRegister(controllerReconcileRequeueCount)
	metrics.Registry.MustRegister(controllerReconcileRequeueErrorCount)
	metrics.Registry.MustRegister(ReceivingMetricsFromOperator)
	metrics.Registry.MustRegister(webhookRequestCount)
	metrics.Registry.MustRegister(webhookRequestPanicCount)
	metrics.Registry.MustRegister(webhookRequestPassCount)
	metrics.Registry.MustRegister(webhookRequestRejectCount)
}
