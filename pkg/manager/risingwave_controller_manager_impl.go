/*
 * Copyright 2023 RisingWave Labs
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
	"sort"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/event"

	"github.com/risingwavelabs/ctrlkit"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type risingWaveControllerManagerImpl struct {
	client             client.Client
	risingwaveManager  *object.RisingWaveManager
	objectFactory      *factory.RisingWaveObjectFactory
	eventMessageStore  *event.MessageStore
	forceUpdateEnabled bool
}

func buildNodeGroupStatus[T any, TP ptrAsObject[T], G any](groups []G, nameAndReplicas func(*G) (string, int32), workloads []T, groupAndReadyReplicas func(TP) (string, int32)) risingwavev1alpha1.ComponentReplicasStatus {
	status := risingwavev1alpha1.ComponentReplicasStatus{
		Target: 0,
	}

	expectedGroups := make(map[string]int32)

	for _, group := range groups {
		name, replicas := nameAndReplicas(&group)
		expectedGroups[name] = replicas
		status.Target += replicas
	}

	foundGroups := make(map[string]int)
	for _, obj := range workloads {
		group, readyReplicas := groupAndReadyReplicas(&obj)
		foundGroups[group] = 1
		status.Running += readyReplicas
		if replicas, ok := expectedGroups[group]; ok {
			status.Groups = append(status.Groups, risingwavev1alpha1.ComponentGroupReplicasStatus{
				Name:    group,
				Target:  replicas,
				Running: readyReplicas,
				Exists:  true,
			})
		} else {
			status.Groups = append(status.Groups, risingwavev1alpha1.ComponentGroupReplicasStatus{
				Name:    group + "(-)",
				Target:  0,
				Running: readyReplicas,
				Exists:  true,
			})
		}
	}

	// Groups expected but not found.
	for group, replicas := range expectedGroups {
		if _, ok := foundGroups[group]; !ok {
			status.Groups = append(status.Groups, risingwavev1alpha1.ComponentGroupReplicasStatus{
				Name:    group,
				Target:  replicas,
				Running: 0,
				Exists:  false,
			})
		}
	}

	// Sort the groups in status.
	sort.Slice(status.Groups, func(i, j int) bool {
		return status.Groups[i].Name < status.Groups[j].Name
	})

	return status
}

func isGroupMissing(group risingwavev1alpha1.ComponentGroupReplicasStatus) bool {
	return !group.Exists
}

func buildMetaStoreType(metaStore *risingwavev1alpha1.RisingWaveMetaStoreBackend) risingwavev1alpha1.RisingWaveMetaStoreBackendType {
	switch {
	case metaStore.Memory != nil && *metaStore.Memory:
		return risingwavev1alpha1.RisingWaveMetaStoreBackendTypeMemory
	case metaStore.Etcd != nil:
		return risingwavev1alpha1.RisingWaveMetaStoreBackendTypeEtcd
	default:
		return risingwavev1alpha1.RisingWaveMetaStoreBackendTypeUnknown
	}
}

func buildStateStoreType(stateStore *risingwavev1alpha1.RisingWaveStateStoreBackend) risingwavev1alpha1.RisingWaveStateStoreBackendType {
	switch {
	case ptr.Deref(stateStore.Memory, false):
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeMemory
	case stateStore.MinIO != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeMinIO
	case stateStore.S3 != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeS3
	case stateStore.GCS != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeGCS
	case stateStore.AliyunOSS != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeAliyunOSS
	case stateStore.AzureBlob != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeAzureBlob
	case stateStore.HDFS != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeHDFS
	case stateStore.WebHDFS != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeWebHDFS
	case stateStore.LocalDisk != nil:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeLocalDisk
	default:
		return risingwavev1alpha1.RisingWaveStateStoreBackendTypeUnknown
	}
}

// CollectOpenKruiseRunningStatisticsAndSyncStatus implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectOpenKruiseRunningStatisticsAndSyncStatus(ctx context.Context, logger logr.Logger,
	frontendService *corev1.Service, metaService *corev1.Service,
	computeService *corev1.Service, compactorService *corev1.Service, connectorService *corev1.Service,
	metaAdvancedStatefulSets []kruiseappsv1beta1.StatefulSet, frontendCloneSets []kruiseappsv1alpha1.CloneSet,
	computeStatefulSets []kruiseappsv1beta1.StatefulSet, compactorCloneSets []kruiseappsv1alpha1.CloneSet, connectorCloneSets []kruiseappsv1alpha1.CloneSet,
	configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	embeddedConnectorEnabled := mgr.risingwaveManager.IsEmbeddedConnectorEnabled()

	componentsSpec := &risingwave.Spec.Components

	// Update the replicas and storage status.
	getNameAndReplicasFromNodeGroup := func(g *risingwavev1alpha1.RisingWaveNodeGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	getGroupAndReadyReplicasForCloneSets := func(t *kruiseappsv1alpha1.CloneSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	getGroupAndReadyReplicasForStatefulSet := func(t *kruiseappsv1beta1.StatefulSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	componentReplicas := risingwavev1alpha1.RisingWaveComponentsReplicasStatus{
		Meta:      buildNodeGroupStatus(componentsSpec.Meta.NodeGroups, getNameAndReplicasFromNodeGroup, metaAdvancedStatefulSets, getGroupAndReadyReplicasForStatefulSet),
		Frontend:  buildNodeGroupStatus(componentsSpec.Frontend.NodeGroups, getNameAndReplicasFromNodeGroup, frontendCloneSets, getGroupAndReadyReplicasForCloneSets),
		Compactor: buildNodeGroupStatus(componentsSpec.Compactor.NodeGroups, getNameAndReplicasFromNodeGroup, compactorCloneSets, getGroupAndReadyReplicasForCloneSets),
		Connector: buildNodeGroupStatus(componentsSpec.Connector.NodeGroups, getNameAndReplicasFromNodeGroup, connectorCloneSets, getGroupAndReadyReplicasForCloneSets),
		Compute:   buildNodeGroupStatus(componentsSpec.Compute.NodeGroups, getNameAndReplicasFromNodeGroup, computeStatefulSets, getGroupAndReadyReplicasForStatefulSet),
	}
	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStore := &risingwave.Spec.MetaStore
		status.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreStatus{
			Backend: buildMetaStoreType(metaStore),
		}

		// Report object storage status.
		stateStore := &risingwave.Spec.StateStore
		status.StateStore = risingwavev1alpha1.RisingWaveStateStoreStatus{
			Backend: buildStateStoreType(stateStore),
		}

		// Report Version status.
		status.Version = utils.GetVersionFromImage(mgr.risingwaveManager.RisingWave().Spec.Image)

		// Report component replicas.
		status.ComponentReplicas = componentReplicas
	})

	// If any of these states is missing or any of the groups is missing, turn the condition Running to false.
	recoverConditionAndReasons := []struct {
		cond      bool
		component string
	}{
		{
			cond:      frontendService == nil,
			component: "Service(frontend)",
		},
		{
			cond:      metaService == nil,
			component: "Service(meta)",
		},
		{
			cond:      computeService == nil,
			component: "Service(compute)",
		},
		{
			cond:      compactorService == nil,
			component: "Service(compactor)",
		},
		{
			cond:      !embeddedConnectorEnabled && connectorService == nil,
			component: "Service(connector)",
		},
		{
			cond:      configConfigMap == nil,
			component: "ConfigMap(config)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Meta.Groups, isGroupMissing),
			component: "CloneSets(meta)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Frontend.Groups, isGroupMissing),
			component: "CloneSets(frontend)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Compute.Groups, isGroupMissing),
			component: "AdvancedStatefulSets(compute)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Compactor.Groups, isGroupMissing),
			component: "CloneSets(compactor)",
		},
		{
			cond:      !embeddedConnectorEnabled && lo.ContainsBy(componentReplicas.Connector.Groups, isGroupMissing),
			component: "CloneSets(connector)",
		},
	}

	brokenOrMissingComponents := lo.FilterMap(recoverConditionAndReasons, func(t struct {
		cond      bool
		component string
	}, _ int) (string, bool) {
		return t.component, t.cond
	})

	if len(brokenOrMissingComponents) > 0 {
		mgr.risingwaveManager.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionFalse,
		})

		// Set the message for Unhealthy event when it's Running.
		if mgr.risingwaveManager.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true) {
			mgr.eventMessageStore.SetMessage(consts.RisingWaveEventTypeUnhealthy.Name, fmt.Sprintf("Found components broken or missing: %s", strings.Join(brokenOrMissingComponents, ",")))
		}
	}

	return ctrlkit.Continue()
}

// CollectRunningStatisticsAndSyncStatus implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectRunningStatisticsAndSyncStatus(ctx context.Context, logger logr.Logger,
	frontendService *corev1.Service, metaService *corev1.Service,
	computeService *corev1.Service, compactorService *corev1.Service, connectorService *corev1.Service,
	metaStatefulSets []appsv1.StatefulSet, frontendDeployments []appsv1.Deployment,
	computeStatefulSets []appsv1.StatefulSet, compactorDeployments []appsv1.Deployment, connectorDeployments []appsv1.Deployment,
	configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	embeddedConnectorEnabled := mgr.risingwaveManager.IsEmbeddedConnectorEnabled()

	componentsSpec := &risingwave.Spec.Components

	getNameAndReplicasFromNodeGroup := func(g *risingwavev1alpha1.RisingWaveNodeGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	getGroupAndReadyReplicasForDeployment := func(t *appsv1.Deployment) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	getGroupAndReadyReplicasForStatefulSet := func(t *appsv1.StatefulSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	componentReplicas := risingwavev1alpha1.RisingWaveComponentsReplicasStatus{
		Meta:      buildNodeGroupStatus(componentsSpec.Meta.NodeGroups, getNameAndReplicasFromNodeGroup, metaStatefulSets, getGroupAndReadyReplicasForStatefulSet),
		Frontend:  buildNodeGroupStatus(componentsSpec.Frontend.NodeGroups, getNameAndReplicasFromNodeGroup, frontendDeployments, getGroupAndReadyReplicasForDeployment),
		Compactor: buildNodeGroupStatus(componentsSpec.Compactor.NodeGroups, getNameAndReplicasFromNodeGroup, compactorDeployments, getGroupAndReadyReplicasForDeployment),
		Connector: buildNodeGroupStatus(componentsSpec.Connector.NodeGroups, getNameAndReplicasFromNodeGroup, connectorDeployments, getGroupAndReadyReplicasForDeployment),
		Compute:   buildNodeGroupStatus(componentsSpec.Compute.NodeGroups, getNameAndReplicasFromNodeGroup, computeStatefulSets, getGroupAndReadyReplicasForStatefulSet),
	}

	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStore := &risingwave.Spec.MetaStore
		status.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreStatus{
			Backend: buildMetaStoreType(metaStore),
		}

		// Report object storage status.
		stateStore := &risingwave.Spec.StateStore
		status.StateStore = risingwavev1alpha1.RisingWaveStateStoreStatus{
			Backend: buildStateStoreType(stateStore),
		}

		// Report Version status.
		status.Version = utils.GetVersionFromImage(mgr.risingwaveManager.RisingWave().Spec.Image)

		// Report component replicas.
		status.ComponentReplicas = componentReplicas
	})

	// If any of these states is missing or any of the groups is missing, turn the condition Running to false.
	recoverConditionAndReasons := []struct {
		cond      bool
		component string
	}{
		{
			cond:      frontendService == nil,
			component: "Service(frontend)",
		},
		{
			cond:      metaService == nil,
			component: "Service(meta)",
		},
		{
			cond:      computeService == nil,
			component: "Service(compute)",
		},
		{
			cond:      compactorService == nil,
			component: "Service(compactor)",
		},
		{
			cond:      !embeddedConnectorEnabled && connectorService == nil,
			component: "Service(connector)",
		},
		{
			cond:      configConfigMap == nil,
			component: "ConfigMap(config)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Meta.Groups, isGroupMissing),
			component: "Deployments(meta)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Frontend.Groups, isGroupMissing),
			component: "Deployments(frontend)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Compute.Groups, isGroupMissing),
			component: "StatefulSets(compute)",
		},
		{
			cond:      lo.ContainsBy(componentReplicas.Compactor.Groups, isGroupMissing),
			component: "Deployments(compactor)",
		},
		{
			cond:      !embeddedConnectorEnabled && lo.ContainsBy(componentReplicas.Connector.Groups, isGroupMissing),
			component: "Deployments(connector)",
		},
	}

	brokenOrMissingComponents := lo.FilterMap(recoverConditionAndReasons, func(t struct {
		cond      bool
		component string
	}, _ int) (string, bool) {
		return t.component, t.cond
	})

	if len(brokenOrMissingComponents) > 0 {
		mgr.risingwaveManager.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionFalse,
		})

		// Set the message for Unhealthy event when it's Running.
		if mgr.risingwaveManager.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true) {
			mgr.eventMessageStore.SetMessage(consts.RisingWaveEventTypeUnhealthy.Name, fmt.Sprintf("Found components broken or missing: %s", strings.Join(brokenOrMissingComponents, ",")))
		}
	}

	return ctrlkit.Continue()
}

type ptrAsObject[T any] interface {
	client.Object
	*T
}

func syncComponentGroupWorkloads[T any, TP ptrAsObject[T]](
	mgr *risingWaveControllerManagerImpl,
	ctx context.Context,
	logger logr.Logger,
	component string,
	objects []T,
	factory func(group string) TP,
	enabled bool,
) (reconcile.Result, error) {
	logger = logger.WithValues("component", component)

	var expectedGroupSet map[string]int
	if enabled {
		expectedGroupSet = buildKeyMapFromList(mgr.risingwaveManager.GetNodeGroups(component), getNameFromNodeGroup)
	}

	// Decide to delete or to sync.
	observedGroupSet := make(map[string]int)
	toDelete := make([]TP, 0)
	toSyncGroupObjects := make(map[string]TP, 0)
	foundGroups := make(map[string]int)
	for i := range objects {
		workloadObjPtr := TP(&objects[i])
		group := workloadObjPtr.GetLabels()[consts.LabelRisingWaveGroup]
		foundGroups[group] = 1
		if _, exists := observedGroupSet[group]; exists {
			logger.Info("Duplicate group found, mark as to delete", "group", group, "workload", workloadObjPtr.GetName())
			toDelete = append(toDelete, workloadObjPtr)
		} else {
			if !mgr.isObjectSynced(workloadObjPtr) {
				_, expectExists := expectedGroupSet[group]
				if expectExists {
					toSyncGroupObjects[group] = workloadObjPtr
				} else {
					toDelete = append(toDelete, workloadObjPtr)
				}
			}
		}
		observedGroupSet[group] = 1
	}

	for group := range expectedGroupSet {
		if _, found := foundGroups[group]; !found {
			toSyncGroupObjects[group] = TP(nil) // Not found
		}
	}

	// Delete the unexpected. Note it won't delete any workload object that is created with a newer generation,
	// so it is safe to do the deletion.
	for _, workloadObj := range toDelete {
		group := workloadObj.GetLabels()[consts.LabelRisingWaveGroup]
		if err := mgr.client.Delete(ctx, workloadObj, client.PropagationPolicy(metav1.DeletePropagationBackground)); client.IgnoreNotFound(err) != nil {
			logger.Error(err, "Failed to delete object", "workload", workloadObj.GetName(), "group", group)
			return ctrlkit.RequeueIfErrorAndWrap("unable to delete object", err)
		}
	}

	// Sync the outdated.
	if len(toSyncGroupObjects) > 0 {
		for group, workloadObj := range toSyncGroupObjects {
			if err := syncObject(mgr, ctx, workloadObj, func() TP {
				return factory(group)
			}, logger.WithValues("group", group)); err != nil {
				return ctrlkit.RequeueIfErrorAndWrap("unable to sync object", err)
			}
		}
	}

	return ctrlkit.Continue()
}

func getNameFromNodeGroup(g *risingwavev1alpha1.RisingWaveNodeGroup) string {
	return g.Name
}

func buildKeyMapFromList[Elem any](list []Elem, key func(*Elem) string) map[string]int {
	r := make(map[string]int)
	for _, group := range list {
		name := key(&group)
		r[name] = 1
	}
	return r
}

// SyncCompactorDeployments implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorDeployments(ctx context.Context, logger logr.Logger, compactorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentCompactor,
		compactorDeployments, mgr.objectFactory.NewCompactorDeployment,
		// Only sync if Open Kruise is enabled.
		!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncCompactorCloneSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorCloneSets(ctx context.Context, logger logr.Logger, compactorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentCompactor,
		compactorCloneSets, mgr.objectFactory.NewCompactorCloneSet,
		// Only sync if Open Kruise is enabled.
		mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncConnectorDeployments implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConnectorDeployments(ctx context.Context, logger logr.Logger, connectorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentConnector,
		connectorDeployments, mgr.objectFactory.NewConnectorDeployment,
		// Only sync if Open Kruise is disabled.
		!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncConnectorCloneSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConnectorCloneSets(ctx context.Context, logger logr.Logger, connectorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentConnector,
		connectorCloneSets, mgr.objectFactory.NewConnectorCloneSet,
		// Only sync if Open Kruise is enabled.
		mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncComputeStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeStatefulSets(ctx context.Context, logger logr.Logger, computeStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentCompute,
		computeStatefulSets, mgr.objectFactory.NewComputeStatefulSet,
		// Only sync if Open Kruise is disabled.
		!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncComputeAdvancedStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeAdvancedStatefulSets(ctx context.Context, logger logr.Logger, computeStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentCompute,
		computeStatefulSets, mgr.objectFactory.NewComputeAdvancedStatefulSet,
		// Only sync if Open Kruise is enabled.
		mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncFrontendDeployments implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendDeployments(ctx context.Context, logger logr.Logger, frontendDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentFrontend,
		frontendDeployments, mgr.objectFactory.NewFrontendDeployment,
		// Only sync if Open Kruise is disabled.
		!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncFrontendCloneSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendCloneSets(ctx context.Context, logger logr.Logger, frontendCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentFrontend,
		frontendCloneSets, mgr.objectFactory.NewFrontendCloneSet,
		// Only sync if Open Kruise is enabled.
		mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncMetaStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaStatefulSets(ctx context.Context, logger logr.Logger, metaStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentMeta,
		metaStatefulSets, mgr.objectFactory.NewMetaStatefulSet,
		// Only sync if Open Kruise is disabled.
		!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

// SyncMetaAdvancedStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaAdvancedStatefulSets(ctx context.Context, logger logr.Logger, metaStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	return syncComponentGroupWorkloads(mgr, ctx, logger,
		consts.ComponentMeta,
		metaStatefulSets, mgr.objectFactory.NewMetaAdvancedStatefulSet,
		// Only sync if Open Kruise is enabled.
		mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
	)
}

func waitComponentGroupWorkloadsReady[T any, TP ptrAsObject[T]](ctx context.Context, logger logr.Logger, component string,
	groups map[string]int, objects []T, isReady func(TP) bool) (reconcile.Result, error) {
	logger = logger.WithValues("component", component)

	foundGroups := make(map[string]int)
	for _, workloadObj := range objects {
		group := TP(&workloadObj).GetLabels()[consts.LabelRisingWaveGroup]
		foundGroups[group] = 1
		_, expectGroup := groups[group]
		if !expectGroup {
			logger.Info("Found unexpected group, keep waiting...", "group", group)
			return ctrlkit.Exit()
		}

		if !isReady(&workloadObj) {
			logger.Info("Found not-ready groups, keep waiting...", "group", group)
			return ctrlkit.Exit()
		}
	}

	for group := range groups {
		if _, found := foundGroups[group]; !found {
			logger.Info("Workload object not found, keep waiting...", "group", group)
			return ctrlkit.Exit()
		}
	}

	return ctrlkit.Continue()
}

func (mgr *risingWaveControllerManagerImpl) buildExpectedGroupSet(component string) map[string]int {
	return buildKeyMapFromList(mgr.risingwaveManager.GetNodeGroups(component), getNameFromNodeGroup)
}

// WaitBeforeCompactorDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorDeploymentsReady(ctx context.Context, logger logr.Logger, compactorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentCompactor,
		lo.If(!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentCompactor)).Else(nil),
		compactorDeployments,
		func(t *appsv1.Deployment) bool { return mgr.isObjectSynced(t) && utils.IsDeploymentRolledOut(t) },
	)
}

// WaitBeforeCompactorCloneSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorCloneSetsReady(ctx context.Context, logger logr.Logger, compactorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentCompactor,
		lo.If(mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentCompactor)).Else(nil),
		compactorCloneSets,
		func(t *kruiseappsv1alpha1.CloneSet) bool {
			return mgr.isObjectSynced(t) && utils.IsCloneSetRolledOut(t)
		},
	)
}

// WaitBeforeConnectorDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeConnectorDeploymentsReady(ctx context.Context, logger logr.Logger, connectorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentConnector,
		lo.If(!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentConnector)).Else(nil),
		connectorDeployments,
		func(t *appsv1.Deployment) bool { return mgr.isObjectSynced(t) && utils.IsDeploymentRolledOut(t) },
	)
}

// WaitBeforeConnectorCloneSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeConnectorCloneSetsReady(ctx context.Context, logger logr.Logger, connectorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentConnector,
		lo.If(mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentConnector)).Else(nil),
		connectorCloneSets,
		func(t *kruiseappsv1alpha1.CloneSet) bool {
			return mgr.isObjectSynced(t) && utils.IsCloneSetRolledOut(t)
		},
	)
}

// WaitBeforeComputeStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeStatefulSetsReady(ctx context.Context, logger logr.Logger, computeStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentCompute,
		lo.If(!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentCompute)).Else(nil),
		computeStatefulSets,
		func(t *appsv1.StatefulSet) bool { return mgr.isObjectSynced(t) && utils.IsStatefulSetRolledOut(t) },
	)
}

// WaitBeforeComputeAdvancedStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeAdvancedStatefulSetsReady(ctx context.Context, logger logr.Logger, computeStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentCompute,
		lo.If(mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentCompute)).Else(nil),
		computeStatefulSets,
		func(t *kruiseappsv1beta1.StatefulSet) bool {
			return mgr.isObjectSynced(t) && utils.IsAdvancedStatefulSetRolledOut(t)
		},
	)
}

// WaitBeforeFrontendDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendDeploymentsReady(ctx context.Context, logger logr.Logger, frontendDeployments []appsv1.Deployment) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentFrontend,
		lo.If(!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentFrontend)).Else(nil),
		frontendDeployments,
		func(t *appsv1.Deployment) bool { return mgr.isObjectSynced(t) && utils.IsDeploymentRolledOut(t) },
	)
}

// WaitBeforeFrontendCloneSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendCloneSetsReady(ctx context.Context, logger logr.Logger, frontendCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentFrontend,
		lo.If(mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentFrontend)).Else(nil),
		frontendCloneSets,
		func(t *kruiseappsv1alpha1.CloneSet) bool {
			return mgr.isObjectSynced(t) && utils.IsCloneSetRolledOut(t)
		},
	)
}

// WaitBeforeMetaStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaStatefulSetsReady(ctx context.Context, logger logr.Logger, metaStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentMeta,
		lo.If(!mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentMeta)).Else(nil),
		metaStatefulSets,
		func(t *appsv1.StatefulSet) bool { return mgr.isObjectSynced(t) && utils.IsStatefulSetRolledOut(t) },
	)
}

// WaitBeforeMetaAdvancedStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaAdvancedStatefulSetsReady(ctx context.Context, logger logr.Logger, metaAdvancedStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	return waitComponentGroupWorkloadsReady(ctx, logger, consts.ComponentMeta,
		lo.If(mgr.risingwaveManager.IsOpenKruiseEnabled() && !mgr.risingwaveManager.IsStandaloneModeEnabled(),
			mgr.buildExpectedGroupSet(consts.ComponentMeta)).Else(nil),
		metaAdvancedStatefulSets,
		func(t *kruiseappsv1beta1.StatefulSet) bool {
			return mgr.isObjectSynced(t) && utils.IsAdvancedStatefulSetRolledOut(t)
		},
	)
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
	if !isObjectNil(obj) && !isObjectNil(newObj) {
		if obj.GetName() != newObj.GetName() || obj.GetNamespace() != newObj.GetNamespace() {
			panic(fmt.Sprintf("objects not the same: %s/%s vs. %s/%s",
				obj.GetNamespace(), obj.GetName(),
				newObj.GetNamespace(), newObj.GetName(),
			))
		}
	} else if obj == nil || newObj == nil {
		panic("objects not the same: either interface is nil")
	} else if isObjectNil(newObj) {
		panic("objects not the same: new object is nil")
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

func (mgr *risingWaveControllerManagerImpl) syncObject(ctx context.Context, obj client.Object, factory func() (client.Object, error), logger logr.Logger) error {
	scheme := mgr.client.Scheme()

	if isObjectNil(obj) {
		// Not found. Going to create one.
		newObj, err := factory()
		if err != nil {
			return fmt.Errorf("unable to build new object: %w", err)
		}
		newObj = ensureTheSameObject(obj, newObj)

		gvk, err := apiutil.GVKForObject(newObj, scheme)
		if err != nil {
			return err
		}

		logger.Info(fmt.Sprintf("Create an object of %s", gvk.Kind), "object", utils.GetNamespacedName(newObj))
		err = mgr.client.Create(ctx, newObj)
		return client.IgnoreAlreadyExists(err)
	}

	gvk, err := apiutil.GVKForObject(obj, scheme)
	if err != nil {
		return err
	}

	// Found. Update/Sync if not synced.
	if !mgr.isObjectSynced(obj) {
		newObj, err := factory()
		if err != nil {
			return fmt.Errorf("unable to build new object: %w", err)
		}
		newObj = ensureTheSameObject(obj, newObj)
		// Set the resource version for update.
		newObj.SetResourceVersion(obj.GetResourceVersion())
		logger.Info(fmt.Sprintf("Update the object of %s", gvk.Kind), "object", utils.GetNamespacedName(newObj),
			"generation", mgr.risingwaveManager.RisingWave().Generation)
		if err = mgr.client.Update(ctx, newObj); err == nil {
			return nil
		}
		if !apierrors.IsInvalid(err) {
			return err
		}
		if !mgr.forceUpdateEnabled ||
			obj.GetLabels()[consts.LabelRisingWaveOperatorVersion] == newObj.GetLabels()[consts.LabelRisingWaveOperatorVersion] {
			return err
		}
		if err := mgr.client.Delete(ctx, obj); err != nil {
			return err
		}
		newObj.SetResourceVersion("")
		if err := mgr.client.Create(ctx, newObj); err != nil {
			return client.IgnoreAlreadyExists(err)
		}
	}
	return nil
}

// Helper function for compile time type assertion.
func syncObject[T client.Object](mgr *risingWaveControllerManagerImpl, ctx context.Context, obj T, factory func() T, logger logr.Logger) error {
	return mgr.syncObject(ctx, obj, func() (client.Object, error) {
		return factory(), nil
	}, logger)
}

// SyncCompactorService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorService(ctx context.Context, logger logr.Logger, compactorService *corev1.Service) (reconcile.Result, error) {
	if !mgr.risingwaveManager.IsStandaloneModeEnabled() {
		err := syncObject(mgr, ctx, compactorService, mgr.objectFactory.NewCompactorService, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor service", err)
	}
	if compactorService != nil {
		err := mgr.client.Delete(ctx, compactorService, client.Preconditions{UID: &compactorService.UID})
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor service", err)
	}
	return ctrlkit.NoRequeue()
}

// SyncConnectorService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConnectorService(ctx context.Context, logger logr.Logger, connectorService *corev1.Service) (reconcile.Result, error) {
	if !mgr.risingwaveManager.IsStandaloneModeEnabled() {
		err := syncObject(mgr, ctx, connectorService, mgr.objectFactory.NewConnectorService, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync connector service", err)
	}
	if connectorService != nil {
		err := mgr.client.Delete(ctx, connectorService, client.Preconditions{UID: &connectorService.UID})
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync connector service", err)
	}
	return ctrlkit.NoRequeue()
}

// SyncComputeService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeService(ctx context.Context, logger logr.Logger, computeService *corev1.Service) (reconcile.Result, error) {
	if !mgr.risingwaveManager.IsStandaloneModeEnabled() {
		err := syncObject(mgr, ctx, computeService, mgr.objectFactory.NewComputeService, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute service", err)
	}
	if computeService != nil {
		err := mgr.client.Delete(ctx, computeService, client.Preconditions{UID: &computeService.UID})
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute service", err)
	}
	return ctrlkit.NoRequeue()
}

// SyncFrontendService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendService(ctx context.Context, logger logr.Logger, frontendService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendService, mgr.objectFactory.NewFrontendService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend service", err)
}

// SyncMetaService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaService(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	if !mgr.risingwaveManager.IsStandaloneModeEnabled() {
		err := syncObject(mgr, ctx, metaService, mgr.objectFactory.NewMetaService, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta service", err)
	}
	if metaService != nil {
		err := mgr.client.Delete(ctx, metaService, client.Preconditions{UID: &metaService.UID})
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta service", err)
	}
	return ctrlkit.NoRequeue()
}

// WaitBeforeMetaServiceIsAvailable implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaServiceIsAvailable(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	if !mgr.risingwaveManager.IsStandaloneModeEnabled() {
		if mgr.isObjectSynced(metaService) {
			return ctrlkit.NoRequeue()
		}
		logger.Info("Meta service hasn't been ready")
		return ctrlkit.Exit()
	}
	if metaService != nil {
		logger.Info("Meta service hasn't been deleted")
		return ctrlkit.Exit()
	}
	return ctrlkit.NoRequeue()
}

// SyncConfigConfigMap implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConfigConfigMap(ctx context.Context, logger logr.Logger, configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, configConfigMap, func() *corev1.ConfigMap {
		return mgr.objectFactory.NewConfigConfigMap("")
	}, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync config configmap", err)
}

// SyncServiceMonitor implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncServiceMonitor(ctx context.Context, logger logr.Logger, serviceMonitor *monitoringv1.ServiceMonitor) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, serviceMonitor, mgr.objectFactory.NewServiceMonitor, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync service monitor", err)
}

// SyncStandaloneService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncStandaloneService(ctx context.Context, logger logr.Logger, standaloneService *corev1.Service) (ctrl.Result, error) {
	if mgr.risingwaveManager.IsStandaloneModeEnabled() {
		err := syncObject(mgr, ctx, standaloneService, mgr.objectFactory.NewStandaloneService, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone service", err)
	} else if standaloneService != nil {
		err := mgr.client.Delete(ctx, standaloneService, client.Preconditions{UID: &standaloneService.UID})
		if err != nil {
			logger.Error(err, "Failed to delete standalone service!", "service", standaloneService.Name)
		}
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone service", err)
	}
	return ctrlkit.NoRequeue()
}

// SyncStandaloneStatefulSet implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncStandaloneStatefulSet(ctx context.Context, logger logr.Logger, standaloneStatefulSet *appsv1.StatefulSet) (ctrl.Result, error) {
	if mgr.risingwaveManager.IsStandaloneModeEnabled() && !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		err := syncObject(mgr, ctx, standaloneStatefulSet, mgr.objectFactory.NewStandaloneStatefulSet, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone statefulset", err)
	} else if standaloneStatefulSet != nil {
		err := mgr.client.Delete(ctx, standaloneStatefulSet, client.Preconditions{UID: &standaloneStatefulSet.UID})
		if err != nil {
			logger.Error(err, "Failed to delete standalone statefulset!", "sts", standaloneStatefulSet.Name)
		}
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone statefulset", err)
	}
	return ctrlkit.NoRequeue()
}

// SyncStandaloneAdvancedStatefulSet implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncStandaloneAdvancedStatefulSet(ctx context.Context, logger logr.Logger, standaloneAdvancedStatefulSet *kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
	if mgr.risingwaveManager.IsStandaloneModeEnabled() && mgr.risingwaveManager.IsOpenKruiseEnabled() {
		err := syncObject(mgr, ctx, standaloneAdvancedStatefulSet, mgr.objectFactory.NewStandaloneAdvancedStatefulSet, logger)
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone statefulset", err)
	} else if standaloneAdvancedStatefulSet != nil {
		err := mgr.client.Delete(ctx, standaloneAdvancedStatefulSet, client.Preconditions{UID: &standaloneAdvancedStatefulSet.UID})
		if err != nil {
			logger.Error(err, "Failed to delete standalone statefulset!", "sts", standaloneAdvancedStatefulSet.Name)
		}
		return ctrlkit.RequeueIfErrorAndWrap("unable to sync standalone statefulset", err)
	}
	return ctrlkit.NoRequeue()
}

// WaitBeforeStandaloneStatefulSetReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeStandaloneStatefulSetReady(ctx context.Context, logger logr.Logger, standaloneStatefulSet *appsv1.StatefulSet) (ctrl.Result, error) {
	if mgr.risingwaveManager.IsStandaloneModeEnabled() && !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		if mgr.isObjectSynced(standaloneStatefulSet) && utils.IsStatefulSetRolledOut(standaloneStatefulSet) {
			return ctrlkit.NoRequeue()
		}
		logger.Info("Standalone StatefulSet hasn't been ready!")
		return ctrlkit.Exit()
	} else if standaloneStatefulSet != nil {
		logger.Info("Standalone StatefulSet hasn't been deleted!")
		return ctrlkit.Exit()
	}
	return ctrlkit.NoRequeue()
}

// WaitBeforeStandaloneAdvancedStatefulSetReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeStandaloneAdvancedStatefulSetReady(ctx context.Context, logger logr.Logger, standaloneAdvancedStatefulSet *kruiseappsv1beta1.StatefulSet) (ctrl.Result, error) {
	if mgr.risingwaveManager.IsStandaloneModeEnabled() && mgr.risingwaveManager.IsOpenKruiseEnabled() {
		if mgr.isObjectSynced(standaloneAdvancedStatefulSet) && utils.IsAdvancedStatefulSetRolledOut(standaloneAdvancedStatefulSet) {
			return ctrlkit.NoRequeue()
		}
		logger.Info("Standalone advanced StatefulSet hasn't been ready!")
		return ctrlkit.Exit()
	} else if standaloneAdvancedStatefulSet != nil {
		logger.Info("Standalone advanced StatefulSet hasn't been deleted!")
		return ctrlkit.Exit()
	}
	return ctrlkit.NoRequeue()
}

// CollectRunningStatisticsAndSyncStatusForStandalone implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectRunningStatisticsAndSyncStatusForStandalone(ctx context.Context, logger logr.Logger, standaloneService *corev1.Service, standaloneStatefulSet *appsv1.StatefulSet, configConfigMap *corev1.ConfigMap) (ctrl.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStore := &risingwave.Spec.MetaStore
		status.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreStatus{
			Backend: buildMetaStoreType(metaStore),
		}

		// Report object storage status.
		stateStore := &risingwave.Spec.StateStore
		status.StateStore = risingwavev1alpha1.RisingWaveStateStoreStatus{
			Backend: buildStateStoreType(stateStore),
		}

		// Report Version status.
		status.Version = utils.GetVersionFromImage(mgr.risingwaveManager.RisingWave().Spec.Image)

		// Report component replicas.
		status.ComponentReplicas = risingwavev1alpha1.RisingWaveComponentsReplicasStatus{}
	})

	recoverConditionAndReasons := []struct {
		cond      bool
		component string
	}{
		{
			cond:      standaloneService == nil,
			component: "Service(standalone)",
		},
		{
			cond:      configConfigMap == nil,
			component: "ConfigMap(config)",
		},
		{
			cond:      standaloneStatefulSet == nil,
			component: "StatefulSet(standalone)",
		},
	}

	brokenOrMissingComponents := lo.FilterMap(recoverConditionAndReasons, func(t struct {
		cond      bool
		component string
	}, _ int) (string, bool) {
		return t.component, t.cond
	})

	if len(brokenOrMissingComponents) > 0 {
		mgr.risingwaveManager.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionFalse,
		})

		// Set the message for Unhealthy event when it's Running.
		if mgr.risingwaveManager.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true) {
			mgr.eventMessageStore.SetMessage(consts.RisingWaveEventTypeUnhealthy.Name, fmt.Sprintf("Found components broken or missing: %s", strings.Join(brokenOrMissingComponents, ",")))
		}
	}

	return ctrlkit.Continue()
}

// CollectOpenKruiseRunningStatisticsAndSyncStatusForStandalone implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectOpenKruiseRunningStatisticsAndSyncStatusForStandalone(ctx context.Context, logger logr.Logger, standaloneService *corev1.Service, standaloneAdvancedStatefulSet *kruiseappsv1beta1.StatefulSet, configConfigMap *corev1.ConfigMap) (ctrl.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStore := &risingwave.Spec.MetaStore
		status.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreStatus{
			Backend: buildMetaStoreType(metaStore),
		}

		// Report object storage status.
		stateStore := &risingwave.Spec.StateStore
		status.StateStore = risingwavev1alpha1.RisingWaveStateStoreStatus{
			Backend: buildStateStoreType(stateStore),
		}

		// Report Version status.
		status.Version = utils.GetVersionFromImage(mgr.risingwaveManager.RisingWave().Spec.Image)

		// Report component replicas.
		status.ComponentReplicas = risingwavev1alpha1.RisingWaveComponentsReplicasStatus{}
	})

	recoverConditionAndReasons := []struct {
		cond      bool
		component string
	}{
		{
			cond:      standaloneService == nil,
			component: "Service(standalone)",
		},
		{
			cond:      configConfigMap == nil,
			component: "ConfigMap(config)",
		},
		{
			cond:      standaloneAdvancedStatefulSet == nil,
			component: "AdvancedStatefulSet(standalone)",
		},
	}

	brokenOrMissingComponents := lo.FilterMap(recoverConditionAndReasons, func(t struct {
		cond      bool
		component string
	}, _ int) (string, bool) {
		return t.component, t.cond
	})

	if len(brokenOrMissingComponents) > 0 {
		mgr.risingwaveManager.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
			Type:   risingwavev1alpha1.RisingWaveConditionRunning,
			Status: metav1.ConditionFalse,
		})

		// Set the message for Unhealthy event when it's Running.
		if mgr.risingwaveManager.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true) {
			mgr.eventMessageStore.SetMessage(consts.RisingWaveEventTypeUnhealthy.Name, fmt.Sprintf("Found components broken or missing: %s", strings.Join(brokenOrMissingComponents, ",")))
		}
	}

	return ctrlkit.Continue()
}

func newRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager, messageStore *event.MessageStore, forceUpdateEnabled bool, operatorVersion string) *risingWaveControllerManagerImpl {
	return &risingWaveControllerManagerImpl{
		client:             client,
		risingwaveManager:  risingwaveManager,
		objectFactory:      factory.NewRisingWaveObjectFactory(risingwaveManager.RisingWave(), client.Scheme(), operatorVersion),
		eventMessageStore:  messageStore,
		forceUpdateEnabled: forceUpdateEnabled,
	}
}

// NewRisingWaveControllerManagerImpl creates an object that implements the RisingWaveControllerManagerImpl.
func NewRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager, messageStore *event.MessageStore, forceUpdateEnabled bool, operatorVersion string) RisingWaveControllerManagerImpl {
	return newRisingWaveControllerManagerImpl(client, risingwaveManager, messageStore, forceUpdateEnabled, operatorVersion)
}
