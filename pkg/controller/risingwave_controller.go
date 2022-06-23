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

package risingwave

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit"
	"github.com/singularity-data/risingwave-operator/pkg/manager"
	"github.com/singularity-data/risingwave-operator/pkg/object"
	"github.com/singularity-data/risingwave-operator/pkg/utils"
)

// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=risingwave.singularity-data.com,resources=risingwaves/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;delete

type RisingWaveController struct {
	Client client.Client
	DryRun bool
}

func (c *RisingWaveController) runWorkflow(ctx context.Context, workflow ctrlkit.ReconcileAction) (result reconcile.Result, err error) {
	if c.DryRun {
		ctrlkit.DryRun(workflow)
		return ctrlkit.NoRequeue()
	} else {
		return ctrlkit.IgnoreExit(workflow.Run(ctx))
	}
}

func (c *RisingWaveController) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	logger := log.FromContext(ctx)

	// Get the risingwave object.
	var risingwave risingwavev1alpha1.RisingWave
	if err := c.Client.Get(ctx, request.NamespacedName, &risingwave); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Not found, abort")
			return ctrlkit.NoRequeue()
		} else {
			logger.Error(err, "Failed to get RisingWave")
			return ctrlkit.RequeueIfErrorAndWrap("unable to get risingwave", err)
		}
	}

	// Abort if deleted.
	if utils.IsDeleted(&risingwave) {
		logger.Info("Deleted, abort")
		return ctrlkit.NoRequeue()
	}

	risingwaveManager := object.NewRisingWaveManager(c.Client, risingwave.DeepCopy())

	mgr := manager.NewRisingWaveControllerManager(
		manager.NewRisingWaveControllerManagerState(c.Client, risingwave.DeepCopy()),
		manager.NewRisingWaveControllerManagerImpl(c.Client, risingwaveManager),
		logger,
	)

	// Defer the status update.
	defer func() {
		if err1 := risingwaveManager.UpdateRemoteRisingWaveStatus(ctx); err1 != nil {
			// Overwrite err if it's nil.
			if err == nil {
				err = fmt.Errorf("unable to update status: %w", err1)
			}
		}
	}()

	// Build a workflow and run.
	workflow := ctrlkit.OptimizeWorkflow(c.reactiveWorkflow(risingwaveManager, &mgr))
	return c.runWorkflow(ctx, workflow)
}

