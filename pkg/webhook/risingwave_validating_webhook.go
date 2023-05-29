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

package webhook

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/risingwavelabs/risingwave-operator/pkg/factory/envs"

	"github.com/distribution/distribution/reference"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	"github.com/risingwavelabs/risingwave-operator/pkg/scaleview"
)

// RisingWaveValidatingWebhook is the validating webhook for RisingWaves.
type RisingWaveValidatingWebhook struct {
	openKruiseAvailable bool
}

func isImageValid(image string) bool {
	return reference.ReferenceRegexp.MatchString(image)
}

var systemEnv = map[string]bool{
	envs.PodIP:                  true,
	envs.PodName:                true,
	envs.RustBacktrace:          true,
	envs.RWWorkerThreads:        true,
	envs.RWConnectorRPCEndPoint: true,
	envs.JavaOpts:               true,
}

func (v *RisingWaveValidatingWebhook) validateNodeGroup(path *field.Path, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, openKruiseEnabled bool) field.ErrorList {
	fieldErrs := field.ErrorList{}

	if nodeGroup == nil {
		return nil
	}

	// Validate the image.
	if nodeGroup.Template.Spec.Image != "" && !isImageValid(nodeGroup.Template.Spec.Image) {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("template", "spec", "image"), nodeGroup.Template.Spec.Image, "invalid image reference"))
	}

	// Validate the upgrade strategy.
	if nodeGroup.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate {
		if nodeGroup.UpgradeStrategy.RollingUpdate != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "rollingUpdate"), "must be nil when type is Recreate"))
		}
	}

	// Validate upgrade strategy type when open kruise is not enabled.
	if !openKruiseEnabled {
		if nodeGroup.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly ||
			nodeGroup.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("upgradeStrategy", "type"), nodeGroup.UpgradeStrategy.Type, "invalid upgrade strategy type"))
		}
		if nodeGroup.UpgradeStrategy.InPlaceUpdateStrategy != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "inPlaceUpdateStrategy"), "not allowed"))
		}
	} else {
		if nodeGroup.UpgradeStrategy.Type != risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly &&
			nodeGroup.UpgradeStrategy.Type != risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible &&
			nodeGroup.UpgradeStrategy.InPlaceUpdateStrategy != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "inPlaceUpdateStrategy"), "not allowed"))
		}
	}

	// Validate the partition value if it exists.
	if nodeGroupPartitionExistAndIsString(nodeGroup) {
		partitionVal := nodeGroup.UpgradeStrategy.RollingUpdate.Partition.StrVal
		_, err := strconv.Atoi(strings.Replace(partitionVal, "%", "", -1))
		if err != nil {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("upgradeStrategy", "rollingUpdate", "partition"),
					nodeGroup.UpgradeStrategy.RollingUpdate.Partition,
					"percentage/string unable to be converted to an integer value"))
		}
	}

	// Validate labels of the RisingWave's Pods
	for label := range nodeGroup.Template.ObjectMeta.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("template", "metadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	// Validate env of the RisingWave's Pods
	for i, v := range nodeGroup.Template.Spec.Env {
		if systemEnv[v.Name] {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("template", "spec", "env").Index(i).Child("name"), v.Name, fmt.Sprintf("Env with the name %s is system reserved", v.Name)))
		}
	}

	// Validate the resources only when limits exist
	if nodeGroup.Template.Spec.Resources.Limits != nil {
		// Validate the cpu resources.
		if _, ok := nodeGroup.Template.Spec.Resources.Limits[corev1.ResourceCPU]; ok &&
			nodeGroup.Template.Spec.Resources.Limits.Cpu().Cmp(*nodeGroup.Template.Spec.Resources.Requests.Cpu()) == -1 {
			fieldErrs = append(fieldErrs, field.Required(path.Child("template", "spec", "resources", "cpu"), "insufficient cpu resource"))
		}

		// Validate the memory resources.
		if _, ok := nodeGroup.Template.Spec.Resources.Limits[corev1.ResourceMemory]; ok &&
			nodeGroup.Template.Spec.Resources.Limits.Memory().Cmp(*nodeGroup.Template.Spec.Resources.Requests.Memory()) == -1 {
			fieldErrs = append(fieldErrs, field.Required(path.Child("template", "spec", "resources", "memory"), "insufficient memory resource"))
		}
	}

	return fieldErrs
}

