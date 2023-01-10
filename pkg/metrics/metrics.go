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

package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	prometheusclient "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

var (
	// ReceivingMetricsFromOperator is used to test if metric collection works.
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
	// TODO: verb: The verb (action) on the object which triggers the webhook, the value should be one of "create", "update", and "delete".
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
	controllerReconcileDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "controller_reconcile_duration",
		Help: "Length of time per reconciliation per controller",
		Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.15, 0.2, 0.25, 0.3, 0.35, 0.4, 0.45, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0,
			1.25, 1.5, 1.75, 2.0, 2.5, 3.0, 3.5, 4.0, 4.5, 5, 6, 7, 8, 9, 10, 15, 20, 25, 30, 40, 50, 60},
	}, []string{"controller", "group", "version", "kind", "namespace", "name"})
	controllerReconcilePanicCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "controller_reconcile_panic_count",
			Help: "Total number of reconcile panics",
		},
		[]string{"group", "version", "kind", "namespace", "name"},
	)
)

// toNamespacedName returns the relevant data about the RisingWave request.
func toNamespacedName(obj client.Object) types.NamespacedName {
	return types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}
}

// incWebhooksWithLabelValues increments the webhooks metric counter 'metric' by one.
func incWebhooksWithLabelValues(metric prometheus.CounterVec, wt utils.WebhookType, obj runtime.Object) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	nn := toNamespacedName(obj.(client.Object))
	metric.WithLabelValues(wt.String(), gvk.Group, gvk.Version, gvk.Kind, nn.Namespace, nn.Name).Inc()
}

// getWebhooksWithLabelValues returns the numeric current counter of a metric.
func getWebhooksWithLabelValues(metric prometheus.CounterVec, wt utils.WebhookType, obj runtime.Object) int {
	gvk := obj.GetObjectKind().GroupVersionKind()
	nn := toNamespacedName(obj.(client.Object))
	counter, _ := metric.GetMetricWith(prometheus.Labels{
		"type": wt.String(), "group": gvk.Group, "version": gvk.Version,
		"kind": gvk.Version, "namespace": nn.Namespace, "name": nn.Name,
	})
	var m prometheusclient.Metric
	_ = counter.Write(&m)
	return int(*m.Counter.Value)
}

// IncWebhookRequestCount increases the request count for the given webhook type and target object by 1.
func IncWebhookRequestCount(wt utils.WebhookType, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestCount, wt, obj)
}

// IncWebhookRequestPassCount increases the request pass count for the given webhook type and target object by 1.
func IncWebhookRequestPassCount(wt utils.WebhookType, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestPassCount, wt, obj)
}

// IncWebhookRequestRejectCount increases the request reject count for the given webhook type and target object by 1.
func IncWebhookRequestRejectCount(wt utils.WebhookType, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestRejectCount, wt, obj)
}

// IncWebhookRequestPanicCount increases the request panic count for the given webhook type and target object by 1.
func IncWebhookRequestPanicCount(wt utils.WebhookType, obj runtime.Object) {
	incWebhooksWithLabelValues(*webhookRequestPanicCount, wt, obj)
}

// Get webhook metric count

// GetWebhookRequestPanicCountWith gets the request panic count for the given webhook type and target object.
func GetWebhookRequestPanicCountWith(wt utils.WebhookType, obj runtime.Object) int {
	return getWebhooksWithLabelValues(*webhookRequestPanicCount, wt, obj)
}

// GetWebhookRequestRejectCount gets the request reject count for the given webhook type and target object.
func GetWebhookRequestRejectCount(wt utils.WebhookType, obj runtime.Object) int {
	return getWebhooksWithLabelValues(*webhookRequestRejectCount, wt, obj)
}

// GetWebhookRequestCount gets the request count for the given webhook type and target object.
func GetWebhookRequestCount(wt utils.WebhookType, obj runtime.Object) int {
	return getWebhooksWithLabelValues(*webhookRequestCount, wt, obj)
}

// GetWebhookRequestPassCount gets the request pass count for the given webhook type and target object.
func GetWebhookRequestPassCount(wt utils.WebhookType, obj runtime.Object) int {
	return getWebhooksWithLabelValues(*webhookRequestPassCount, wt, obj)
}

// Increment/update controller metric

// incControllersWithLabelValues increments the controller metric counter 'metric' by one.
func incControllersWithLabelValues(metric prometheus.CounterVec, target types.NamespacedName, gvk schema.GroupVersionKind) {
	metric.WithLabelValues(gvk.Group, gvk.Version, gvk.Kind, target.Namespace, target.Name).Inc()
}

// IncControllerReconcileCount increases reconcile count of the given object by 1.
func IncControllerReconcileCount(target types.NamespacedName, gvk schema.GroupVersionKind) {
	incControllersWithLabelValues(*controllerReconcileCount, target, gvk)
}

// IncControllerReconcilePanicCount increases reconcile panic count of the given object by 1.
func IncControllerReconcilePanicCount(target types.NamespacedName, gvk schema.GroupVersionKind) {
	incControllersWithLabelValues(*controllerReconcilePanicCount, target, gvk)
}

// IncControllerReconcileRequeueCount increases reconcile requeue count of the given object by 1.
func IncControllerReconcileRequeueCount(target types.NamespacedName, gvk schema.GroupVersionKind) {
	incControllersWithLabelValues(*controllerReconcileRequeueCount, target, gvk)
}

// UpdateControllerReconcileRequeueAfter updates reconcile requeue after histogram with the give time for the given object.
func UpdateControllerReconcileRequeueAfter(timeInMilliSecond int64, target types.NamespacedName, gvk schema.GroupVersionKind) {
	controllerReconcileRequeueAfter.WithLabelValues(gvk.Group, gvk.Version,
		gvk.Kind, target.Namespace, target.Name).Observe(float64(timeInMilliSecond))
}

// IncControllerReconcileRequeueErrorCount increases reconcile error count of the given object by 1.
func IncControllerReconcileRequeueErrorCount(target types.NamespacedName, gvk schema.GroupVersionKind) {
	incControllersWithLabelValues(*controllerReconcileRequeueErrorCount, target, gvk)
}

// UpdateControllerReconcileDuration updates reconcile duration histogram with the give time for the given object and webhook.
func UpdateControllerReconcileDuration(timeInMilliSeconds int64, gvk schema.GroupVersionKind, webhookName string, target types.NamespacedName) {
	controllerReconcileDuration.WithLabelValues(webhookName, gvk.Group, gvk.Version,
		gvk.Kind, target.Namespace, target.Name).Observe(float64(timeInMilliSeconds))
}

// ResetMetrics resets all metrics. Use for testing only.
func ResetMetrics() {
	_ = ReceivingMetricsFromOperator.Write(&prometheusclient.Metric{})
	controllerReconcileCount.Reset()
	controllerReconcilePanicCount.Reset()
	controllerReconcileRequeueAfter.Reset()
	controllerReconcileRequeueCount.Reset()
	controllerReconcileRequeueErrorCount.Reset()
	webhookRequestCount.Reset()
	webhookRequestPanicCount.Reset()
	webhookRequestPassCount.Reset()
	webhookRequestRejectCount.Reset()
}

// InitMetrics registers custom metrics with the global prometheus registry.
func InitMetrics() {
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
