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

	"github.com/distribution/distribution/reference"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type RisingWaveValidatingWebhook struct {
}

func isImageValid(image string) bool {
	return reference.ReferenceRegexp.MatchString(image)
}

func (h *RisingWaveValidatingWebhook) validateGroupTemplate(path *field.Path, groupTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate) field.ErrorList {
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

	return fieldErrs
}

func ptrValueNotZero[T comparable](ptr *T) bool {
	var zero T
	return ptr != nil && *ptr != zero
}

func (h *RisingWaveValidatingWebhook) validateStorages(path *field.Path, storages *risingwavev1alpha1.RisingWaveStoragesSpec) field.ErrorList {
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

	isObjectMemory, isObjectMinIO, isObjectS3 := ptrValueNotZero(storages.Object.Memory), storages.Object.MinIO != nil, storages.Object.S3 != nil
	validObjectStorageTypeCount := lo.CountBy([]bool{isObjectMemory, isObjectMinIO, isObjectS3}, func(x bool) bool { return x })
	if validObjectStorageTypeCount == 0 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("object"), storages.Object, "must configure the object storage"))
	} else if validObjectStorageTypeCount > 1 {
		fieldErrs = append(fieldErrs, field.Invalid(path.Child("object"), storages.Object, "multiple object storage types"))
	}

	return fieldErrs
}

func (h *RisingWaveValidatingWebhook) validateSecurity(path *field.Path, security *risingwavev1alpha1.RisingWaveSecuritySpec) field.ErrorList {
	return nil
}

func (h *RisingWaveValidatingWebhook) validateConfiguration(path *field.Path, configuration *risingwavev1alpha1.RisingWaveConfigurationSpec) field.ErrorList {
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

func (h *RisingWaveValidatingWebhook) validateComponents(path *field.Path, components *risingwavev1alpha1.RisingWaveComponentsSpec, storages *risingwavev1alpha1.RisingWaveStoragesSpec) field.ErrorList {
	fieldErrs := field.ErrorList{}

	metaGroupsPath := path.Child("meta", "groups")
	for i, group := range components.Meta.Groups {
		fieldErrs = append(fieldErrs, h.validateGroupTemplate(metaGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate)...)
	}

	frontendGroupsPath := path.Child("frontend", "groups")
	for i, group := range components.Frontend.Groups {
		fieldErrs = append(fieldErrs, h.validateGroupTemplate(frontendGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate)...)
	}

	compactorGroupsPath := path.Child("compactor", "groups")
	for i, group := range components.Compactor.Groups {
		fieldErrs = append(fieldErrs, h.validateGroupTemplate(compactorGroupsPath.Index(i), group.RisingWaveComponentGroupTemplate)...)
	}

	pvClaims := make(map[string]int)
	for _, pvc := range storages.PVCTemplates {
		pvClaims[pvc.Name] = 1
	}

	computeGroupsPath := path.Child("compute", "groups")
	for i, group := range components.Compute.Groups {
		if group.RisingWaveComputeGroupTemplate != nil {
			fieldErrs = append(fieldErrs, h.validateGroupTemplate(computeGroupsPath.Index(i), &group.RisingWaveComponentGroupTemplate)...)

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

func (h *RisingWaveValidatingWebhook) validateCreate(ctx context.Context, obj *risingwavev1alpha1.RisingWave) error {
	gvk := obj.GroupVersionKind()

	fieldErrs := field.ErrorList{}

	// Validate the global spec.
	fieldErrs = append(fieldErrs, h.validateGroupTemplate(field.NewPath("spec", "global"), &obj.Spec.Global.RisingWaveComponentGroupTemplate)...)

	// Validate the storages spec.
	fieldErrs = append(fieldErrs, h.validateStorages(field.NewPath("spec", "storages"), &obj.Spec.Storages)...)

	// Validate the security spec.
	fieldErrs = append(fieldErrs, h.validateSecurity(field.NewPath("spec", "security"), &obj.Spec.Security)...)

	// Validate the configuration spec.
	fieldErrs = append(fieldErrs, h.validateConfiguration(field.NewPath("spec", "configuration"), &obj.Spec.Configuration)...)

	// Validate the components spec.
	fieldErrs = append(fieldErrs, h.validateComponents(field.NewPath("spec", "components"), &obj.Spec.Components, &obj.Spec.Storages)...)

	if len(fieldErrs) > 0 {
		return apierrors.NewInvalid(gvk.GroupKind(), obj.Name, fieldErrs)
	}
	return nil
}

// ValidateCreate implements admission.CustomValidator.
func (h *RisingWaveValidatingWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	return h.validateCreate(ctx, obj.(*risingwavev1alpha1.RisingWave))
}

// ValidateDelete implements admission.CustomValidator.
func (h *RisingWaveValidatingWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	return nil
}

func (h *RisingWaveValidatingWebhook) isMetaStoragesTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.Storages.Meta, newObj.Spec.Storages.Meta)
}

func (h *RisingWaveValidatingWebhook) isObjectStoragesTheSame(oldObj, newObj *risingwavev1alpha1.RisingWave) bool {
	return equality.Semantic.DeepEqual(oldObj.Spec.Storages.Object, newObj.Spec.Storages.Object)
}

func (h *RisingWaveValidatingWebhook) validateUpdate(ctx context.Context, oldObj, newObj *risingwavev1alpha1.RisingWave) error {
	gvk := oldObj.GroupVersionKind()

	// The storages must not be changed, especially meta and object.
	if !h.isMetaStoragesTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "meta"), "meta storage must be kept consistent"),
		)
	}

	if !h.isObjectStoragesTheSame(oldObj, newObj) {
		return apierrors.NewForbidden(
			schema.GroupResource{Group: gvk.Group, Resource: gvk.Kind},
			oldObj.Name,
			field.Forbidden(field.NewPath("spec", "storages", "object"), "object storage must be kept consistent"),
		)
	}

	return nil
}

// ValidateUpdate implements admission.CustomValidator.
func (h *RisingWaveValidatingWebhook) ValidateUpdate(ctx context.Context, oldObj runtime.Object, newObj runtime.Object) error {
	// Validate the new object first.
	if err := h.ValidateCreate(ctx, newObj); err != nil {
		return err
	}

	return h.validateUpdate(ctx, oldObj.(*risingwavev1alpha1.RisingWave), newObj.(*risingwavev1alpha1.RisingWave))
}

func NewRisingWaveValidatingWebhook() webhook.CustomValidator {
	return &RisingWaveValidatingWebhook{}
}
