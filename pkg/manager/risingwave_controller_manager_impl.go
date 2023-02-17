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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/event"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

type risingWaveControllerManagerImpl struct {
	client            client.Client
	risingwaveManager *object.RisingWaveManager
	objectFactory     *factory.RisingWaveObjectFactory
	eventMessageStore *event.MessageStore
}

func buildGroupStatus[T any, TP ptrAsObject[T], G any](globalReplicas int32, groups []G, nameAndReplicas func(*G) (string, int32), workloads []T, groupAndReadyReplicas func(*T) (string, int32)) risingwavev1alpha1.ComponentReplicasStatus {
	status := risingwavev1alpha1.ComponentReplicasStatus{
		Target: globalReplicas,
	}

	expectedGroups := make(map[string]int32)
	if globalReplicas > 0 {
		expectedGroups[""] = globalReplicas
	}
	status.Target = globalReplicas
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

func buildMetaStorageType(metaStorage *risingwavev1alpha1.RisingWaveMetaStorage) risingwavev1alpha1.MetaStorageType {
	switch {
	case metaStorage.Memory != nil && *metaStorage.Memory:
		return risingwavev1alpha1.MetaStorageTypeMemory
	case metaStorage.Etcd != nil:
		return risingwavev1alpha1.MetaStorageTypeEtcd
	default:
		return risingwavev1alpha1.MetaStorageTypeUnknown
	}
}

func buildObjectStorageType(objectStorage *risingwavev1alpha1.RisingWaveObjectStorage) risingwavev1alpha1.ObjectStorageType {
	switch {
	case pointer.BoolDeref(objectStorage.Memory, false):
		return risingwavev1alpha1.ObjectStorageTypeMemory
	case objectStorage.MinIO != nil:
		return risingwavev1alpha1.ObjectStorageTypeMinIO
	case objectStorage.S3 != nil:
		return risingwavev1alpha1.ObjectStorageTypeS3
	case objectStorage.AliyunOSS != nil:
		return risingwavev1alpha1.ObjectStorageTypeAliyunOSS
	default:
		return risingwavev1alpha1.ObjectStorageTypeUnknown
	}
}

// CollectOpenKruiseRunningStatisticsAndSyncStatus implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) CollectOpenKruiseRunningStatisticsAndSyncStatus(ctx context.Context, logger logr.Logger,
	frontendService *corev1.Service, metaService *corev1.Service,
	computeService *corev1.Service, compactorService *corev1.Service,
	metaAdvancedStatefulSets []kruiseappsv1beta1.StatefulSet, frontendCloneSets []kruiseappsv1alpha1.CloneSet,
	computeStatefulSets []kruiseappsv1beta1.StatefulSet, compactorCloneSets []kruiseappsv1alpha1.CloneSet,
	configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	globalSpec := &risingwave.Spec.Global
	componentsSpec := &risingwave.Spec.Components

	// Update the replicas and storage status.
	nameAndReplicasForComponentGroup := func(g *risingwavev1alpha1.RisingWaveComponentGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	groupAndReadyReplicasForCloneSets := func(t *kruiseappsv1alpha1.CloneSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	nameAndReplicasForComputeGroup := func(g *risingwavev1alpha1.RisingWaveComputeGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	groupAndReadyReplicasForStatefulSet := func(t *kruiseappsv1beta1.StatefulSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	componentReplicas := risingwavev1alpha1.RisingWaveComponentsReplicasStatus{
		Meta:      buildGroupStatus(globalSpec.Replicas.Meta, componentsSpec.Meta.Groups, nameAndReplicasForComponentGroup, metaAdvancedStatefulSets, groupAndReadyReplicasForStatefulSet),
		Frontend:  buildGroupStatus(globalSpec.Replicas.Frontend, componentsSpec.Frontend.Groups, nameAndReplicasForComponentGroup, frontendCloneSets, groupAndReadyReplicasForCloneSets),
		Compactor: buildGroupStatus(globalSpec.Replicas.Compactor, componentsSpec.Compactor.Groups, nameAndReplicasForComponentGroup, compactorCloneSets, groupAndReadyReplicasForCloneSets),
		Compute:   buildGroupStatus(globalSpec.Replicas.Compute, componentsSpec.Compute.Groups, nameAndReplicasForComputeGroup, computeStatefulSets, groupAndReadyReplicasForStatefulSet),
	}
	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStorage := &risingwave.Spec.Storages.Meta
		status.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorageStatus{
			Type: buildMetaStorageType(metaStorage),
		}

		// Report object storage status.
		objectStorage := &risingwave.Spec.Storages.Object
		status.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorageStatus{
			Type: buildObjectStorageType(objectStorage),
		}

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
	computeService *corev1.Service, compactorService *corev1.Service,
	metaStatefulSets []appsv1.StatefulSet, frontendDeployments []appsv1.Deployment,
	computeStatefulSets []appsv1.StatefulSet, compactorDeployments []appsv1.Deployment,
	configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	globalSpec := &risingwave.Spec.Global
	componentsSpec := &risingwave.Spec.Components

	// Update the replicas and storage status.
	nameAndReplicasForComponentGroup := func(g *risingwavev1alpha1.RisingWaveComponentGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	groupAndReadyReplicasForDeployments := func(t *appsv1.Deployment) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	nameAndReplicasForComputeGroup := func(g *risingwavev1alpha1.RisingWaveComputeGroup) (string, int32) {
		return g.Name, g.Replicas
	}
	groupAndReadyReplicasForStatefulSet := func(t *appsv1.StatefulSet) (string, int32) {
		return t.Labels[consts.LabelRisingWaveGroup], t.Status.ReadyReplicas
	}
	componentReplicas := risingwavev1alpha1.RisingWaveComponentsReplicasStatus{
		Meta:      buildGroupStatus(globalSpec.Replicas.Meta, componentsSpec.Meta.Groups, nameAndReplicasForComponentGroup, metaStatefulSets, groupAndReadyReplicasForStatefulSet),
		Frontend:  buildGroupStatus(globalSpec.Replicas.Frontend, componentsSpec.Frontend.Groups, nameAndReplicasForComponentGroup, frontendDeployments, groupAndReadyReplicasForDeployments),
		Compactor: buildGroupStatus(globalSpec.Replicas.Compactor, componentsSpec.Compactor.Groups, nameAndReplicasForComponentGroup, compactorDeployments, groupAndReadyReplicasForDeployments),
		Compute:   buildGroupStatus(globalSpec.Replicas.Compute, componentsSpec.Compute.Groups, nameAndReplicasForComputeGroup, computeStatefulSets, groupAndReadyReplicasForStatefulSet),
	}
	mgr.risingwaveManager.UpdateStatus(func(status *risingwavev1alpha1.RisingWaveStatus) {
		// Report meta storage status.
		metaStorage := &risingwave.Spec.Storages.Meta
		status.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorageStatus{
			Type: buildMetaStorageType(metaStorage),
		}

		// Report object storage status.
		objectStorage := &risingwave.Spec.Storages.Object
		status.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorageStatus{
			Type: buildObjectStorageType(objectStorage),
		}

		versions := strings.Split(mgr.risingwaveManager.RisingWave().Spec.Global.Image, ":")
		status.Version = versions[len(versions)-1]

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

func (mgr *risingWaveControllerManagerImpl) getPodTemplates(ctx context.Context, logger logr.Logger, templateNames []string) (map[string]risingwavev1alpha1.RisingWavePodTemplate, error) {
	podTemplates := make(map[string]risingwavev1alpha1.RisingWavePodTemplate)

	for _, templateName := range lo.Uniq(templateNames) {
		if templateName == "" {
			continue
		}

		var podTemplate risingwavev1alpha1.RisingWavePodTemplate

		if err := mgr.client.Get(ctx, types.NamespacedName{
			Namespace: mgr.risingwaveManager.RisingWave().Namespace,
			Name:      templateName,
		}, &podTemplate); err != nil {
			logger.Error(err, "Failed to get pod template", "template-name", templateName)
			return nil, err
		}

		podTemplates[templateName] = podTemplate
	}

	return podTemplates, nil
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
	groupPodTemplates map[string]string,
	objects []T,
	factory func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) TP,
) (reconcile.Result, error) {
	logger = logger.WithValues("component", component)

	// Build expected group set.
	expectedGroupSet := make(map[string]int)
	for group := range groupPodTemplates {
		expectedGroupSet[group] = 1
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
		// Build the pod templates.
		templateNames := make([]string, 0)
		for group, podTemplate := range groupPodTemplates {
			_, toSync := toSyncGroupObjects[group]
			if !toSync {
				continue
			}
			if podTemplate != "" {
				templateNames = append(templateNames, podTemplate)
			}
		}

		podTemplates, err := mgr.getPodTemplates(ctx, logger, templateNames)
		if err != nil {
			return ctrlkit.RequeueIfErrorAndWrap("unable to get pod templates", err)
		}

		for group, workloadObj := range toSyncGroupObjects {
			if err := syncObject(mgr, ctx, workloadObj, func() TP {
				return factory(group, podTemplates)
			}, logger.WithValues("group", group)); err != nil {
				return ctrlkit.RequeueIfErrorAndWrap("unable to sync object", err)
			}
		}
	}

	return ctrlkit.Continue()
}

func extractNameAndPodTemplateFromComponentGroup(g *risingwavev1alpha1.RisingWaveComponentGroup) (string, string) {
	podTemplate := ""
	if g.RisingWaveComponentGroupTemplate != nil && g.PodTemplate != nil {
		podTemplate = *g.PodTemplate
	}
	return g.Name, podTemplate
}

func extractNameAndPodTemplateFromComputeGroup(g *risingwavev1alpha1.RisingWaveComputeGroup) (string, string) {
	podTemplate := ""
	if g.RisingWaveComputeGroupTemplate != nil && g.PodTemplate != nil {
		podTemplate = *g.PodTemplate
	}
	return g.Name, podTemplate
}

func followPtrOrDefault[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func buildGroupPodTemplateMap[G any](groups []G, extract func(*G) (string, string)) map[string]string {
	r := make(map[string]string)
	for _, group := range groups {
		name, podTemplate := extract(&group)
		r[name] = podTemplate
	}
	return r
}

// SyncCompactorDeployments implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorDeployments(ctx context.Context, logger logr.Logger, compactorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	var groupPodTemplates = make(map[string]string)

	// We only want groupPodTemplates to be populated for deployments if open Kruise is disabled
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Compactor.Groups, extractNameAndPodTemplateFromComponentGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compactor > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}

	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		compactorDeployments,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
			return mgr.objectFactory.NewCompactorDeployment(group, podTemplates)
		},
	)
}

// SyncCompactorCloneSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorCloneSets(ctx context.Context, logger logr.Logger, compactorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	// We only want to populate groupPodTemplates for CloneSets if Open Kruise is enabled.
	var groupPodTemplates = make(map[string]string)
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Compactor.Groups, extractNameAndPodTemplateFromComponentGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compactor > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}
	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		compactorCloneSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1alpha1.CloneSet {
			return mgr.objectFactory.NewCompactorCloneSet(group, podTemplates)
		},
	)
}

// SyncComputeStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeStatefulSets(ctx context.Context, logger logr.Logger, computeStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	var groupPodTemplates = make(map[string]string)

	// We only want groupPodTemplates to be populated for statefulsets if open Kruise is disabled
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Compute.Groups, extractNameAndPodTemplateFromComputeGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compute > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}
	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		computeStatefulSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.StatefulSet {
			return mgr.objectFactory.NewComputeStatefulSet(group, podTemplates)
		},
	)
}

// SyncComputeAdvancedStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeAdvancedStatefulSets(ctx context.Context, logger logr.Logger, computeStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	var groupPodTemplates = make(map[string]string)

	// We only want groupPodTemplates to be populated for deployments if open Kruise is disabled
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Compute.Groups, extractNameAndPodTemplateFromComputeGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compute > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}

	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		computeStatefulSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1beta1.StatefulSet {
			return mgr.objectFactory.NewComputeAdvancedStatefulSet(group, podTemplates)
		},
	)
}

// SyncFrontendDeployments implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendDeployments(ctx context.Context, logger logr.Logger, frontendDeployments []appsv1.Deployment) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	var groupPodTemplates = make(map[string]string)

	// We only want groupPodTemplates to be populated for deployments if open Kruise is disabled
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Frontend.Groups, extractNameAndPodTemplateFromComponentGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Frontend > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}

	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		frontendDeployments,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
			return mgr.objectFactory.NewFrontendDeployment(group, podTemplates)
		},
	)
}

// SyncFrontendCloneSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendCloneSets(ctx context.Context, logger logr.Logger, frontendCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	// We only want to populate groupPodTemplates for CloneSets if Open Kruise is enabled.
	var groupPodTemplates = make(map[string]string)
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Frontend.Groups, extractNameAndPodTemplateFromComponentGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Frontend > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}
	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		frontendCloneSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1alpha1.CloneSet {
			return mgr.objectFactory.NewFrontendCloneSet(group, podTemplates)
		},
	)
}

// SyncMetaStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaStatefulSets(ctx context.Context, logger logr.Logger, metaStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()
	var groupPodTemplates = make(map[string]string)

	// We only want groupPodTemplates to be populated for deployments if open Kruise is disabled
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Meta.Groups, extractNameAndPodTemplateFromComponentGroup)
		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Meta > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}

	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		metaStatefulSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.StatefulSet {
			return mgr.objectFactory.NewMetaStatefulSet(group, podTemplates)
		},
	)
}

// SyncMetaAdvancedStatefulSets implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaAdvancedStatefulSets(ctx context.Context, logger logr.Logger, metaCloneSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	// We only want to populate groupPodTemplates for CloneSets if Open Kruise is enabled.
	var groupPodTemplates = make(map[string]string)

	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		groupPodTemplates = buildGroupPodTemplateMap(risingwave.Spec.Components.Meta.Groups, extractNameAndPodTemplateFromComponentGroup)

		// Enable the default group only if the global replicas > 0.
		if risingwave.Spec.Global.Replicas.Meta > 0 {
			groupPodTemplates[""] = followPtrOrDefault(risingwave.Spec.Global.PodTemplate)
		}
	}

	return syncComponentGroupWorkloads(
		mgr, ctx, logger,
		consts.ComponentCompactor,
		groupPodTemplates,
		metaCloneSets,
		func(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1beta1.StatefulSet {
			return mgr.objectFactory.NewMetaAdvancedStatefulSet(group, podTemplates)
		},
	)
}

