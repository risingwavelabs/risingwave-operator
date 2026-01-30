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
	"reflect"
	"strconv"
	"strings"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory/envs"

	"github.com/distribution/reference"
	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"

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
	envs.PodIP:           true,
	envs.PodName:         true,
	envs.PodNamespace:    true,
	envs.RustBacktrace:   true,
	envs.RWWorkerThreads: true,
	envs.JavaOpts:        true,
}

func (v *RisingWaveValidatingWebhook) isBypassed(obj client.Object) bool {
	val, ok := obj.GetAnnotations()[consts.AnnotationBypassValidatingWebhook]
	if !ok {
		return false
	}

	boolVal, _ := strconv.ParseBool(val)

	return boolVal
}

//nolint:gocognit
func (v *RisingWaveValidatingWebhook) validateNodeGroup(path *field.Path, nodeGroup *risingwavev1alpha1.RisingWaveNodeGroup, openKruiseEnabled bool) field.ErrorList {
	fieldErrs := field.ErrorList{}

	if nodeGroup == nil {
		return nil
	}

	if nodeGroup.Name != "" {
		if errs := validation.IsDNS1123Subdomain(nodeGroup.Name); len(errs) > 0 {
			fieldErrs = append(fieldErrs, field.Invalid(
				path.Child("name"),
				nodeGroup.Name,
				fmt.Sprintf("invalid node group name, must be a valid DNS-1123 subdomain: %s", strings.Join(errs, "; ")),
			))
		}
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

	//nolint:gocritic
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

		_, err := strconv.Atoi(strings.ReplaceAll(partitionVal, "%", ""))
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

func ptrValueNotZero[T any](ptr *T) bool {
	var zero T

	return ptr != nil && !reflect.DeepEqual(ptr, &zero)
}

func (v *RisingWaveValidatingWebhook) validateMetaStoreAndStateStore(path *field.Path, metaStore *risingwavev1alpha1.RisingWaveMetaStoreBackend, stateStore *risingwavev1alpha1.RisingWaveStateStoreBackend) field.ErrorList {
	fieldErrs := field.ErrorList{}

	isMetaMemory := ptrValueNotZero(metaStore.Memory)
	isMetaEtcd := ptrValueNotZero(metaStore.Etcd)
	isMetaSQLite := ptrValueNotZero(metaStore.SQLite)
	isMetaMySQL := ptrValueNotZero(metaStore.MySQL)
	isMetaPG := ptrValueNotZero(metaStore.PostgreSQL)

	validMetaStoreTypeCount := lo.CountBy([]bool{isMetaMemory, isMetaEtcd, isMetaSQLite,
		isMetaMySQL, isMetaPG}, func(x bool) bool { return x })
	if validMetaStoreTypeCount == 0 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("metaStore"), metaStore, "must configure the meta store"))
	} else if validMetaStoreTypeCount > 1 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("metaStore"), metaStore, "multiple meta store types"))
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
	isStateHuaweiCloudOBS := stateStore.HuaweiCloudOBS != nil

	if isStateS3 {
		if len(stateStore.S3.Endpoint) > 0 {
			// S3-compatible mode, secretName is required.
			if stateStore.S3.SecretName == "" {
				fieldErrs = append(fieldErrs, field.Required(path.Child("stateStore", "s3", "credentials", "secretName"), "secretName is required"))
			}
		} else {
			// AWS S3.
			if !ptr.Deref(stateStore.S3.UseServiceAccount, false) && stateStore.S3.SecretName == "" {
				fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "s3", "credentials"), stateStore.S3.SecretName, "either secretName or useServiceAccount must be specified"))
			}
		}
	}

	if isStateAzureBlob {
		if !ptr.Deref(stateStore.AzureBlob.UseServiceAccount, false) &&
			stateStore.AzureBlob.SecretName == "" {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "azureBlob", "credentials"), stateStore.S3.SecretName, "either secretName or useServiceAccount must be specified"))
		}
	}

	if isStateGCS {
		if !ptr.Deref(stateStore.GCS.UseWorkloadIdentity, false) && (stateStore.GCS.SecretName == "") {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "gcs", "credentials"), stateStore.GCS.SecretName, "either secretName or useWorkloadIdentity must be specified"))
		}
	}

	validStateStoreTypeCount := lo.CountBy([]bool{isStateMemory, isStateMinIO, isStateS3, isStateGCS, isStateAliyunOSS,
		isStateAzureBlob, isStateHDFS, isStateWebHDFS, isStateLocalDisk, isStateHuaweiCloudOBS}, func(x bool) bool { return x })
	if validStateStoreTypeCount == 0 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore"), stateStore, "must configure the state store"))
	} else if validStateStoreTypeCount > 1 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore"), stateStore, "multiple state store types"))
	}

	if strings.HasSuffix(stateStore.DataDirectory, "/") ||
		strings.HasPrefix(stateStore.DataDirectory, "/") ||
		strings.Contains(stateStore.DataDirectory, "//") ||
		len(stateStore.DataDirectory) > 800 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("stateStore", "dataDirectory"), stateStore.DataDirectory, "must be a valid path"))
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

	computeGroupsPath := path.Child("compute", "nodeGroups")
	for i, ng := range components.Compute.NodeGroups {
		fieldErrs = append(fieldErrs, v.validateNodeGroup(computeGroupsPath.Index(i), &ng, openKruiseEnabled)...)
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateMetaReplicas(obj *risingwavev1alpha1.RisingWave) field.ErrorList {
	// When the meta storage isn't memory, there's no limitation on the replicas.
	if !ptr.Deref(obj.Spec.MetaStore.Memory, false) {
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

func (v *RisingWaveValidatingWebhook) validateSecretStore(obj *risingwavev1alpha1.RisingWave) field.ErrorList {
	fieldErrs := field.ErrorList{}
	if obj.Spec.SecretStore.PrivateKey.Value != nil && obj.Spec.SecretStore.PrivateKey.SecretRef != nil {
		fieldErrs = append(fieldErrs, field.Forbidden(field.NewPath("spec", "secretStore", "privateKey"), "both value and secretRef are set"))
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	gvk := obj.GroupVersionKind()

	fieldErrs := field.ErrorList{}

	// Validate the image.
	if !isImageValid(obj.Spec.Image) {
		fieldErrs = append(fieldErrs, field.Invalid(field.NewPath("spec", "image"), obj.Spec.Image, "invalid image reference"))
	}

	// Validate the additional frontend service metadata.
	for label := range obj.Spec.AdditionalFrontendServiceMetadata.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(field.NewPath("spec", "additionalFrontendServiceMetadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	// Validate the additional meta service metadata.
	for label := range obj.Spec.AdditionalMetaServiceMetadata.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(field.NewPath("spec", "additionalMetaServiceMetadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	// Validate to make sure open kruise cannot be set to true when it is disabled at operator level.
	if !v.openKruiseAvailable && ptr.Deref(obj.Spec.EnableOpenKruise, false) {
		fieldErrs = append(fieldErrs, field.Forbidden(field.NewPath("spec", "enableOpenKruise"), "OpenKruise is disabled."))
	}

	// Validate the meta store and state store spec.
	fieldErrs = append(fieldErrs, v.validateMetaStoreAndStateStore(field.NewPath("spec"), &obj.Spec.MetaStore, &obj.Spec.StateStore)...)

	// Validate the configuration spec.
	fieldErrs = append(fieldErrs, v.validateConfiguration(field.NewPath("spec", "configuration"), &obj.Spec.Configuration)...)

	// Validate the components spec.
	//   * If the global image is empty, then the image of all groups must not be empty.
	fieldErrs = append(fieldErrs, v.validateComponents(
		field.NewPath("spec", "components"),
		&obj.Spec.Components,
		v.openKruiseAvailable && ptr.Deref(obj.Spec.EnableOpenKruise, false),
	)...)

	// Validate the meta replicas.
	fieldErrs = append(fieldErrs, v.validateMetaReplicas(obj)...)

	// Validate the secret store.
	fieldErrs = append(fieldErrs, v.validateSecretStore(obj)...)

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, fieldErrs)
	}

	return nil
}

// ValidateCreate implements admission.Validator.
func (v *RisingWaveValidatingWebhook) ValidateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWave) (warnings admission.Warnings, err error) {
	if v.isBypassed(obj) {
		return nil, nil
	}

	err = v.validateCreate(ctx, obj)

	return
}

// ValidateDelete implements admission.Validator.
func (v *RisingWaveValidatingWebhook) ValidateDelete(ctx context.Context, obj *risingwavev1alpha1.RisingWave) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (v *RisingWaveValidatingWebhook) isMetaStoresTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.MetaStore, newObj.Spec.MetaStore)
}

func (v *RisingWaveValidatingWebhook) isStateStoresTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.StateStore, newObj.Spec.StateStore)
}

func pathForGroupReplicas(obj *risingwavev1alpha1.RisingWave, component, group string) *field.Path {
	index, _ := scaleview.NewRisingWaveScaleViewHelper(obj, component).GetGroupIndex(group)

	return field.NewPath("spec", "components", component, "nodeGroups").Index(index).Child("replicas")
}

func (v *RisingWaveValidatingWebhook) isSecretStoreChangeAllowed(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	oldStore, newStore := oldObj.Spec.SecretStore, newObj.Spec.SecretStore

	isPrivateKeySet := func(store *risingwavev1alpha1.RisingWaveSecretStore) bool {
		return store.PrivateKey.Value != nil || store.PrivateKey.SecretRef != nil
	}

	// Not set to set is allowed.
	if !isPrivateKeySet(&oldStore) {
		return true
	}

	// Changes on the private key are not allowed.
	oldPk, newPk := oldStore.PrivateKey, newStore.PrivateKey

	return equality.Semantic.DeepEqual(oldPk, newPk)
}

func (v *RisingWaveValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWave) error {
	gvk := oldObj.GroupVersionKind()

	// The meta store and state store must be kept consistent.
	if !v.isMetaStoresTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "metaStore"), "meta store must be kept consistent"),
		)
	}

	if !v.isStateStoresTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "stateStore"), "state store must be kept consistent"),
		)
	}

	if !v.isSecretStoreChangeAllowed(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "secretStore"), "secret store must be kept consistent"),
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
				switch cur {
				case lock.Replicas:
					updateCnt++
				case old:
					unchangedCnt++
				default:
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

// ValidateUpdate implements admission.Validator.
func (v *RisingWaveValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWave) (warnings admission.Warnings, err error) {
	if v.isBypassed(newObj) {
		return nil, nil
	}

	// Validate the new object first.
	if warnings, err := v.ValidateCreate(ctx, newObj); err != nil {
		return warnings, err
	}

	err = v.validateUpdate(ctx, oldObj, newObj)

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
func NewRisingWaveValidatingWebhook(openKruiseAvailable bool) admission.Validator[*risingwavev1alpha1.RisingWave] {
	return metrics.NewValidatingWebhookMetricsRecorder(&RisingWaveValidatingWebhook{openKruiseAvailable: openKruiseAvailable})
}
