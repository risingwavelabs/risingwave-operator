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
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type controllerMetricsRecorder struct {
	name  string
	gvk   schema.GroupVersionKind
	inner reconcile.Reconciler
}

// Reconcile implements the Reconciler.
func (r *controllerMetricsRecorder) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	startTime := time.Now()
	r.beforeReconcile(request.NamespacedName)
	defer r.afterReconcile(ctx, request, &result, &err, startTime)

	return r.inner.Reconcile(ctx, request)
}

func (r *controllerMetricsRecorder) beforeReconcile(target types.NamespacedName) {
	IncControllerReconcileCount(target, r.gvk)
}

func (r *controllerMetricsRecorder) afterReconcile(ctx context.Context, request reconcile.Request,
	result *reconcile.Result, err *error, startTime time.Time) {
	namespace := request.NamespacedName

	if rec := recover(); rec != nil {
		IncControllerReconcilePanicCount(namespace, r.gvk)
		log.FromContext(ctx).Error(fmt.Errorf("%v", rec), "Panic in reconciliation run\n")
		*result, *err = reconcile.Result{}, nil
		return
	}

	if *err != nil {
		IncControllerReconcileRequeueErrorCount(namespace, r.gvk)
	} else if result.RequeueAfter > 0 {
		UpdateControllerReconcileRequeueAfter(result.RequeueAfter.Milliseconds(), namespace, r.gvk)
	} else if result.Requeue {
		IncControllerReconcileRequeueCount(namespace, r.gvk)
	}

	UpdateControllerReconcileDuration(time.Since(startTime).Milliseconds(), r.gvk, r.name, namespace)
}

// NewControllerMetricsRecorder returns a new ControllerMetricsRecorder.
func NewControllerMetricsRecorder(r reconcile.Reconciler, name string, gvk schema.GroupVersionKind) reconcile.Reconciler {
	return &controllerMetricsRecorder{
		name:  name,
		gvk:   gvk,
		inner: r,
	}
}