func ptrValueNotZero[T comparable](ptr *T) bool {
	var zero T
	return ptr != nil && *ptr != zero
}

func (v *RisingWaveValidatingWebhook) validateMetaStoreAndStateStore(path *field.Path, metaStore *risingwavev1alpha1.RisingWaveMetaStoreBackend, stateStore *risingwavev1alpha1.RisingWaveStateStoreBackend) field.ErrorList {
	fieldErrs := field.ErrorList{}

	isMetaMemory, isMetaEtcd := ptrValueNotZero(metaStore.Memory), ptrValueNotZero(metaStore.Etcd)
	if isMetaMemory {
		if isMetaEtcd {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("metaStore", "etcd"), "must not specified when type is memory"))
		}
	} else {
		if !isMetaEtcd {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("metaStore"), metaStore, "either memory or etcd must be specified"))
		}
	}

	isStateMemory := ptrValueNotZero(stateStore.Memory)
	isStateMinIO := stateStore.MinIO != nil
	isStateS3 := stateStore.S3 != nil
	isStateGCS := stateStore.GCS != nil
	isStateAliyunOSS := stateStore.AliyunOSS != nil
	isStateAzureBlob := stateStore.AzureBlob != nil
	isStateHDFS := stateStore.HDFS != nil
	isStateWebHDFS := stateStore.WebHDFS != nil
	isStateLocalDisk := stateStore.LocalDisk != nil

	if isStateS3 {
		if len(stateStore.S3.Endpoint) > 0 {
			// S3-compatible mode, secretName is required.
			if stateStore.S3.RisingWaveS3Credentials.SecretName == "" {
				fieldErrs = append(fieldErrs, field.Required(path.Child("stateStore", "s3", "credentials", "secretName"), "secretName is required"))
			}
		} else {
			// AWS S3.
			if !pointer.BoolDeref(stateStore.S3.UseServiceAccount, false) && stateStore.S3.RisingWaveS3Credentials.SecretName == "" {
				fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "s3", "credentials"), stateStore.S3.SecretName, "either secretName or useServiceAccount must be specified"))
			}
		}
	}

	if isStateGCS {
		if !pointer.BoolDeref(stateStore.GCS.UseWorkloadIdentity, false) && (stateStore.GCS.RisingWaveGCSCredentials.SecretName == "") {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "gcs", "credentials"), stateStore.GCS.RisingWaveGCSCredentials.SecretName, "either secretName or useWorkloadIdentity must be specified"))
		}
	}

	validStateStoreTypeCount := lo.CountBy([]bool{isStateMemory, isStateMinIO, isStateS3, isStateGCS, isStateAliyunOSS, isStateAzureBlob, isStateHDFS, isStateWebHDFS, isStateLocalDisk}, func(x bool) bool { return x })
	if validStateStoreTypeCount == 0 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore"), stateStore, "must configure the state store"))
	} else if validStateStoreTypeCount > 1 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore"), stateStore, "multiple state store types"))
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateConfiguration(path *field.Path, configuration *risingwavev1alpha1.RisingWaveConfigurationSpec) field.ErrorList {
	if configuration.ConfigMap != nil {
		if configuration.ConfigMap.Name == "" {
			return field.ErrorList{
				field.Required(path.Child("configmap", "name"), "must be specified"),
			}
		} else if configuration.ConfigMap.Key == "" {
			return field.ErrorList{
				field.Required(path.Child("configmap", "key"), "must be specified"),
			}
		}
	}
	return nil
}

