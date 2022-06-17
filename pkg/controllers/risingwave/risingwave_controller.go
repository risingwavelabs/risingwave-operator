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

	if !c.DryRun {
		defer func() {
			if _, err1 := c.runWorkflow(ctx, mgr.UpdateRisingWaveStatus()); err1 != nil {
				// Overwrite err if it's nil.
				if err == nil {
					err = err1
				}
			}
		}()
	}

	var workflow ctrlkit.ReconcileAction
	if len(risingwave.Status.Conditions) == 0 {
		workflow = ctrlkit.WrapAction("MarkConditionInitializing", func(ctx context.Context) (ctrl.Result, error) {
			risingwaveManager.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
				Type:   risingwavev1alpha1.Initializing,
				Status: metav1.ConditionTrue,
			})
			return ctrlkit.RequeueImmediately()
		})
	} else {
		actions := make([]ctrlkit.ReconcileAction, 0)
		for _, cond := range risingwave.Status.Conditions {
			if cond.Status == metav1.ConditionTrue {
				actions = append(actions, c.buildWorkflow(cond.Type, &risingwave, risingwaveManager, &mgr))
			}
		}
		if len(actions) == 0 {
			workflow = ctrlkit.Nop
		} else {
			workflow = ctrlkit.Join(actions...)
		}
	}

	logger.Info("Describe workflow", "workflow", workflow.Description())

	return c.runWorkflow(ctx, workflow)
}

func (c *RisingWaveController) buildWorkflow(condition risingwavev1alpha1.RisingWaveType, risingwave *risingwavev1alpha1.RisingWave,
	risingwaveManger *object.RisingWaveManager, mgr *manager.RisingWaveControllerManager) ctrlkit.ReconcileAction {
	switch condition {
	case risingwavev1alpha1.Initializing:
		return ctrlkit.Sequential(
			// Sync the service and deployment for meta components.
			ctrlkit.Join(mgr.SyncMetaService(), mgr.SyncMetaDeployment()),

			// Wait before meta deployment's ready.
			ctrlkit.Timeout(5*time.Second, mgr.WaitBeforeMetaDeploymentReady()),

			// Sync object storage and wait.
			ctrlkit.If(risingwave.Spec.ObjectStorage.MinIO != nil,
				ctrlkit.Sequential(
					ctrlkit.Join(mgr.SyncMinIOService(), mgr.SyncMinIODeployment()),
					mgr.WaitBeforeMinIODeploymentReady(),
				),
			),

			// Sync the other components.
			ctrlkit.Join(
				ctrlkit.Join(mgr.SyncFrontendService(), mgr.SyncFrontendDeployment()),
				ctrlkit.Join(mgr.SyncComputeSerivce(), mgr.SyncComputeDeployment()),
				ctrlkit.Join(mgr.SyncCompactorService(), mgr.SyncCompactorDeployment()),
			),

			// Wait before these components' ready.
			ctrlkit.Join(
				mgr.WaitBeforeFrontendDeploymentReady(),
				mgr.WaitBeforeComputeDeploymentReady(),
				mgr.WaitBeforeCompactorDeploymentReady(),
			),

			// Update the condition to Running.
			ctrlkit.WrapAction("MarkConditionRunning", func(ctx context.Context) (ctrl.Result, error) {
				risingwaveManger.RemoveCondition(risingwavev1alpha1.Initializing)
				risingwaveManger.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
					Type:   risingwavev1alpha1.Running,
					Status: metav1.ConditionTrue,
				})
				return ctrlkit.RequeueImmediately()
			}),
		)
	case risingwavev1alpha1.Running:
		return ctrlkit.Nop
	case risingwavev1alpha1.Upgrading:
		return ctrlkit.Sequential(
			// Sync all these components.
			ctrlkit.Join(
				mgr.SyncMetaService(), mgr.SyncMetaDeployment(),
				mgr.SyncFrontendService(), mgr.SyncFrontendDeployment(),
				mgr.SyncComputeSerivce(), mgr.SyncComputeDeployment(),
				mgr.SyncCompactorService(), mgr.SyncCompactorDeployment(),
				ctrlkit.If(risingwave.Spec.ObjectStorage.MinIO != nil,
					ctrlkit.Join(mgr.SyncMinIOService(), mgr.SyncMinIODeployment()),
				),
			),
			// Wait before these components' ready.
			ctrlkit.Join(
				mgr.WaitBeforeMetaDeploymentReady(),
				mgr.WaitBeforeFrontendDeploymentReady(),
				mgr.WaitBeforeComputeDeploymentReady(),
				mgr.WaitBeforeCompactorDeploymentReady(),
				ctrlkit.If(risingwave.Spec.ObjectStorage.MinIO != nil,
					mgr.WaitBeforeMinIODeploymentReady(),
				),
			),
			ctrlkit.WrapAction("RemoveConditionUpgrading", func(ctx context.Context) (ctrl.Result, error) {
				risingwaveManger.RemoveCondition(risingwavev1alpha1.Upgrading)
				return ctrlkit.NoRequeue()
			}),
		)
	default:
		return ctrlkit.Nop
	}
}

func (c *RisingWaveController) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&risingwavev1alpha1.RisingWave{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(c)
}
