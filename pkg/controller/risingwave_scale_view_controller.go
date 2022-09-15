/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"time"

	"github.com/samber/lo"
	"golang.org/x/time/rate"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/manager"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type RisingWaveScaleViewController struct {
	Client client.Client
}

// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwaves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwaves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwavescaleviews,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwavescaleviews/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (c *RisingWaveScaleViewController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)

	var scaleView risingwavev1alpha1.RisingWaveScaleView
	err := c.Client.Get(ctx, request.NamespacedName, &scaleView)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(1).Info("Not found, abort")
			return ctrlkit.NoRequeue()
		} else {
			logger.Error(err, "Failed to get risingwavescaleview")
			return ctrlkit.RequeueIfErrorAndWrap("unable to get risingwavescaleview", err)
		}
	}

	logger = logger.WithValues("generation", scaleView.Generation)

	// Build manager and workflow.
	mgr := manager.NewRisingWaveScaleViewControllerManager(
		manager.NewRisingWaveScaleViewControllerManagerState(c.Client, scaleView.DeepCopy()),
		manager.NewRisingWaveScaleViewControllerManagerImpl(c.Client, scaleView.DeepCopy()),
		logger,
	)

	isScaleViewDeleted := utils.IsDeleted(&scaleView)

	return ctrlkit.IgnoreExit(ctrlkit.OptimizeWorkflow(
		ctrlkit.IfElse(isScaleViewDeleted,
			mgr.HandleScaleViewFinalizer(),

			ctrlkit.OrderedJoin(
				ctrlkit.Sequential(
					ctrlkit.RetryInterval(2, 5*time.Millisecond, mgr.GrabOrUpdateScaleViewLock()),
					ctrlkit.Join(
						mgr.SyncGroupReplicasToRisingWave(),
						mgr.SyncGroupReplicasStatusFromRisingWave(),
					),
				),
				mgr.UpdateScaleViewStatus(),
			),
		),
	).Run(ctx))
}

func (c *RisingWaveScaleViewController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 64,
			RateLimiter: workqueue.NewMaxOfRateLimiter(
				// Exponential rate limiter, for immediate requeue (result.Requeue == true || err != nil).
				workqueue.NewItemExponentialFailureRateLimiter(5*time.Millisecond, 10*time.Second),
				// Bucket limiter of 10 qps, 100 bucket size.
				&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
			),
		}).
		For(&risingwavev1alpha1.RisingWaveScaleView{}).
		Watches(
			&source.Kind{Type: &risingwavev1alpha1.RisingWave{}},
			handler.EnqueueRequestsFromMapFunc(func(object client.Object) []reconcile.Request {
				obj := object.(*risingwavev1alpha1.RisingWave)
				return lo.Map(obj.Status.ScaleViews, func(t risingwavev1alpha1.RisingWaveScaleViewLock, _ int) reconcile.Request {
					return reconcile.Request{NamespacedName: types.NamespacedName{
						Namespace: obj.Namespace,
						Name:      t.Name,
					}}
				})
			}),
		).
		Complete(c)
}

func NewRisingWaveScaleViewController(client client.Client) *RisingWaveScaleViewController {
	return &RisingWaveScaleViewController{
		Client: client,
	}
}
