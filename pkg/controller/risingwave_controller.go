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

	"github.com/go-logr/logr"
	"golang.org/x/time/rate"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/event"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/manager"
	m "github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

const (
	// Pre-defined actions. Import from manager package.
	RisingWaveAction_SyncMetaService                       = manager.RisingWaveAction_SyncMetaService
	RisingWaveAction_SyncMetaDeployments                   = manager.RisingWaveAction_SyncMetaDeployments
	RisingWaveAction_WaitBeforeMetaServiceIsAvailable      = manager.RisingWaveAction_WaitBeforeMetaServiceIsAvailable
	RisingWaveAction_WaitBeforeMetaDeploymentsReady        = manager.RisingWaveAction_WaitBeforeMetaDeploymentsReady
	RisingWaveAction_SyncFrontendService                   = manager.RisingWaveAction_SyncFrontendService
	RisingWaveAction_SyncFrontendDeployments               = manager.RisingWaveAction_SyncFrontendDeployments
	RisingWaveAction_WaitBeforeFrontendDeploymentsReady    = manager.RisingWaveAction_WaitBeforeFrontendDeploymentsReady
	RisingWaveAction_SyncComputeService                    = manager.RisingWaveAction_SyncComputeService
	RisingWaveAction_SyncComputeStatefulSets               = manager.RisingWaveAction_SyncComputeStatefulSets
	RisingWaveAction_WaitBeforeComputeStatefulSetsReady    = manager.RisingWaveAction_WaitBeforeComputeStatefulSetsReady
	RisingWaveAction_SyncCompactorService                  = manager.RisingWaveAction_SyncCompactorService
	RisingWaveAction_SyncCompactorDeployments              = manager.RisingWaveAction_SyncCompactorDeployments
	RisingWaveAction_WaitBeforeCompactorDeploymentsReady   = manager.RisingWaveAction_WaitBeforeCompactorDeploymentsReady
	RisingWaveAction_SyncConfigConfigMap                   = manager.RisingWaveAction_SyncConfigConfigMap
	RisingWaveAction_CollectRunningStatisticsAndSyncStatus = manager.RisingWaveAction_CollectRunningStatisticsAndSyncStatus
	RisingWaveAction_SyncServiceMonitor                    = manager.RisingWaveAction_SyncServiceMonitor

	// Actions defined in controller.
	RisingWaveAction_UpdateRisingWaveStatusViaClient    = "UpdateRisingWaveStatusViaClient"
	RisingWaveAction_BarrierFirstTimeObserved           = "BarrierFirstTimeObserved"
	RisingWaveAction_MarkConditionInitializingAsTrue    = "MarkConditionInitializingAsTrue"
	RisingWaveAction_MarkConditionRunningAsFalse        = "MarkConditionRunningAsFalse"
	RisingWaveAction_BarrierConditionInitializingIsTrue = "BarrierConditionInitializingIsTrue"
	RisingWaveAction_MarkConditionRunningAsTrue         = "MarkConditionRunningAsTrue"
	RisingWaveAction_RemoveConditionInitializing        = "RemoveConditionInitializing"
	RisingWaveAction_BarrierConditionRunningIsTrue      = "BarrierConditionRunningIsTrue"
	RisingWaveAction_BarrierConditionRunningIsFalse     = "BarrierConditionRunningIsFalse"
	RisingWaveAction_MarkConditionUpgradingAsTrue       = "MarkConditionUpgradingAsTrue"
	RisingWaveAction_BarrierConditionUpgradingIsTrue    = "BarrierConditionUpgradingIsTrue"
	RisingWaveAction_MarkConditionUpgradingAsFalse      = "MarkConditionUpgradingAsFalse"
	RisingWaveAction_BarrierObservedGenerationOutdated  = "BarrierObservedGenerationOutdated"
	RisingWaveAction_SyncObservedGeneration             = "SyncObservedGeneration"
	RisingWaveAction_BarrierPrometheusCRDsInstalled     = "BarrierPrometheusCRDsInstalled"
)

// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwaves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwavepodtemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwavepodtemplates/finalizers,verbs=update
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwaves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=risingwave.risingwavelabs.com,resources=risingwaves/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;delete;
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apiextensions.k8s.io,resources=customresourcedefinitions,verbs=get;list;watch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=servicemonitors,verbs=get;list;watch;create;delete;update;patch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

type RisingWaveController struct {
	Client            client.Client
	Recorder          record.EventRecorder
	ActionHookFactory func() ctrlkit.ActionHook
}

func (c *RisingWaveController) runWorkflow(ctx context.Context, workflow ctrlkit.Action) (result reconcile.Result, err error) {
	return ctrlkit.IgnoreExit(workflow.Run(ctx))
}

