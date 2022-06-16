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

package manager

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/consts"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/factory"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/object"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/utils"
	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit"
)

type risingWaveControllerManagerImpl struct {
	ctrlkit.EmptyCrontollerManagerActionLifeCycleHook

	client            client.Client
	risingwaveManager *object.RisingWaveManager
	objectFactory     *factory.RisingWaveObjectFactory
}

func (mgr *risingWaveControllerManagerImpl) isObjectSynced(obj client.Object) bool {
	if obj == nil {
		return true
	}

	generationLabel := obj.GetLabels()[consts.LabelRisingWaveGeneration]
	// Ignore the parse error, as generation label should always be numbers.
	// And if not, it must be synced. So a default value of 0 on error is good enough.
	observedGeneration, _ := strconv.ParseInt(generationLabel, 10, 64)
	currentGeneration := mgr.risingwaveManager.RisingWave().Generation

	// Use larger than to avoid cases that we observed an old RisingWave object and
	// a newer object.
	return observedGeneration >= currentGeneration
}

func ensureTheSameObject(obj, newObj client.Object) client.Object {
	// Ensure that they are the same object.
	if obj.GetName() != newObj.GetName() || obj.GetNamespace() != newObj.GetNamespace() {
		panic(fmt.Sprintf("objects not the same: %s/%s vs. %s/%s",
			obj.GetNamespace(), obj.GetName(),
			newObj.GetNamespace(), newObj.GetName(),
		))
	}
	if reflect.TypeOf(obj) == reflect.TypeOf(newObj) {
		panic(fmt.Sprintf("object types' not equal: %T vs. %T", obj, newObj))
	}

	return newObj
}

func (mgr *risingWaveControllerManagerImpl) syncObject(ctx context.Context, obj client.Object, factory func() client.Object, logger logr.Logger) error {
	scheme := mgr.client.Scheme()
	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}

	if obj == nil {
		// Not found. Going to create one.
		newObj := ensureTheSameObject(obj, factory())
		logger.Info(fmt.Sprintf("Create an object of %s", gvk.Kind), "object", utils.GetNamespacedName(newObj))
		return mgr.client.Create(ctx, newObj)
	} else {
		// Found. Update/Sync if not synced.
		if !mgr.isObjectSynced(obj) {
			newObj := ensureTheSameObject(obj, factory())
			logger.Info(fmt.Sprintf("Update the object of %s", gvk.Kind), "object", utils.GetNamespacedName(newObj),
				"generation", mgr.risingwaveManager.RisingWave().Generation)
			return mgr.client.Update(ctx, newObj)
		}
		return nil
	}
}

// Helper function for compile time type assertion.
func syncObject[T client.Object](mgr *risingWaveControllerManagerImpl, ctx context.Context, obj T, factory func() T, logger logr.Logger) error {
	return mgr.syncObject(ctx, obj, func() client.Object {
		return factory()
	}, logger)
}

// SyncCompactorDeployment implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncCompactorDeployment(ctx context.Context, logger logr.Logger, compactorDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, compactorDeployment, mgr.objectFactory.NewCompactorDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor deployment", err)
}

// SyncCompactorService implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncCompactorService(ctx context.Context, logger logr.Logger, compactorService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, compactorService, mgr.objectFactory.NewCompactorService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor service", err)
}

// SyncComputeDeployment implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncComputeDeployment(ctx context.Context, logger logr.Logger, computeDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, computeDeployment, mgr.objectFactory.NewComputeDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute deployment", err)
}

// SyncComputeSerivce implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncComputeSerivce(ctx context.Context, logger logr.Logger, computeService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, computeService, mgr.objectFactory.NewComputeService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute service", err)
}

// SyncFrontendDeployment implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncFrontendDeployment(ctx context.Context, logger logr.Logger, frontendDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendDeployment, mgr.objectFactory.NewFrontendDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend deployment", err)
}

// SyncFrontendService implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncFrontendService(ctx context.Context, logger logr.Logger, frontendService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendService, mgr.objectFactory.NewFrontendService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend service", err)
}

// SyncMetaDeployment implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncMetaDeployment(ctx context.Context, logger logr.Logger, metaDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, metaDeployment, mgr.objectFactory.NewMetaDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta deployment", err)
}

// SyncMetaService implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncMetaService(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, metaService, mgr.objectFactory.NewMetaService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta service", err)
}

// SyncMinIODeployment implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncMinIODeployment(ctx context.Context, logger logr.Logger, minioDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, minioDeployment, mgr.objectFactory.NewMinIODeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync minio deployment", err)
}

// SyncMinIOService implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) SyncMinIOService(ctx context.Context, logger logr.Logger, minioService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, minioService, mgr.objectFactory.NewMinIOService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync minio service", err)
}

// WaitBeforeCompactorDeploymentReady implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorDeploymentReady(ctx context.Context, logger logr.Logger, compactorDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if utils.IsDeploymentRolledOut(compactorDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Compactor deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeComputeDeploymentReady implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeDeploymentReady(ctx context.Context, logger logr.Logger, computeDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if utils.IsDeploymentRolledOut(computeDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Compute deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeFrontendDeploymentReady implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendDeploymentReady(ctx context.Context, logger logr.Logger, frontendDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if utils.IsDeploymentRolledOut(frontendDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Frontend deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeMetaDeploymentReady implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaDeploymentReady(ctx context.Context, logger logr.Logger, metaDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if utils.IsDeploymentRolledOut(metaDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Meta deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeMetaServiceIsAvailable implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaServiceIsAvailable(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	if utils.IsServiceReady(metaService) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Meta service hasn't been ready")
		return ctrlkit.Exit()
	}
}

// WaitBeforeMinIODeploymentReady implements RisingWaveControllerManagerImpl
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMinIODeploymentReady(ctx context.Context, logger logr.Logger, minioDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if utils.IsDeploymentRolledOut(minioDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("MinIO deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

func NewRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager) RisingWaveControllerManagerImpl {
	return &risingWaveControllerManagerImpl{
		client:            client,
		risingwaveManager: risingwaveManager,
		objectFactory:     factory.NewRisingWaveObjectFactory(risingwaveManager.RisingWave(), client.Scheme()),
	}
}
