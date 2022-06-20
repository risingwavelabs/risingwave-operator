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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/manager"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/object"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/utils"
	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit"
)

type RisingWaveController struct {
	Client client.Client
	Logger logr.Logger
	DryRun bool
}

func (c *RisingWaveController) runWorkflow(ctx context.Context, workflow ctrlkit.ReconcileAction) (result reconcile.Result, err error) {
	if c.DryRun {
		ctrlkit.DryRun(workflow)
		return ctrlkit.NoRequeue()
	} else {
		return workflow.Run(ctx)
	}
}

func (c *RisingWaveController) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	logger := c.Logger.WithValues("risingwave", request)

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

	// Build a workflow and run.
	mgr := manager.NewRisingWaveControllerManager(
		manager.NewRisingWaveControllerManagerState(c.Client, risingwave.DeepCopy()),
		manager.NewRisingWaveControllerManagerImpl(c.Client, risingwaveManager),
		logger,
	)

	// Defer the status update.
	defer func() {
		if _, err1 := c.runWorkflow(ctx, mgr.UpdateRisingWaveStatus()); err1 != nil {
			// Overwrite err if it's nil.
			if err == nil {
				err = err1
			}
		}
	}()

	workflow := ctrlkit.OptimizeWorkflow(c.reactiveWorkflow(risingwaveManager, &mgr))

	logger.Info("Describe workflow", "workflow", workflow.Description())

	return c.runWorkflow(ctx, workflow)
}

func (c *RisingWaveController) reactiveWorkflow(risingwaveManger *object.RisingWaveManager, mgr *manager.RisingWaveControllerManager) ctrlkit.ReconcileAction {
	syncMetaComponent := ctrlkit.Join(ctrlkit.Join(mgr.SyncMetaService(), mgr.SyncMetaDeployment()))
	metaComponentReadyBarrier := ctrlkit.Join(mgr.WaitBeforeMetaDeploymentReady(), mgr.WaitBeforeMetaServiceIsAvailable())
	syncOtherComponents := ctrlkit.Join(
		ctrlkit.Join(mgr.SyncComputeSerivce(), mgr.SyncComputeDeployment()),
		ctrlkit.Join(mgr.SyncCompactorService(), mgr.SyncCompactorDeployment()),
		ctrlkit.Join(mgr.SyncFrontendService(), mgr.SyncFrontendDeployment()),
	)
	otherComponentsReadyBarrier := ctrlkit.Join(
		mgr.WaitBeforeFrontendDeploymentReady(),
		mgr.WaitBeforeComputeDeploymentReady(),
		mgr.WaitBeforeCompactorDeploymentReady(),
	)
	syncAllComponents := ctrlkit.Join(syncMetaComponent, syncOtherComponents)
	allComponentsReadyBarrier := ctrlkit.Join(metaComponentReadyBarrier, otherComponentsReadyBarrier)

	observedGenerationOutdatedBarrier := mgr.WrapAction("ObservedGenerationOutdatedBarrier", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		return ctrlkit.ExitIf(!risingwaveManger.IsObservedGenerationOutdated())
	})
	syncObservedGeneration := mgr.WrapAction("SyncObservedGeneration", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
		risingwaveManger.SyncObservedGeneration()
		return ctrlkit.Continue()
	})

	return ctrlkit.Join(
		// => Initializing
		ctrlkit.Sequential(
			mgr.WrapAction("BarrierFirstTimeObserved", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
				return ctrlkit.ExitIf(len(risingwaveManger.RisingWave().Status.Conditions) != 0)
			}),
			mgr.WrapAction("MarkConditionInitializingAsTrue", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
				risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
					Type:   risingwavev1alpha1.Initializing,
					Status: metav1.ConditionTrue,
				})
				return ctrlkit.Continue()
			}),
		),

		// Initializing => Running
		ctrlkit.Sequential(
			mgr.WrapAction("BarrierConditionInitializingIsTrue", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
				condition := risingwaveManger.GetCondition(risingwavev1alpha1.Initializing)
				return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
			}),

			// Sync observed generation.
			syncObservedGeneration,

			// Sync component for meta, and wait for the ready barrier.
			syncMetaComponent, metaComponentReadyBarrier,

			// Sync other components, and wait for the ready barrier.
			syncOtherComponents, otherComponentsReadyBarrier,

			mgr.WrapAction("MarkConditionRunningAsTrueAndInitializingToFalse", func(ctx context.Context, logger logr.Logger) (ctrl.Result, error) {
				risingwaveManger.RemoveCondition(risingwavev1alpha1.Initializing)
				risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
					Type:   risingwavev1alpha1.Initializing,
					Status: metav1.ConditionTrue,
				})
				return ctrlkit.Continue()
			}),
		),

		// Running maintenance, => Upgrading
		ctrlkit.Sequential(
			mgr.WrapAction("BarrierConditionRunningIsTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
				condition := risingwaveManger.GetCondition(risingwavev1alpha1.Running)
				return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
			}),

			ctrlkit.Join(
				// Branch, upgrade detection.
				ctrlkit.Sequential(
					observedGenerationOutdatedBarrier,

					syncObservedGeneration,

					mgr.WrapAction("MarkConditionUpgradingToTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
						risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
							Type:   risingwavev1alpha1.Upgrading,
							Status: metav1.ConditionTrue,
						})
						return ctrlkit.Continue()
					}),
				),

				// Branch, running maintenance.
				ctrlkit.Sequential(
				// TODO, nothing to do now.
				),
			),
		),

		// Upgrading
		ctrlkit.Sequential(
			mgr.WrapAction("BarrierConditionUpgradingIsTrue", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
				condition := risingwaveManger.GetCondition(risingwavev1alpha1.Upgrading)
				return ctrlkit.ExitIf(condition == nil || condition.Status != metav1.ConditionTrue)
			}),

			// Sync observed generation.
			syncObservedGeneration,

			// Sync all components, and wait for the ready barrier.
			syncAllComponents, allComponentsReadyBarrier,

			mgr.WrapAction("MarkConditionUpgradingToFalse", func(ctx context.Context, l logr.Logger) (ctrl.Result, error) {
				risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
					Type:   risingwavev1alpha1.Upgrading,
					Status: metav1.ConditionFalse,
				})
				return ctrlkit.Continue()
			}),
		),
	)
}

func (c *RisingWaveController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&risingwavev1alpha1.RisingWave{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(c)
}
