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

const (
	// Pre-defined actions. Import from manager package.
	RisingWaveAction_SyncMetaService                       = manager.RisingWaveAction_SyncMetaService
	RisingWaveAction_SyncMetaDeployment                    = manager.RisingWaveAction_SyncMetaDeployment
	RisingWaveAction_WaitBeforeMetaServiceIsAvailable      = manager.RisingWaveAction_WaitBeforeMetaServiceIsAvailable
	RisingWaveAction_WaitBeforeMetaDeploymentReady         = manager.RisingWaveAction_WaitBeforeMetaDeploymentReady
	RisingWaveAction_SyncFrontendService                   = manager.RisingWaveAction_SyncFrontendService
	RisingWaveAction_SyncFrontendDeployment                = manager.RisingWaveAction_SyncFrontendDeployment
	RisingWaveAction_WaitBeforeFrontendDeploymentReady     = manager.RisingWaveAction_WaitBeforeFrontendDeploymentReady
	RisingWaveAction_SyncComputeService                    = manager.RisingWaveAction_SyncComputeService
	RisingWaveAction_SyncComputeStatefulSet                = manager.RisingWaveAction_SyncComputeStatefulSet
	RisingWaveAction_WaitBeforeComputeStatefulSetReady     = manager.RisingWaveAction_WaitBeforeComputeStatefulSetReady
	RisingWaveAction_SyncCompactorService                  = manager.RisingWaveAction_SyncCompactorService
	RisingWaveAction_SyncCompactorDeployment               = manager.RisingWaveAction_SyncCompactorDeployment
	RisingWaveAction_WaitBeforeCompactorDeploymentReady    = manager.RisingWaveAction_WaitBeforeCompactorDeploymentReady
	RisingWaveAction_SyncConfigConfigMap                   = manager.RisingWaveAction_SyncConfigConfigMap
	RisingWaveAction_CollectRunningStatisticsAndSyncStatus = manager.RisingWaveAction_CollectRunningStatisticsAndSyncStatus

	// Actions defined in controller.
	RisingWaveAction_UpdateRisingWaveStatusViaClient    = "UpdateRisingWaveStatusViaClient"
	RisingWaveAction_BarrierFirstTimeObserved           = "BarrierFirstTimeObserved"
	RisingWaveAction_MarkConditionInitializingAsTrue    = "MarkConditionInitializingAsTrue"
	RisingWaveAction_MarkConditionRunningAsFalse        = "MarkConditionRunningAsFalse"
	RisingWaveAction_BarrierConditionInitializingIsTrue = "BarrierConditionInitializingIsTrue"
	RisingWaveAction_MarkConditionRunningAsTrue         = "MarkConditionRunningAsTrue"
	RisingWaveAction_RemoveConditionInitializing        = "RemoveConditionInitializing"
	RisingWaveAction_BarrierConditionRunningIsTrue      = "BarrierConditionRunningIsTrue"
	RisingWaveAction_MarkConditionUpgradingAsTrue       = "MarkConditionUpgradingAsTrue"
	RisingWaveAction_BarrierConditionUpgradingIsTrue    = "BarrierConditionUpgradingIsTrue"
	RisingWaveAction_MarkConditionUpgradingAsFalse      = "MarkConditionUpgradingAsFalse"
	RisingWaveAction_BarrierObservedGenerationOutdated  = "BarrierObservedGenerationOutdated"
	RisingWaveAction_SyncObservedGeneration             = "SyncObservedGeneration"
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
	Client            client.Client
	ActionHookFactory func() ctrlkit.ActionHook
	DryRun            bool
}

func (c *RisingWaveController) runWorkflow(ctx context.Context, workflow ctrlkit.Action) (result reconcile.Result, err error) {
	if c.DryRun {
		ctrlkit.DryRun(workflow)
		return ctrlkit.NoRequeue()
	} else {
		return ctrlkit.IgnoreExit(workflow.Run(ctx))
	}
}

func (c *RisingWaveController) managerOpts() []manager.RisingWaveControllerManagerOption {
	opts := make([]manager.RisingWaveControllerManagerOption, 0)
	if c.ActionHookFactory != nil {
		opts = append(opts, manager.RisingWaveControllerManager_WithActionHook(c.ActionHookFactory()))
	}
	return opts
}

func (c *RisingWaveController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
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
		c.managerOpts()...,
	)

	// Build a workflow and run.
	workflow := c.reactiveWorkflow(risingwaveManager, &mgr)

	updateRisingWaveStatus := mgr.NewAction(RisingWaveAction_UpdateRisingWaveStatusViaClient, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		err := risingwaveManager.UpdateRemoteRisingWaveStatus(ctx)
		return ctrlkit.RequeueIfErrorAndWrap("unable to update status", err)
	})

	return c.runWorkflow(ctx, ctrlkit.OptimizeWorkflow(ctrlkit.SequentialJoin(
		workflow,               // Run workflow first,
		updateRisingWaveStatus, // then update the status, and join the result.
	)))
}

func (c *RisingWaveController) reactiveWorkflow(risingwaveManger *object.RisingWaveManager, mgr *manager.RisingWaveControllerManager) ctrlkit.Action {
	firstTimeObservedBarrier := mgr.NewAction(RisingWaveAction_BarrierFirstTimeObserved, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(len(risingwaveManger.RisingWave().Status.Conditions) != 0)
	})
	markConditionInitializingAsTrue := mgr.NewAction(RisingWaveAction_MarkConditionInitializingAsTrue, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Initializing,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	markConditionRunningAsFalse := mgr.NewAction(RisingWaveAction_MarkConditionRunningAsFalse, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Running,
			Status: metav1.ConditionFalse,
		})
		return ctrlkit.Continue()
	})
	conditionInitializingIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionInitializingIsTrue, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Initializing)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionRunningAsTrue := mgr.NewAction(RisingWaveAction_MarkConditionRunningAsTrue, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Running,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	removeConditionInitializing := mgr.NewAction(RisingWaveAction_RemoveConditionInitializing, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.RemoveCondition(risingwavev1alpha1.Initializing)
		return ctrlkit.Continue()
	})
	markConditionRunningAsTrueAndRemoveConditionInitializing := ctrlkit.Join(markConditionRunningAsTrue, removeConditionInitializing)

	conditionRunningIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionRunningIsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Running)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionUpgradingAsTrue := mgr.NewAction(RisingWaveAction_MarkConditionUpgradingAsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.Upgrading,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	conditionUpgradingIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionUpgradingIsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.Upgrading)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionUpgradingAsFalse := mgr.NewAction(RisingWaveAction_MarkConditionUpgradingAsFalse, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
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

	observedGenerationOutdatedBarrier := mgr.NewAction(RisingWaveAction_BarrierObservedGenerationOutdated, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(!risingwaveManger.IsObservedGenerationOutdated())
	})
	syncObservedGeneration := mgr.NewAction(RisingWaveAction_SyncObservedGeneration, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
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

				// Sync all components, and wait for the ready barrier.
				syncAllComponents, allComponentsReadyBarrier,

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