func (v *RisingWaveValidatingWebhook) validateComponents(path *field.Path, components *risingwavev1alpha1.RisingWaveComponentsSpec, openKruiseEnabled bool) field.ErrorList {
	fieldErrs := field.ErrorList{}

	metaGroupsPath := path.Child("meta", "nodeGroups")
	for i, ng := range components.Meta.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(metaGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	frontendGroupsPath := path.Child("frontend", "nodeGroups")
	for i, ng := range components.Frontend.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(frontendGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	compactorGroupsPath := path.Child("compactor", "nodeGroups")
	for i, ng := range components.Compactor.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(compactorGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	connectorGroupsPath := path.Child("connector", "nodeGroups")
	for i, ng := range components.Connector.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(connectorGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	computeGroupsPath := path.Child("compute", "nodeGroups")
	for i, ng := range components.Compute.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(computeGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateMetaReplicas(obj *risingwavev1alpha1.RisingWave) field.ErrorList {
	// When the meta storage isn't memory, there's no limitation on the replicas.
	if !pointer.BoolDeref(obj.Spec.MetaStore.Memory, false) {
		return nil
	}

	fieldErrs := field.ErrorList{}

	metaReplicas := int32(0)
	for _, ng := range obj.Spec.Components.Meta.NodeGroups {
		metaReplicas += ng.Replicas
	}

	if metaReplicas > 1 {
		fieldErrs = append(fieldErrs, field.Forbidden(field.NewPath("spec", "components", "meta"), "meta with replicas over 1 isn't allowed"))
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	gvk := obj.GroupVersionKind()

	fieldErrs := field.ErrorList{}

	// Validate the image.
	if !isImageValid(obj.Spec.Image) {
		fieldErrs = append(fieldErrs, field.Invalid(field.NewPath("template", "spec", "image"), obj.Spec.Image, "invalid image reference"))
	}

	// Validate the additional frontend service metadata.
	for label := range obj.Spec.AdditionalFrontendServiceMetadata.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(field.NewPath("spec", "additionalFrontendServiceMetadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	// Validate to make sure open kruise cannot be set to true when it is disabled at operator level.
	if !v.openKruiseAvailable && pointer.BoolDeref(obj.Spec.EnableOpenKruise, false) {
		fieldErrs = append(fieldErrs, field.Forbidden(field.NewPath("spec", "enableOpenKruise"), "OpenKruise is disabled."))
	}

	// Validate the storages spec.
	fieldErrs = append(fieldErrs, v.validateMetaStoreAndStateStore(field.NewPath("spec", "storages"), &obj.Spec.MetaStore, &obj.Spec.StateStore)...)

	// Validate the configuration spec.
	fieldErrs = append(fieldErrs, v.validateConfiguration(field.NewPath("spec", "configuration"), &obj.Spec.Configuration)...)

	// Validate the components spec.
	//   * If the global image is empty, then the image of all groups must not be empty.
	fieldErrs = append(fieldErrs, v.validateComponents(
		field.NewPath("spec", "components"),
		&obj.Spec.Components,
		v.openKruiseAvailable && pointer.BoolDeref(obj.Spec.EnableOpenKruise, false),
	)...)

	// Validate the meta replicas.
	fieldErrs = append(fieldErrs, v.validateMetaReplicas(obj)...)

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, fieldErrs)
	}
	return nil
}

// ValidateCreate implements admission.CustomValidator.
func (v *RisingWaveValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	err = v.validateCreate(ctx, obj.(*risingwavev1alpha1.RisingWave))
	return
}

// ValidateDelete implements admission.CustomValidator.
func (v *RisingWaveValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (v *RisingWaveValidatingWebhook) isMetaStoresTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.MetaStore, newObj.Spec.MetaStore)
}

func (v *RisingWaveValidatingWebhook) isConvertFromMeta(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.Storages.Meta, newObj.Spec.MetaStore)
}

func (v *RisingWaveValidatingWebhook) isStateStoresTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.StateStore, newObj.Spec.StateStore)
}

func (v *RisingWaveValidatingWebhook) isConvertFromObject(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	oldStateStore := &oldObj.Spec.Storages.Object
	newStateStore := &newObj.Spec.StateStore

	if newStateStore.MinIO != nil && oldStateStore.MinIO != nil {
		return newStateStore.MinIO.RisingWaveMinIOCredentials.SecretName == oldStateStore.MinIO.Secret
	} else if newStateStore.S3 != nil && oldStateStore.S3 != nil {
		return newStateStore.S3.RisingWaveS3Credentials.SecretName == oldStateStore.S3.Secret
	} else if newStateStore.GCS != nil && oldStateStore.GCS != nil {
		return newStateStore.GCS.RisingWaveGCSCredentials.SecretName == oldStateStore.GCS.Secret
	} else if newStateStore.AliyunOSS != nil && oldStateStore.AliyunOSS != nil {
		return newStateStore.AliyunOSS.RisingWaveS3Credentials.SecretName == oldStateStore.AliyunOSS.Secret
	} else if newStateStore.AzureBlob != nil && oldStateStore.AzureBlob != nil {
		return newStateStore.AzureBlob.RisingWaveAzureBlobCredentials.SecretName == oldStateStore.AzureBlob.Secret
	} else {
		return equality.Semantic.DeepEqual(oldStateStore, newStateStore)
	}
}

func pathForGroupReplicas(obj *risingwavev1alpha1.RisingWave, component, group string) *field.Path {
	index, _ := scaleview.NewRisingWaveScaleViewHelper(obj, component).GetGroupIndex(group)
	return field.NewPath("spec", "components", component, "nodeGroups").Index(index).Child("replicas")
}

func (v *RisingWaveValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWave) error {
	gvk := oldObj.GroupVersionKind()

	// The storages must not be changed, especially meta and state.
	if !v.isMetaStoresTheSame(oldObj, newObj) && !v.isConvertFromMeta(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "meta"), "meta storage must be kept consistent"),
		)
	}

	if !v.isStateStoresTheSame(oldObj, newObj) && !v.isConvertFromObject(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "state"), "state storage must be kept consistent"),
		)
	}

	fieldErrs := field.ErrorList{}

	// Validate the locks from scale views.
	for _, scaleView := range newObj.Status.ScaleViews {
		oldHelper := scaleview.NewRisingWaveScaleViewHelper(oldObj, scaleView.Component)
		newHelper := scaleview.NewRisingWaveScaleViewHelper(newObj, scaleView.Component)

		unchangedCnt, updateCnt := 0, 0
		for _, lock := range scaleView.GroupLocks {
			// Ignore the existence of the old group and allow adding locked (but not exist) groups.
			old, _ := oldHelper.ReadReplicas(lock.Name)

			if cur, ok := newHelper.ReadReplicas(lock.Name); !ok {
				fieldErrs = append(fieldErrs, field.Forbidden(
					pathForGroupReplicas(oldObj, scaleView.Component, lock.Name),
					"group is locked (delete)",
				))
			} else {
				// Either
				if cur == lock.Replicas {
					updateCnt++
				} else if cur == old {
					unchangedCnt++
				} else {
					updateCnt++
				}

				if (unchangedCnt > 0 && cur != old) || (updateCnt > 0 && cur != lock.Replicas) {
					fieldErrs = append(fieldErrs, field.Forbidden(
						pathForGroupReplicas(newObj, scaleView.Component, lock.Name),
						"group is locked (update)",
					))
				}
			}
		}
	}

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), oldObj.Name, fieldErrs)
	}

	return nil
}

// ValidateUpdate implements admission.CustomValidator.
func (v *RisingWaveValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) (warnings admission.Warnings, err error) {
	// Validate the new object first.
	if warnings, err := v.ValidateCreate(ctx, newObj); err != nil {
		return warnings, err
	}

	err = v.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWave), newObj.(*risingwavev1alpha1.RisingWave))
	return
}

func nodeGroupPartitionExistAndIsString(nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup) bool {
	if nodeGroup.UpgradeStrategy.RollingUpdate == nil {
		return false
	}
	if nodeGroup.UpgradeStrategy.RollingUpdate.Partition == nil {
		return false
	}
	return nodeGroup.UpgradeStrategy.RollingUpdate.Partition.Type == intstr.String
}

// NewRisingWaveValidatingWebhook returns a new validator for the RisingWave. The behavior differs on different values of the
// openKruiseAvailable.
func NewRisingWaveValidatingWebhook(openKruiseAvailable bool) webhook.CustomValidator {
	return metrics.NewValidatingWebhookMetricsRecorder(&RisingWaveValidatingWebhook{openKruiseAvailable: openKruiseAvailable})
}
