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

package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:path=/validate-risingwave-singularity-data-com-v1alpha1-risingwave,mutating=false,failurePolicy=fail,sideEffects=None,groups=risingwave.singularity-data.com,resources=risingwaves,verbs=create;update;delete,versions=v1alpha1,name=vrisingwave.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RisingWave{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type.
func (r *RisingWave) ValidateCreate() error {
	logger.V(1).Info("validate create", "name", r.Name)

	if r.Spec.MetaNode != nil {
		err := validateImage(r.Spec.MetaNode.DeployDescriptor.Image)
		if err != nil {
			return err
		}
	}

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type.
func (r *RisingWave) ValidateUpdate(old runtime.Object) error {
	logger.V(1).Info("validate update", "name", r.Name)

	oldRisingWave := old.(*RisingWave)

	newSpec := r.Spec
	oldSpec := oldRisingWave.Spec

	if newSpec.Arch != oldSpec.Arch {
		return NewValidateError("Arch cannot be changed")
	}

	err := validateMetaNode(newSpec.MetaNode, oldSpec.MetaNode)
	if err != nil {
		return err
	}

	err = validateObjectStorage(newSpec.ObjectStorage, oldSpec.ObjectStorage)
	if err != nil {
		return err
	}

	err = validateComputeNode(newSpec.ComputeNode, oldSpec.ComputeNode)
	if err != nil {
		return err
	}

	err = validateFrontend(newSpec.Frontend, oldSpec.Frontend)
	if err != nil {
		return err
	}

	return nil
}

func validateMetaNode(n, o *MetaNodeSpec) error {
	if o.Storage == nil || n.Storage == nil || n.Storage.Type != o.Storage.Type {
		return NewValidateError("Meta storage cannot be changed")
	}

	err := validateDeployDescriptor(n.DeployDescriptor, o.DeployDescriptor)
	if err != nil {
		return err
	}
	return nil
}

func validateObjectStorage(n, o *ObjectStorageSpec) error {
	if n.MinIO != nil {
		err := validateDeployDescriptor(n.MinIO.DeployDescriptor, o.MinIO.DeployDescriptor)
		if err != nil {
			return err
		}
	}

	if n.S3 != nil {
		if n.S3.Bucket == nil {
			return &validateError{
				Msg: fmt.Sprintf("bucket cannot empty when use S3"),
			}
		}
	}

	return nil
}

func validateComputeNode(n, o *ComputeNodeSpec) error {
	err := validateDeployDescriptor(n.DeployDescriptor, o.DeployDescriptor)
	if err != nil {
		return err
	}
	return nil
}

func validateFrontend(n, o *FrontendSpec) error {
	err := validateDeployDescriptor(n.DeployDescriptor, o.DeployDescriptor)
	if err != nil {
		return err
	}
	return nil
}

func validateDeployDescriptor(n, o DeployDescriptor) *validateError {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type.
func (r *RisingWave) ValidateDelete() error {
	logger.V(1).Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func validateImage(descriptor *ImageDescriptor) error {
	if descriptor == nil {
		return nil
	}

	if descriptor.Repository == nil {
		return NewValidateError("image repository cannot be empty")
	}
	return nil
}

type validateError struct {
	Msg string
}

func NewValidateError(msg string) error {
	return &validateError{
		Msg: msg,
	}
}

func (e *validateError) Error() string {
	return fmt.Sprintf("failed to validate, error message: [%s]", e.Msg)
}
