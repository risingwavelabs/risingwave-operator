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

package webhook

import (
	"context"
	"strconv"
	"strings"

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

func (v *RisingWaveValidatingWebhook) validateGroupTemplate(path *field.Path, groupTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate, isOpenKruiseEnabled bool) field.ErrorList {
	fieldErrs := field.ErrorList{}

	if groupTemplate == nil {
		return fieldErrs
	}

	// Validate the image.
	if groupTemplate.Image != "" && !isImageValid(groupTemplate.Image) {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("image"), groupTemplate.Image, "invalid image reference"))
	}

	// Validate the upgrade strategy.
	if groupTemplate.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate {
		if groupTemplate.UpgradeStrategy.RollingUpdate != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "rollingUpdate"), "must be nil when type is Recreate"))
		}
	}

	// Validate upgrade strategy type when open kruise is not enabled
	if !isOpenKruiseEnabled {
		if groupTemplate.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly ||
			groupTemplate.UpgradeStrategy.Type == risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("upgradeStrategy", "type"), groupTemplate.UpgradeStrategy.Type, "invalid upgrade strategy type"))
		}
		if groupTemplate.UpgradeStrategy.InPlaceUpdateStrategy != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "inPlaceUpdateStrategy"), "not allowed"))
		}
	} else {
		if groupTemplate.UpgradeStrategy.Type != risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly &&
			groupTemplate.UpgradeStrategy.Type != risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible &&
			groupTemplate.UpgradeStrategy.InPlaceUpdateStrategy != nil {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("upgradeStrategy", "inPlaceUpdateStrategy"), "not allowed"))
		}
	}

	// Validate the partition value if it exists.
	if groupTemplatePartitionExistAndIsString(groupTemplate) {
		partitionVal := groupTemplate.UpgradeStrategy.RollingUpdate.Partition.StrVal
		_, err := strconv.Atoi(strings.Replace(partitionVal, "%", "", -1))
		if err != nil {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("upgradeStrategy", "rollingUpdate", "partition"),
					groupTemplate.UpgradeStrategy.RollingUpdate.Partition,
					"percentage/string unable to be converted to an integer value"))
		}
	}

	// Validate labels of the RisingWave's Pods
	for label := range groupTemplate.Metadata.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("Metadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	// Validate the resources only when limits exist
	if groupTemplate.Resources.Limits == nil {
		return fieldErrs
	}

	// Validate the cpu resources.
	if _, ok := groupTemplate.Resources.Limits[corev1.ResourceCPU]; ok &&
		groupTemplate.Resources.Limits.Cpu().Cmp(*groupTemplate.Resources.Requests.Cpu()) == -1 {
		fieldErrs = append(fieldErrs, field.Required(path.Child("resources", "cpu"), "insufficient cpu resource"))
	}

	// Validate the memory resources.
	if _, ok := groupTemplate.Resources.Limits[corev1.ResourceMemory]; ok &&
		groupTemplate.Resources.Limits.Memory().Cmp(*groupTemplate.Resources.Requests.Memory()) == -1 {
		fieldErrs = append(fieldErrs, field.Required(path.Child("resources", "memory"), "insufficient memory resource"))
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateGlobal(path *field.Path, global *risingwavev1alpha1.RisingWaveGlobalSpec, isOpenKruiseEnabled bool) field.ErrorList {
	fieldErrs := v.validateGroupTemplate(path, &global.RisingWaveComponentGroupTemplate, isOpenKruiseEnabled)

	if global.Replicas.Meta > 0 || global.Replicas.Frontend > 0 ||
		global.Replicas.Compute > 0 || global.Replicas.Compactor > 0 {
		if global.Image == "" {
			fieldErrs = append(fieldErrs, field.Required(path.Child("image"), "must be specified when there're global replicas"))
		}
	}

	// Validate labels of the RisingWave's service metadata
	for label := range global.ServiceMeta.Labels {
		if strings.HasPrefix(label, "risingwave/") {
			fieldErrs = append(fieldErrs,
				field.Invalid(path.Child("serviceMetadata", "labels"), label, "Labels with the prefix 'risingwave/' are system reserved"))
		}
	}

	return fieldErrs
}

func ptrValueNotZero[T comparable](ptr *T) bool {
	var zero T
	return ptr != nil && *ptr != zero
}

func (v *RisingWaveValidatingWebhook) validateStorages(path *field.Path, storages *risingwavev1alpha1.RisingWaveStoragesSpec) field.ErrorList {
	fieldErrs := field.ErrorList{}

	isMetaMemory, isMetaEtcd := ptrValueNotZero(storages.Meta.Memory), ptrValueNotZero(storages.Meta.Etcd)
	if isMetaMemory {
		if isMetaEtcd {
			fieldErrs = append(fieldErrs, field.Forbidden(path.Child("meta", "etcd"), "must not specified when type is memory"))
		}
	} else {
		if !isMetaEtcd {
			fieldErrs = append(fieldErrs, field.Invalid(path.Child("meta"), storages.Meta, "either memory or etcd must be specified"))
		}
	}

	isObjectMemory := ptrValueNotZero(storages.Object.Memory)
	isObjectMinIO := storages.Object.MinIO != nil
	isObjectS3 := storages.Object.S3 != nil
	isObjectAliyunOSS := storages.Object.AliyunOSS != nil
	validObjectStorageTypeCount := lo.CountBy([]bool{isObjectMemory, isObjectMinIO, isObjectS3, isObjectAliyunOSS}, func(x bool) bool { return x })
	if validObjectStorageTypeCount == 0 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("object"), storages.Object, "must configure the object storage"))
	} else if validObjectStorageTypeCount > 1 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("object"), storages.Object, "multiple object storage types"))
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateSecurity(path *field.Path, security *risingwavev1alpha1.RisingWaveSecuritySpec) field.ErrorList {
	return nil
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

func (v *RisingWaveValidatingWebhook) validateComponents(path *field.Path, components *risingwavev1alpha1.RisingWaveComponentsSpec, storages *risingwavev1alpha1.RisingWaveStoragesSpec, globalImageProvided bool, openKruiseEnabled bool) field.ErrorList {
	fieldErrs := field.ErrorList{}

	metaGroupsPath := path.Child("meta", "groups")
	for i, group := range components.Meta.Groups {
		fieldErrs = append(fieldErrs, v.validateGroupTemplate(metaGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate, openKruiseEnabled)...)
		if !globalImageProvided && (group.RisingWaveComponentGroupTemplate == nil || group.Image == "") {
			fieldErrs = append(fieldErrs, field.Required(metaGroupsPath.Index(i).Child("image"), "must be specified when there's no global image"))
		}
	}

	frontendGroupsPath := path.Child("frontend", "groups")
	for i, group := range components.Frontend.Groups {
		fieldErrs = append(fieldErrs, v.validateGroupTemplate(frontendGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate, openKruiseEnabled)...)
		if !globalImageProvided && (group.RisingWaveComponentGroupTemplate == nil || group.Image == "") {
			fieldErrs = append(fieldErrs, field.Required(frontendGroupsPath.Index(i).Child("image"), "must be specified when there's no global image"))
		}
	}

	compactorGroupsPath := path.Child("compactor", "groups")
	for i, group := range components.Compactor.Groups {
		fieldErrs = append(fieldErrs, v.validateGroupTemplate(compactorGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate, openKruiseEnabled)...)
		if !globalImageProvided && (group.RisingWaveComponentGroupTemplate == nil || group.Image == "") {
			fieldErrs = append(fieldErrs, field.Required(compactorGroupsPath.Index(i).Child("image"), "must be specified when there's no global image"))
		}
	}

	pvClaims := make(map[string]int)
	for _, pvc := range storages.PVCTemplates {
		pvClaims[pvc.Name] = 1
	}

	computeGroupsPath := path.Child("compute", "groups")
	for i, group := range components.Compute.Groups {
		if !globalImageProvided && (group.RisingWaveComputeGroupTemplate == nil || group.Image == "") {
			fieldErrs = append(fieldErrs, field.Required(computeGroupsPath.Index(i).Child("image"), "must be specified when there's no global image"))
		}

		if group.RisingWaveComputeGroupTemplate != nil {
			fieldErrs = append(fieldErrs, v.validateGroupTemplate(computeGroupsPath.Index(i), &group.RisingWaveComponentGroupTemplate, openKruiseEnabled)...)

			for vi, volumeMount := range group.VolumeMounts {
				if _, pvcExists := pvClaims[volumeMount.Name]; !pvcExists {
					fieldErrs = append(fieldErrs, field.Invalid(
						computeGroupsPath.Index(i).Child("volumeMounts").Index(vi).Child("name"),
						volumeMount.Name,
						"volume not declared in pvcTemplates",
					))
				}
			}
		}
	}

	return fieldErrs
}

func (v *RisingWaveValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	gvk := obj.GroupVersionKind()

	fieldErrs := field.ErrorList{}

	// Validate to make sure open kruise cannot be set to true when it is disabled at operator level.
	if !v.openKruiseAvailable && pointer.BoolDeref(obj.Spec.EnableOpenKruise, false) {
		fieldErrs = append(fieldErrs, field.Forbidden(field.NewPath("spec", "enableOpenKruise"), "OpenKruise is disabled."))
	}

	// Validate the global spec.
	//   * If global replicas of any component is larger than 1, then the image in global must not be empty.
	fieldErrs = append(fieldErrs, v.validateGlobal(field.NewPath("spec", "global"), &obj.Spec.Global, pointer.BoolDeref(obj.Spec.EnableOpenKruise, false))...)

	// Validate the storages spec.
	fieldErrs = append(fieldErrs, v.validateStorages(field.NewPath("spec", "storages"), &obj.Spec.Storages)...)

	// Validate the security spec.
	fieldErrs = append(fieldErrs, v.validateSecurity(field.NewPath("spec", "security"), &obj.Spec.Security)...)

	// Validate the configuration spec.
	fieldErrs = append(fieldErrs, v.validateConfiguration(field.NewPath("spec", "configuration"), &obj.Spec.Configuration)...)

	// Validate the components spec.
	//   * If the global image is empty, then the image of all groups must not be empty.
	fieldErrs = append(fieldErrs, v.validateComponents(
		field.NewPath("spec", "components"),
		&obj.Spec.Components,
		&obj.Spec.Storages,
		obj.Spec.Global.Image != "",
		v.openKruiseAvailable && pointer.BoolDeref(obj.Spec.EnableOpenKruise, false),
	)...)

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, fieldErrs)
	}
	return nil
}

