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
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	DidMutate = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_mutated_request_total",
			Help: "Incremented if mutating webhooks mutates at least one attribute",
		},
	)
	TestMetrics = prometheus.NewCounter( // TODO: remove metric?
		prometheus.CounterOpts{
			Name: "a_test_metric",
			Help: "test metric only",
		},
	)
	WebhookRequestCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_request_count",
			Help: "Total number of validating and mutating webhook calls",
		},
	)
	WebhookRequestPassCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_request_pass_count",
			Help: "Total number of accepted validating and mutating webhook calls",
		},
	)
	WebhookRequestRejectCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_request_reject_count",
			Help: "Total number of rejected validating and mutating webhook calls",
		},
	)
	WebhookRequestPanicCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "webhook_request_panic_count",
			Help: "Total number of panics during validating and mutating webhook calls",
		},
	)
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
	ControllerReconcileRequeueAfter = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_requeue_after",
			Help: "Duration of requeue in ms",
		},
	)
	ControllerReconcileDuration = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_duration",
			Help: "Duration of requeue in ms",
		},
	)
	ControllerReconcilePanicCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "controller_reconcile_panic_count",
			Help: "Total number of reconcile panics",
		},
	)
)

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(WebhookRequestPanicCount, WebhookRequestRejectCount, WebhookRequestPassCount, WebhookRequestCount, TestMetrics, ControllerReconcileCount, ControllerReconcileRequeueCount, ControllerReconcileRequeueErrorCount, ControllerReconcileRequeueAfter, ControllerReconcileDuration)
}
