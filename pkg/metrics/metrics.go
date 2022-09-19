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
	TestMetrics = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "a_test_metric",
			Help: "test metric only",
		},
	)
)

func InitMetrics() {
	// Register custom metrics with the global prometheus registry
	metrics.Registry.MustRegister(TestMetrics, Reconciles, RequeueCount, DidMutate, ValidatingWebhookCalls, ValidatingWebhookErr)
}
