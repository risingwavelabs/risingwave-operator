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
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/consts"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/factory"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/object"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/utils"
	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit"
)

type risingWaveControllerManagerImpl struct {
	ctrlkit.EmptyControllerManagerActionLifeCycleHook

	client            client.Client
	risingwaveManager *object.RisingWaveManager
	objectFactory     *factory.RisingWaveObjectFactory
}

func (mgr *risingWaveControllerManagerImpl) isObjectSynced(obj client.Object) bool {
	if isObjectNil(obj) {
		return false
	}

	generationLabel := obj.GetLabels()[consts.LabelRisingWaveGeneration]

	// Do not sync, so return true here.
	if consts.NoSync == generationLabel {
		return true
	}

	// Ignore the parse error, as generation label should always be numbers.
	// And if not, it must be synced. So a default value of 0 on error is good enough.
	observedGeneration, _ := strconv.ParseInt(generationLabel, 10, 64)
	currentGeneration := mgr.risingwaveManager.RisingWave().Generation

	// Use larger than to avoid cases that we observed an old RisingWave object and
	// a newer object.
	return observedGeneration >= currentGeneration
}

func ensureTheSameObject(obj, newObj client.Object) client.Object {
	// Ensure that they are the same object in Kubernetes.
	if !isObjectNil(obj) {
		if obj.GetName() != newObj.GetName() || obj.GetNamespace() != newObj.GetNamespace() {
			panic(fmt.Sprintf("objects not the same: %s/%s vs. %s/%s",
				obj.GetNamespace(), obj.GetName(),
				newObj.GetNamespace(), newObj.GetName(),
			))
		}
	}

	objType, newObjType := reflect.TypeOf(obj).Elem(), reflect.TypeOf(newObj).Elem()
	if objType != newObjType {
		panic(fmt.Sprintf("object types' not equal: %T vs. %T", obj, newObj))
	}

	return newObj
}

func isObjectNil(obj client.Object) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	return v.IsNil()
}

func (mgr *risingWaveControllerManagerImpl) syncObject(ctx context.Context, obj client.Object, factory func() client.Object, logger logr.Logger) error {
	scheme := mgr.client.Scheme()

	if isObjectNil(obj) {
		// Not found. Going to create one.
		newObj := ensureTheSameObject(obj, factory())

		gvk, err := apiutil.GVKForObject(newObj, scheme)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("Create an object of %s", gvk.Kind), "object", utils.GetNamespacedName(newObj))
		return mgr.client.Create(ctx, newObj)
	} else {
		gvk, err := apiutil.GVKForObject(obj, scheme)
		if err != nil {
			return err
		}

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

// SyncCompactorDeployment implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorDeployment(ctx context.Context, logger logr.Logger, compactorDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, compactorDeployment, mgr.objectFactory.NewCompactorDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor deployment", err)
}

// SyncCompactorService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorService(ctx context.Context, logger logr.Logger, compactorService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, compactorService, mgr.objectFactory.NewCompactorService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor service", err)
}

// SyncComputeStatefulSet implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeStatefulSet(ctx context.Context, logger logr.Logger, computeStatefulSet *appsv1.StatefulSet) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, computeStatefulSet, mgr.objectFactory.NewComputeStatefulSet, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute statefulset", err)
}

// SyncComputeSerivce implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeSerivce(ctx context.Context, logger logr.Logger, computeService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, computeService, mgr.objectFactory.NewComputeService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute service", err)
}

// SyncFrontendDeployment implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendDeployment(ctx context.Context, logger logr.Logger, frontendDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendDeployment, mgr.objectFactory.NewFrontendDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend deployment", err)
}

// SyncFrontendService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendService(ctx context.Context, logger logr.Logger, frontendService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendService, mgr.objectFactory.NewFrontendService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend service", err)
}

// SyncMetaDeployment implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaDeployment(ctx context.Context, logger logr.Logger, metaDeployment *appsv1.Deployment) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, metaDeployment, mgr.objectFactory.NewMetaDeployment, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta deployment", err)
}

// SyncMetaService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaService(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, metaService, mgr.objectFactory.NewMetaService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta service", err)
}

// WaitBeforeCompactorDeploymentReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorDeploymentReady(ctx context.Context, logger logr.Logger, compactorDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if mgr.isObjectSynced(compactorDeployment) && utils.IsDeploymentRolledOut(compactorDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Compactor deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeComputeStatefulSetReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeStatefulSetReady(ctx context.Context, logger logr.Logger, computeStatefulSet *appsv1.StatefulSet) (reconcile.Result, error) {
	if mgr.isObjectSynced(computeStatefulSet) && utils.IsStatefulSetRolledOut(computeStatefulSet) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Compute statefulset hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeFrontendDeploymentReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendDeploymentReady(ctx context.Context, logger logr.Logger, frontendDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if mgr.isObjectSynced(frontendDeployment) && utils.IsDeploymentRolledOut(frontendDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Frontend deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeMetaDeploymentReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaDeploymentReady(ctx context.Context, logger logr.Logger, metaDeployment *appsv1.Deployment) (reconcile.Result, error) {
	if mgr.isObjectSynced(metaDeployment) && utils.IsDeploymentRolledOut(metaDeployment) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Meta deployment hasn't been rolled out")
		return ctrlkit.Exit()
	}
}

// WaitBeforeMetaServiceIsAvailable implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaServiceIsAvailable(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	if mgr.isObjectSynced(metaService) && utils.IsServiceReady(metaService) {
		return ctrlkit.NoRequeue()
	} else {
		logger.Info("Meta service hasn't been ready")
		return ctrlkit.Exit()
	}
}

func (mgr *risingWaveControllerManagerImpl) isObjectSyncedAndReady(obj client.Object) (bool, bool) {
	if isObjectNil(obj) {
		return false, false
	}
	switch obj := obj.(type) {
	case *corev1.Service:
		return mgr.isObjectSynced(obj), utils.IsServiceReady(obj)
	case *appsv1.Deployment:
		return mgr.isObjectSynced(obj), utils.IsDeploymentRolledOut(obj)
	case *appsv1.StatefulSet:
		return mgr.isObjectSynced(obj), utils.IsStatefulSetRolledOut(obj)
	default:
		return mgr.isObjectSynced(obj), true
	}
}

func (mgr *risingWaveControllerManagerImpl) reportComponentPhase(objs ...client.Object) risingwavev1alpha1.ComponentPhase {
	inInitializing := mgr.risingwaveManager.GetCondition(risingwavev1alpha1.Initializing) != nil

	if _, foundNil := lo.Find(objs, isObjectNil); foundNil {
		return risingwavev1alpha1.ComponentInitializing
	}

	synced, ready := true, true
	for _, obj := range objs {
		s, r := mgr.isObjectSyncedAndReady(obj)
		synced, ready = synced && s, ready && r
	}

	if synced && ready {
		return risingwavev1alpha1.ComponentReady
	} else {
		if inInitializing {
			return risingwavev1alpha1.ComponentInitializing
		} else {
			return risingwavev1alpha1.ComponentUpgrading
		}
	}
}

// CollectRunningStatisticsAndSyncStatus implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectRunningStatisticsAndSyncStatus(
	ctx context.Context,
	logger logr.Logger,
	frontendService *corev1.Service,
	metaService *corev1.Service,
	computeService *corev1.Service,
	compactorService *corev1.Service,
	metaDeployment *appsv1.Deployment,
	frontendDeployment *appsv1.Deployment,
	computeStatefulSet *appsv1.StatefulSet,
	compactorDeployment *appsv1.Deployment,
	configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// TODO support more object status here.
		if risingwave.Spec.ObjectStorage.Memory {
			status.ObjectStorage = risingwavev1alpha1.ObjectStorageStatus{
				Phase:       risingwavev1alpha1.ComponentReady,
				StorageType: risingwavev1alpha1.MemoryType,
			}
		}

		status.MetaNode = risingwavev1alpha1.MetaNodeStatus{
			Phase:    mgr.reportComponentPhase(configConfigMap, metaService, metaDeployment),
			Replicas: *risingwave.Spec.MetaNode.Replicas,
		}
		status.ComputeNode = risingwavev1alpha1.ComputeNodeStatus{
			Phase:    mgr.reportComponentPhase(configConfigMap, computeService, computeStatefulSet),
			Replicas: *risingwave.Spec.ComputeNode.Replicas,
		}
		status.Frontend = risingwavev1alpha1.FrontendSpecStatus{
			Phase:    mgr.reportComponentPhase(configConfigMap, frontendService, frontendDeployment),
			Replicas: *risingwave.Spec.Frontend.Replicas,
		}

		status.CompactorNode = risingwavev1alpha1.CompactorNodeStatus{
			Phase:    mgr.reportComponentPhase(configConfigMap, compactorService, compactorDeployment),
			Replicas: *risingwave.Spec.CompactorNode.Replicas,
		}
	})

	return ctrlkit.Continue()
}

// SyncConfigConfigMap implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConfigConfigMap(ctx context.Context, logger logr.Logger, configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, configConfigMap, mgr.objectFactory.NewConfigConfigMap, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync config configmap", err)
}

func NewRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager) RisingWaveControllerManagerImpl {
	return &risingWaveControllerManagerImpl{
		client:            client,
		risingwaveManager: risingwaveManager,
		objectFactory:     factory.NewRisingWaveObjectFactory(risingwaveManager.RisingWave(), client.Scheme()),
	}
}