func (c *RisingWaveController) managerOpts(risingwaveMgr *object.RisingWaveManager, messageStore *event.MessageStore) []manager.RisingWaveControllerManagerOption {
	opts := make([]manager.RisingWaveControllerManagerOption, 0)
	chainedHooks := ctrlkit.ChainActionHooks(NewEventHook(c.Recorder, risingwaveMgr, messageStore))
	if c.ActionHookFactory != nil {
		chainedHooks.Add(c.ActionHookFactory())
	}
	opts = append(opts, manager.RisingWaveControllerManager_WithActionHook(chainedHooks))
	return opts
}

func (c *RisingWaveController) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	m.TestMetrics.Inc()
	m.ControllerReconcileCount.Inc()

	logger := log.FromContext(ctx)

	// Get the risingwave object.
	var risingwave risingwavev1alpha1.RisingWave
	if err := c.Client.Get(ctx, request.NamespacedName, &risingwave); err != nil {
		if apierrors.IsNotFound(err) {
			logger.V(1).Info("Not found, abort")
			return ctrlkit.NoRequeue()
		} else {
			logger.Error(err, "Failed to get risingwave")
			return ctrlkit.RequeueIfErrorAndWrap("unable to get risingwave", err)
		}
	}

	logger = logger.WithValues("generation", risingwave.Generation)

	// Pause and skip the reconciliation if the annotation is found.
	if _, ok := risingwave.Annotations[consts.AnnotationPauseReconcile]; ok {
		logger.Info("Found annotation " + consts.AnnotationPauseReconcile + ", pause reconciliation...")
		return ctrlkit.NoRequeue()
	}

	// Abort if deleted.
	if utils.IsDeleted(&risingwave) {
		logger.Info("Deleted, abort")
		return ctrlkit.NoRequeue()
	}

	risingwaveManager := object.NewRisingWaveManager(c.Client, risingwave.DeepCopy())
	eventMessageStore := event.NewMessageStore()

	mgr := manager.NewRisingWaveControllerManager(
		manager.NewRisingWaveControllerManagerState(c.Client, risingwave.DeepCopy()),
		manager.NewRisingWaveControllerManagerImpl(c.Client, risingwaveManager, eventMessageStore),
		logger,
		c.managerOpts(risingwaveManager, eventMessageStore)...,
	)

	// Build a workflow and run.
	workflow := c.reactiveWorkflow(risingwaveManager, &mgr)

	updateRisingWaveStatus := mgr.NewAction(RisingWaveAction_UpdateRisingWaveStatusViaClient, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		err := risingwaveManager.UpdateRemoteRisingWaveStatus(ctx)
		switch {
		case apierrors.IsNotFound(err):
			logger.Info("Object not found, skip")
			return ctrlkit.NoRequeue()
		case apierrors.IsConflict(err):
			logger.Info("Conflict detected while updating status, retry...")
			// Requeue after 10ms to give the cache time to sync.
			return ctrlkit.RequeueAfter(10 * time.Millisecond)
		default:
			return ctrlkit.RequeueIfErrorAndWrap("unable to update status", err)
		}
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
			Type:   risingwavev1alpha1.RisingWaveConditionInitializing,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	markConditionRunningAsFalse := mgr.NewAction(RisingWaveAction_MarkConditionRunningAsFalse, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionFalse,
		})
		return ctrlkit.Continue()
	})
	conditionInitializingIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionInitializingIsTrue, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.RisingWaveConditionInitializing)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionRunningAsTrue := mgr.NewAction(RisingWaveAction_MarkConditionRunningAsTrue, func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	removeConditionInitializing := mgr.NewAction(RisingWaveAction_RemoveConditionInitializing, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.RemoveCondition(risingwavev1alpha1.RisingWaveConditionInitializing)
		return ctrlkit.Continue()
	})

	conditionRunningIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionRunningIsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.RisingWaveConditionRunning)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	conditionRunningIsFalseBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionRunningIsFalse, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.RisingWaveConditionRunning)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionFalse)
	})
	markConditionUpgradingAsTrue := mgr.NewAction(RisingWaveAction_MarkConditionUpgradingAsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionUpgrading,
			Status: metav1.ConditionTrue,
		})
		return ctrlkit.Continue()
	})
	conditionUpgradingIsTrueBarrier := mgr.NewAction(RisingWaveAction_BarrierConditionUpgradingIsTrue, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		condition := risingwaveManger.GetCondition(risingwavev1alpha1.RisingWaveConditionUpgrading)
		return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
	})
	markConditionUpgradingAsFalse := mgr.NewAction(RisingWaveAction_MarkConditionUpgradingAsFalse, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionUpgrading,
			Status: metav1.ConditionFalse,
		})
		return ctrlkit.Continue()
	})
	syncConfigs := mgr.SyncConfigConfigMap()
	syncMetaComponent := ctrlkit.ParallelJoin(mgr.SyncMetaService(), mgr.SyncMetaDeployments())
	metaComponentReadyBarrier := ctrlkit.Sequential(
		mgr.WaitBeforeMetaDeploymentsReady(),
		ctrlkit.Timeout(time.Second, mgr.WaitBeforeMetaServiceIsAvailable()),
	)
	prometheusCRDsInstalledBarrier := mgr.NewAction(RisingWaveAction_BarrierPrometheusCRDsInstalled, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		crd, err := utils.GetCustomResourceDefinition(c.Client, ctx, metav1.GroupKind{
			Group: "monitoring.coreos.com",
			Kind:  "ServiceMonitor",
		})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return ctrlkit.Exit()
			}
			m.ControllerReconcileCount.Inc()
			return ctrlkit.RequeueIfErrorAndWrap("unable to find CRD for service monitor", err)
		}
		return ctrlkit.ExitIf(!utils.IsVersionServingInCustomResourceDefinition(crd, "v1"))
	})
	syncServiceMonitorIfPossible := ctrlkit.Sequential(
		prometheusCRDsInstalledBarrier,
		mgr.SyncServiceMonitor(),
	)
	syncOtherComponents := ctrlkit.ParallelJoin(
		ctrlkit.Sequential(
			mgr.SyncComputeService(),
			mgr.SyncComputeStatefulSets(),
		),
		ctrlkit.ParallelJoin(
			mgr.SyncCompactorService(),
			mgr.SyncCompactorDeployments(),
		),
		ctrlkit.ParallelJoin(
			mgr.SyncFrontendService(),
			mgr.SyncFrontendDeployments(),
		),
	)
	otherComponentsReadyBarrier := ctrlkit.Join(
		mgr.WaitBeforeFrontendDeploymentsReady(),
		mgr.WaitBeforeComputeStatefulSetsReady(),
		mgr.WaitBeforeCompactorDeploymentsReady(),
	)
	syncAllComponents := ctrlkit.ParallelJoin(syncConfigs, syncMetaComponent, syncOtherComponents)
	allComponentsReadyBarrier := ctrlkit.Join(metaComponentReadyBarrier, otherComponentsReadyBarrier)

	observedGenerationOutdatedBarrier := mgr.NewAction(RisingWaveAction_BarrierObservedGenerationOutdated, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(!risingwaveManger.IsObservedGenerationOutdated())
	})
	syncObservedGeneration := mgr.NewAction(RisingWaveAction_SyncObservedGeneration, func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.SyncObservedGeneration()
		return ctrlkit.Continue()
	})
	syncRunningStatus := mgr.CollectRunningStatisticsAndSyncStatus()

	syncAllAndWait := ctrlkit.Sequential(
		// Set .status.observedGeneration = .metadata.generation
		syncObservedGeneration,

		// Sync ConfigMap, and then all component groups, and wait before the components are ready.
		// If possible, also sync the service monitor.
		syncConfigs,
		syncAllComponents,
		allComponentsReadyBarrier,
	)
	sharedSyncAllAndWait := ctrlkit.Shared(syncAllAndWait)

	return ctrlkit.ParallelJoin(
		// => Initializing (Running=false)
		ctrlkit.Sequential(
			firstTimeObservedBarrier,
			markConditionInitializingAsTrue,
			markConditionRunningAsFalse,
		),

		// Initializing
		ctrlkit.Sequential(
			conditionInitializingIsTrueBarrier,

			sharedSyncAllAndWait,

			removeConditionInitializing,
		),

		// Running (false)
		ctrlkit.Sequential(
			conditionRunningIsFalseBarrier,

			sharedSyncAllAndWait,

			markConditionRunningAsTrue,
		),

		// Running maintenance, => Upgrading
		ctrlkit.Sequential(
			conditionRunningIsTrueBarrier,

			observedGenerationOutdatedBarrier,

			markConditionUpgradingAsTrue,
		),

		// Upgrading or Recovering (Running=false)
		ctrlkit.Sequential(
			conditionUpgradingIsTrueBarrier,

			sharedSyncAllAndWait,

			markConditionUpgradingAsFalse,
		),

		// Sync running status, such as storage status, component replicas and
		// if it's not running, turn it to Running=false.
		syncRunningStatus,

		// Always sync the service monitor if possible.
		syncServiceMonitorIfPossible,
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
		// Can't watch an optional CRD. It will cause a panic in manager.
		// So do not uncomment the following line.
		// Owns(&prometheusv1.ServiceMonitor{}).
		Complete(c)
}

func NewRisingWaveController(client client.Client, recorder record.EventRecorder) *RisingWaveController {
	return &RisingWaveController{
		Client:   client,
		Recorder: recorder,
	}
}