func waitComponentGroupWorkloadsReady[T any, TP ptrAsObject[T]](ctx context.Context, logger logr.Logger, component string,
	groups map[string]int, objects []T, isReady func(*T) bool) (reconcile.Result, error) {
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

// WaitBeforeCompactorDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorDeploymentsReady(ctx context.Context, logger logr.Logger, compactorDeployments []appsv1.Deployment) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Compactor.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compactor > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentCompactor, groupMap,
		compactorDeployments,
		func(t *appsv1.Deployment) bool {
			return mgr.isObjectSynced(t) && utils.IsDeploymentRolledOut(t)
		},
	)
}

// WaitBeforeCompactorDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeCompactorCloneSetsReady(ctx context.Context, logger logr.Logger, compactorCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Compactor.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compactor > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentCompactor, groupMap,
		compactorCloneSets,
		func(t *kruiseappsv1alpha1.CloneSet) bool {
			return mgr.isObjectSynced(t) && utils.IsCloneSetRolledOut(t)
		},
	)
}

// WaitBeforeComputeStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeStatefulSetsReady(ctx context.Context, logger logr.Logger, computeStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Compute.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compute > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentCompute, groupMap,
		computeStatefulSets,
		func(t *appsv1.StatefulSet) bool {
			return mgr.isObjectSynced(t) && utils.IsStatefulSetRolledOut(t)
		},
	)
}

// WaitBeforeComputeStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeComputeAdvancedStatefulSetsReady(ctx context.Context, logger logr.Logger, computeStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Compute.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Compute > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentCompute, groupMap,
		computeStatefulSets,
		func(t *kruiseappsv1beta1.StatefulSet) bool {
			return mgr.isObjectSynced(t) && utils.IsAdvancedStatefulSetRolledOut(t)
		},
	)
}

// WaitBeforeFrontendDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendDeploymentsReady(ctx context.Context, logger logr.Logger, frontendDeployments []appsv1.Deployment) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Frontend.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Frontend > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentFrontend, groupMap,
		frontendDeployments,
		func(t *appsv1.Deployment) bool {
			return mgr.isObjectSynced(t) && utils.IsDeploymentRolledOut(t)
		},
	)
}

// WaitBeforeFrontendDeploymentsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeFrontendCloneSetsReady(ctx context.Context, logger logr.Logger, frontendCloneSets []kruiseappsv1alpha1.CloneSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)
	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Frontend.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Frontend > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentFrontend, groupMap,
		frontendCloneSets,
		func(t *kruiseappsv1alpha1.CloneSet) bool {
			return mgr.isObjectSynced(t) && utils.IsCloneSetRolledOut(t)
		},
	)
}

// WaitBeforeMetaStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaStatefulSetsReady(ctx context.Context, logger logr.Logger, metaStatefulSets []appsv1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)

	if !mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Meta.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Meta > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentMeta, groupMap,
		metaStatefulSets,
		func(t *appsv1.StatefulSet) bool {
			return mgr.isObjectSynced(t) && utils.IsStatefulSetRolledOut(t)
		},
	)
}