func (c *RisingWaveController) reactiveWorkflow(risingwaveManger *object.RisingWaveManager, mgr *manager.RisingWaveControllerManager) ctrlkit.ReconcileAction {
	firstTimeObservedBarrier := mgr.WrapAction("BarrierFirstTimeObserved", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(len(risingwaveManger.RisingWave().Status.Conditions) != 0)
	})
	markConditionInitializingAsTrue := mgr.WrapAction("MarkConditionInitializingAsTrue", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Initializing,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	markConditionRunningAsFalse := mgr.WrapAction("MarkConditionInitializingAsTrue", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Running,
			Status: metav1.ConditionFalse,
		})
		return ctrlkit.Continue()
	})
	conditionInitializingIsTrueBarrier := mgr.WrapAction("BarrierConditionInitializingIsTrue", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Initializing)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionRunningAsTrueAndRemoveConditionInitializing := mgr.WrapAction("MarkConditionRunningAsTrueAndRemoveConditionInitializing", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.RemoveCondition(risingwavev1alpha1.Initializing)
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Running,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	conditionRunningIsTrueBarrier := mgr.WrapAction("BarrierConditionRunningIsTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Running)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionUpgradingAsTrue := mgr.WrapAction("MarkConditionUpgradingAsTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Upgrading,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	conditionUpgradingIsTrueBarrier := mgr.WrapAction("BarrierConditionUpgradingIsTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Upgrading)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionUpgradingAsFalse := mgr.WrapAction("MarkConditionUpgradingAsFalse", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Upgrading,
			Status: metav1.ConditionFalse,
		})
		return ctrlkit.Continue()
	})
	syncConfigs := mgr.SyncConfigConfigMap()
	syncMetaComponent := ctrlkit.JoinInParallel(mgr.SyncMetaService(), mgr.SyncMetaDeployment())
	metaComponentReadyBarrier := ctrlkit.Sequential(
		mgr.WaitBeforeMetaDeploymentReady(),
		ctrlkit.Timeout(time.Second, mgr.WaitBeforeMetaServiceIsAvailable()),
	)
	syncOtherComponents := ctrlkit.JoinInParallel(
		ctrlkit.JoinInParallel(
			mgr.SyncComputeService(),
			mgr.SyncComputeStatefulSet(),
		),
		ctrlkit.JoinInParallel(
			mgr.SyncCompactorService(),
			mgr.SyncCompactorDeployment(),
		),

		ctrlkit.JoinInParallel(
			mgr.SyncFrontendService(),
			mgr.SyncFrontendDeployment(),
		),
	)
	otherComponentsReadyBarrier := ctrlkit.Join(
		mgr.WaitBeforeFrontendDeploymentReady(),
		mgr.WaitBeforeComputeStatefulSetReady(),
		mgr.WaitBeforeCompactorDeploymentReady(),
	)
	syncAllComponents := ctrlkit.JoinInParallel(syncConfigs, syncMetaComponent, syncOtherComponents)
	allComponentsReadyBarrier := ctrlkit.Join(metaComponentReadyBarrier, otherComponentsReadyBarrier)

	observedGenerationOutdatedBarrier := mgr.WrapAction("ObservedGenerationOutdatedBarrier", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(!risingwaveManger.IsObservedGenerationOutdated())
	})
	syncObservedGeneration := mgr.WrapAction("SyncObservedGeneration", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.SyncObservedGeneration()
		return ctrlkit.Continue()
	})
	syncStorageAndComponentStatus := mgr.CollectRunningStatisticsAndSyncStatus()

	return ctrlkit.JoinInParallel(
		// => Initializing (Running=false)
		ctrlkit.Sequential(
			firstTimeObservedBarrier,
			markConditionInitializingAsTrue,
			markConditionRunningAsFalse,
		),

		// Initializing => Running
		ctrlkit.Sequential(
			conditionInitializingIsTrueBarrier,

			ctrlkit.Sequential(
				// Sync observed generation.
				syncObservedGeneration,

				// Sync configs.
				syncConfigs,

				// Sync component for meta, and wait for the ready barrier.
				syncMetaComponent, metaComponentReadyBarrier,

				// Sync other components, and wait for the ready barrier.
				syncOtherComponents, otherComponentsReadyBarrier,

				markConditionRunningAsTrueAndRemoveConditionInitializing,
			),
		),

		// Running maintenance, => Upgrading
		ctrlkit.Sequential(
			conditionRunningIsTrueBarrier,

			ctrlkit.JoinInParallel(
				// Branch, upgrade detection.
				ctrlkit.Sequential(
					observedGenerationOutdatedBarrier,

					markConditionUpgradingAsTrue,
				),

				// Branch, running maintenance.
				ctrlkit.Sequential(
				// TODO, nothing to do now.
				),
			),
		),

		// Upgrading
		ctrlkit.Sequential(
			conditionUpgradingIsTrueBarrier,

			// Sync observed generation.
			syncObservedGeneration,

			// Sync all components, and wait for the ready barrier.
			syncAllComponents, allComponentsReadyBarrier,

			markConditionUpgradingAsFalse,
		),

		// Always run.
		ctrlkit.Join(
			// Sync the storage and component status each time we run.
			syncStorageAndComponentStatus,
		),
	)
}

func (c *RisingWaveController) SetupWithManager(mgr ctrl.Manager) error {
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
		For(&risingwavev1alpha1.RisingWave{}).
		Owns(&appsv1.Deployment{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(c)
}

func NewReconciler(client client.Client, _ *runtime.Scheme) *RisingWaveController {
	return &RisingWaveController{
		Client: client,
	}
}