// ValidateCreate implements admission.CustomValidator.
func (v *RisingWaveValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return v.validateCreate(ctx, obj.(*risingwavev1alpha1.RisingWave))
}

// ValidateDelete implements admission.CustomValidator.
func (v *RisingWaveValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (v *RisingWaveValidatingWebhook) isMetaStoragesTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.Storages.Meta, newObj.Spec.Storages.Meta)
}

func (v *RisingWaveValidatingWebhook) isObjectStoragesTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.Storages.Object, newObj.Spec.Storages.Object)
}

func pathForGroupReplicas(obj *risingwavev1alpha1.RisingWave, component, group string) *field.Path {
	if group == "" {
		return field.NewPath("spec", "global", "replicas", component)
	} else {
		index, _ := scaleview.NewRisingWaveScaleViewHelper(obj, component).GetGroupIndex(group)
		return field.NewPath("spec", "components", component, "groups").Index(index).Child("replicas")
	}
}

func (v *RisingWaveValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWave) error {
	gvk := oldObj.GroupVersionKind()

	// The storages must not be changed, especially meta and object.
	if !v.isMetaStoragesTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "meta"), "meta storage must be kept consistent"),
		)
	}

	if !v.isObjectStoragesTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "object"), "object storage must be kept consistent"),
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
func (v *RisingWaveValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error {
	// Validate the new object first.
	if err := v.ValidateCreate(ctx, newObj); err != nil {
		return err
	}

	return v.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWave), newObj.(*risingwavev1alpha1.RisingWave))
}

// groupTemplatePartitionExistAndIsString checks if has been set inside the upgrade strategy and if it is a string.
func groupTemplatePartitionExistAndIsString(groupTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate) bool {
	if groupTemplate.UpgradeStrategy.RollingUpdate == nil {
		return false
	}
	if groupTemplate.UpgradeStrategy.RollingUpdate.Partition == nil {
		return false
	}
	return groupTemplate.UpgradeStrategy.RollingUpdate.Partition.Type == intstr.String
}

// NewRisingWaveValidatingWebhook returns a new validator for the RisingWave. The behavior differs on different values of the
// openKruiseAvailable.
func NewRisingWaveValidatingWebhook(openKruiseAvailable bool) webhook.CustomValidator {
	return metrics.NewValidatingWebhookMetricsRecorder(&RisingWaveValidatingWebhook{openKruiseAvailable: openKruiseAvailable})
}
