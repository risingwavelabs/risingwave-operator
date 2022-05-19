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
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
//+kubebuilder:webhook:path=/mutate-risingwave-singularity-data-com-v1alpha1-risingwave,mutating=true,failurePolicy=fail,sideEffects=None,groups=risingwave.singularity-data.com,resources=risingwaves,verbs=create,versions=v1alpha1,name=mrisingwave.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &RisingWave{}

// Default implements webhook.Defaulter so a webhook will be registered for the type.
func (r *RisingWave) Default() {
	logger.Info("default", "name", r.Name)

	if len(r.Spec.Arch) == 0 {
		r.Spec.Arch = defaultOption.Arch
	}

	r.Finalizers = []string{
		MetaNodeFinalizer,
		ObjectStorageFinalizer,
		ComputeNodeFinalizer,
		CompactorNodeFinalizer,
		FrontendFinalizer,
	}

	r.defaultMeta()
	r.defaultStorage()
	r.defaultComputeNode()
	r.defaultCompactorNode()
	r.defaultFrontend()
}

func (r *RisingWave) defaultImage(o BaseOptions) *ImageDescriptor {
	image := o.Image[r.Spec.Arch]
	return &ImageDescriptor{
		Repository: &image.Repository,
		Tag:        &image.Tag,
		PullPolicy: &defaultOption.MetaNode.PullPolicy,
	}
}

func (r *RisingWave) defaultMeta() {
	if r.Spec.MetaNode == nil {
		r.Spec.MetaNode = &MetaNodeSpec{}
	}

	meta := r.Spec.MetaNode
	if meta.Image == nil {
		meta.Image = r.defaultImage(defaultOption.MetaNode)
	}

	if meta.Storage == nil {
		meta.Storage = &MetaStorage{
			Type: InMemory,
		}
	}

	if meta.Replicas == nil {
		meta.Replicas = &defaultOption.MetaNode.Replicas
	}

	if len(meta.Ports) == 0 {
		meta.Ports = []corev1.ContainerPort{
			{
				Name: MetaServerPortName,

				ContainerPort: MetaServerPort,
			},
			{
				Name: MetaDashboardPortName,

				ContainerPort: MetaDashboardPort,
			},
		}
	}

	if meta.Resources == nil {
		meta.Resources = defaultOption.MetaNode.Resources.DeepCopy()
	}

	meta.NodeSelector = map[string]string{
		ArchKey: string(r.Spec.Arch),
	}
}

func (r *RisingWave) defaultStorage() {
	if r.Spec.ObjectStorage == nil {
		r.Spec.ObjectStorage = &ObjectStorageSpec{
			Memory: true,
		}
		return
	}

	storage := r.Spec.ObjectStorage

	// default minIO spec
	if storage.MinIO != nil {
		minIO := storage.MinIO
		if minIO.Image == nil {
			minIO.Image = r.defaultImage(defaultOption.MinIO)
		}

		if len(minIO.Ports) == 0 {
			minIO.Ports = []corev1.ContainerPort{
				{
					Name: MinIOServerPortName,

					ContainerPort: MinIOServerPort,
				},
				{
					Name: MinIOConsolePortName,

					ContainerPort: MinIOConsolePort,
				},
			}
		}

		if minIO.Replicas == nil {
			minIO.Replicas = &defaultOption.MinIO.Replicas
		}

		if minIO.Resources == nil {
			minIO.Resources = defaultOption.MinIO.Resources.DeepCopy()
		}

		minIO.NodeSelector = map[string]string{
			ArchKey: string(r.Spec.Arch),
		}
		return

	}

	if r.Spec.ObjectStorage.S3 != nil {
		s3 := r.Spec.ObjectStorage.S3
		if len(s3.SecretName) == 0 {
			s3.SecretName = CloudProviderConfigureSecretName
		}
	}
}

func (r *RisingWave) defaultComputeNode() {
	if r.Spec.ComputeNode == nil {
		r.Spec.ComputeNode = &ComputeNodeSpec{}
	}

	compute := r.Spec.ComputeNode
	if compute.Image == nil {
		compute.Image = r.defaultImage(defaultOption.ComputeNode)
	}
	if compute.Replicas == nil {
		compute.Replicas = &defaultOption.ComputeNode.Replicas
	}

	if len(compute.Ports) == 0 {
		compute.Ports = []corev1.ContainerPort{
			{
				Name:          ComputeNodePortName,
				ContainerPort: ComputeNodePort,
			},
		}
	}

	if compute.Resources == nil {
		compute.Resources = defaultOption.ComputeNode.Resources.DeepCopy()
	}

	compute.NodeSelector = map[string]string{
		ArchKey: string(r.Spec.Arch),
	}
}

func (r *RisingWave) defaultCompactorNode() {
	if r.Spec.CompactorNode == nil {
		r.Spec.CompactorNode = &CompactorNodeSpec{}
	}

	compactor := r.Spec.CompactorNode
	defaultValue := defaultOption.CompactorNode
	if compactor.Image == nil {
		compactor.Image = r.defaultImage(defaultValue)
	}

	if compactor.Replicas == nil {
		compactor.Replicas = &defaultValue.Replicas
	}

	if len(compactor.Ports) == 0 {
		compactor.Ports = []corev1.ContainerPort{
			{
				Name:          CompactorNodePortName,
				ContainerPort: CompactorNodePort,
			},
		}
	}

	if compactor.Resources == nil {
		compactor.Resources = defaultValue.Resources.DeepCopy()
	}

	compactor.NodeSelector = map[string]string{
		ArchKey: string(r.Spec.Arch),
	}
}

func (r *RisingWave) defaultFrontend() {
	if r.Spec.Frontend == nil {
		r.Spec.Frontend = &FrontendSpec{}
	}

	spec := r.Spec.Frontend
	if spec.Image == nil {
		spec.Image = r.defaultImage(defaultOption.Frontend)
	}
	if spec.Replicas == nil {
		spec.Replicas = &defaultOption.Frontend.Replicas
	}

	if len(spec.Ports) == 0 {
		spec.Ports = []corev1.ContainerPort{
			{
				Name:          FrontendPortName,
				ContainerPort: FrontendPort,
			},
		}
	}

	if spec.Resources == nil {
		spec.Resources = defaultOption.Frontend.Resources.DeepCopy()
	}

	spec.NodeSelector = map[string]string{
		ArchKey: string(r.Spec.Arch),
	}
}
