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

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// use constants to avoid string literals when using the counters.
const (
	mutatingWebhook   = "mutate"
	validatingWebhook = "validate"
)

// TODO: Where should we move the metric implementation to
// TODO: Do I need additional tests for this?
// TODO: Make other metrics NewCounterVec, too.

var (
	ReceivingMetricsFromOperator = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receiving_metrics_from_operator",
			Help: "Value is 1, if you are able to receive custom metrics from the operator",
		},
	)

	// Webhook vectors have the following attributes:
	// type: The value should be mutating or validating
	// group: The target resource group of the webhook, e.g., risingwave.risingwavelabs.com
	// version: The target API version, e.g., v1alpha1
	// kind: The target API kind, e.g., risingwave, risingwavepodtemplate
	// namespace: The namespace of the object, e.g., default
	// name: The name of the object
	// TODO: implement verb
	// verb: The verb (action) on the object which triggers the webhook, the value should be one of "create", "update", and "delete".
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

	// Controller vectors have...
	ControllerReconcileCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_count",
			Help: "Total number of reconciles",
		},
	)
	ControllerReconcileRequeueCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_requeue_count",
			Help: "Total number of requeue",
		},
	)
	ControllerReconcileRequeueErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_error_count",
			Help: "Total number of requeue errors",
		},
	)
	ControllerReconcileRequeueAfter = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "controller_reconcile_requeue_after",
			Help: "Delay of last delayed requeue in ms",
		},
	)

	// How is controller_reconcile_duration different then ControllerReconcileRequeueAfter?

	ControllerReconcilePanicCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_panic_count",
			Help: "Total number of reconcile panics",
		},
	)
)

func incWebhooksWithLabelValues(metric prometheus.CounterVec, isValidating bool, obj runtime.Object) {
	type_ := mutatingWebhook
	if isValidating {
		type_ = validatingWebhook
	}
	gvk := obj.GetObjectKind().GroupVersionKind()
	rw, ok := obj.(*risingwavev1alpha1.RisingWave)
	if ok {
		metric.WithLabelValues(type_, gvk.Group, gvk.Version, gvk.Kind, rw.Name, rw.Namespace).Inc()
		return
	}
	rw2, _ := obj.(*risingwavev1alpha1.RisingWavePodTemplate)
	metric.WithLabelValues(type_, gvk.Group, gvk.Version, gvk.Kind, rw2.Namespace, rw2.Name).Inc()
}

// IncWebhookRequestCount increments the WebhookRequestCount
// isValidating: If true then increments validating webhook, else mutating webhook
// TODO:
// verb: The verb (action) on the object which triggers the webhook, the value should be one of "create", "update", and "delete".
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

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(ControllerReconcileCount)
	metrics.Registry.MustRegister(ControllerReconcileRequeueAfter)
	metrics.Registry.MustRegister(ControllerReconcileRequeueCount)
	metrics.Registry.MustRegister(ControllerReconcileRequeueErrorCount)
	metrics.Registry.MustRegister(ReceivingMetricsFromOperator)
	metrics.Registry.MustRegister(webhookRequestCount)
	metrics.Registry.MustRegister(webhookRequestPanicCount)
	metrics.Registry.MustRegister(webhookRequestPassCount)
	metrics.Registry.MustRegister(webhookRequestRejectCount)
}