// WaitBeforeMetaAdvancedStatefulSetsReady implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) WaitBeforeMetaAdvancedStatefulSetsReady(ctx context.Context, logger logr.Logger, metaAdvancedStatefulSets []kruiseappsv1beta1.StatefulSet) (reconcile.Result, error) {
	risingwave := mgr.risingwaveManager.RisingWave()

	groupMap := make(map[string]int)

	if mgr.risingwaveManager.IsOpenKruiseEnabled() {
		for _, group := range risingwave.Spec.Components.Meta.Groups {
			groupMap[group.Name] = 1
		}

		// Enable the default group only if global replicas > 0.
		if risingwave.Spec.Global.Replicas.Meta > 0 {
			groupMap[""] = 1
		}
	}

	return waitComponentGroupWorkloadsReady(ctx, logger,
		consts.ComponentMeta, groupMap,
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
		return mgr.client.Create(ctx, newObj)
	} else {
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
			return mgr.client.Update(ctx, newObj)
		}
		return nil
	}
}

// Helper function for compile time type assertion.
func syncObject[T client.Object](mgr *risingWaveControllerManagerImpl, ctx context.Context, obj T, factory func() T, logger logr.Logger) error {
	return mgr.syncObject(ctx, obj, func() (client.Object, error) {
		return factory(), nil
	}, logger)
}

func syncObjectErr[T client.Object](mgr *risingWaveControllerManagerImpl, ctx context.Context, obj T, factory func() (T, error), logger logr.Logger) error {
	return mgr.syncObject(ctx, obj, func() (client.Object, error) {
		return factory()
	}, logger)
}

// SyncCompactorService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncCompactorService(ctx context.Context, logger logr.Logger, compactorService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, compactorService, mgr.objectFactory.NewCompactorService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compactor service", err)
}

// SyncComputeService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncComputeService(ctx context.Context, logger logr.Logger, computeService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, computeService, mgr.objectFactory.NewComputeService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync compute service", err)
}

// SyncFrontendService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncFrontendService(ctx context.Context, logger logr.Logger, frontendService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, frontendService, mgr.objectFactory.NewFrontendService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync frontend service", err)
}

// SyncMetaService implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncMetaService(ctx context.Context, logger logr.Logger, metaService *corev1.Service) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, metaService, mgr.objectFactory.NewMetaService, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync meta service", err)
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

// SyncConfigConfigMap implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncConfigConfigMap(ctx context.Context, logger logr.Logger, configConfigMap *corev1.ConfigMap) (reconcile.Result, error) {
	err := syncObjectErr(mgr, ctx, configConfigMap, func() (*corev1.ConfigMap, error) {
		configurationSpec := &mgr.risingwaveManager.RisingWave().Spec.Configuration
		if configurationSpec.ConfigMap == nil {
			return mgr.objectFactory.NewConfigConfigMap(""), nil
		} else {
			var cm corev1.ConfigMap
			err := mgr.client.Get(ctx, types.NamespacedName{
				Namespace: mgr.risingwaveManager.RisingWave().Namespace,
				Name:      configurationSpec.ConfigMap.Name,
			}, &cm)
			if client.IgnoreNotFound(err) != nil {
				return nil, fmt.Errorf("unable to get configmap %s: %w", configurationSpec.ConfigMap.Name, err)
			}
			val, ok := cm.Data[configurationSpec.ConfigMap.Key]
			if !ok && (configurationSpec.ConfigMap.Optional == nil || !*configurationSpec.ConfigMap.Optional) {
				return nil, fmt.Errorf("key not found in configmap")
			}
			return mgr.objectFactory.NewConfigConfigMap(val), nil
		}
	}, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync config configmap", err)
}

// SyncServiceMonitor implements RisingWaveControllerManagerImpl.
func (mgr *risingWaveControllerManagerImpl) SyncServiceMonitor(ctx context.Context, logger logr.Logger, serviceMonitor *monitoringv1.ServiceMonitor) (reconcile.Result, error) {
	err := syncObject(mgr, ctx, serviceMonitor, mgr.objectFactory.NewServiceMonitor, logger)
	return ctrlkit.RequeueIfErrorAndWrap("unable to sync service monitor", err)
}

func newRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager, messageStore *event.MessageStore) *risingWaveControllerManagerImpl {
	return &risingWaveControllerManagerImpl{
		client:            client,
		risingwaveManager: risingwaveManager,
		objectFactory:     factory.NewRisingWaveObjectFactory(risingwaveManager.RisingWave(), client.Scheme()),
		eventMessageStore: messageStore,
	}
}

// NewRisingWaveControllerManagerImpl creates an object that implements the RisingWaveControllerManagerImpl.
func NewRisingWaveControllerManagerImpl(client client.Client, risingwaveManager *object.RisingWaveManager, messageStore *event.MessageStore) RisingWaveControllerManagerImpl {
	return newRisingWaveControllerManagerImpl(client, risingwaveManager, messageStore)
}
