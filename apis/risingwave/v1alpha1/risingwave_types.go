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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all
// +kubebuilder:printcolumn:name="Running",type=string,JSONPath=`.status.conditions[?(@.type=="Running")].status`
// +kubebuilder:printcolumn:name="Storage",type=string,JSONPath=`.status.objectStorage.type`

// RisingWave is the Schema for the risingwaves API.
type RisingWave struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RisingWaveSpec   `json:"spec,omitempty"`
	Status RisingWaveStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RisingWaveList contains a list of RisingWave.
type RisingWaveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RisingWave `json:"items"`
}

// RisingWaveSpec defines the desired state of RisingWave.
type RisingWaveSpec struct {
	Arch Arch `json:"arch,omitempty"`

	MetaNode *MetaNodeSpec `json:"metaNode,omitempty"`

	ObjectStorage *ObjectStorageSpec `json:"objectStorage,omitempty"`

	ComputeNode *ComputeNodeSpec `json:"computeNode,omitempty"`

	CompactorNode *CompactorNodeSpec `json:"compactorNode,omitempty"`

	Frontend *FrontendSpec `json:"frontend,omitempty"`
}

// Arch defines what machine architecture on which RisingWave should run.
type Arch string

const (
	AMD64Arch Arch = "amd64"
	ARM64Arch Arch = "arm64"
)

// RisingWaveStatus defines the observed state of RisingWave.
type RisingWaveStatus struct {
	MetaNode      MetaNodeStatus      `json:"metaNode,omitempty"`
	ObjectStorage ObjectStorageStatus `json:"objectStorage,omitempty"`
	ComputeNode   ComputeNodeStatus   `json:"computeNode,omitempty"`
	CompactorNode CompactorNodeStatus `json:"compactorNode,omitempty"`
	Frontend      FrontendSpecStatus  `json:"frontend,omitempty"`

	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	Conditions         []RisingWaveCondition `json:"conditions,omitempty"`
}

type RisingWaveType string

const (
	Initializing RisingWaveType = "Initializing"
	Running      RisingWaveType = "Running"
	Upgrading    RisingWaveType = "Upgrading"
	Failed       RisingWaveType = "Failed"
	Unknown      RisingWaveType = "Unknown"
)

// RisingWaveCondition describes the condition for RisingWave.
type RisingWaveCondition struct {
	// Type of the condition
	Type RisingWaveType `json:"type"`

	// Status of the condition
	Status metav1.ConditionStatus `json:"status"`

	// Last time the condition transitioned from one status to another.
	// +optional
	// +nullable
	LastTransitionTime metav1.Time `json:"lastTransitionTime"`

	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

type MetaNodeSpec struct {
	DeployDescriptor `json:",inline"`

	//default Memory
	Storage *MetaStorage `json:"storage"`
}

//MetaStorageType defines the storage type of meta node.
type MetaStorageType string

const (
	// InMemory - use memory for meta node storage.
	InMemory MetaStorageType = "InMemory"
	// ETCD - use ETCD for meta node storage.
	ETCD MetaStorageType = "ETCD"
)

// MetaStorage defines spec of meta service.
type MetaStorage struct {
	Type MetaStorageType `json:"type"`
}

// ComputeNodeSpec defines the spec of compute-node
// No need storage information.
// Which storage to use depends on spec.objectStorageSpec.
type ComputeNodeSpec struct {
	DeployDescriptor `json:",inline"`
}

// CompactorNodeSpec defines the spec of compactor-node.
type CompactorNodeSpec struct {
	DeployDescriptor `json:",inline"`
}

// ObjectStorageSpec defines spec of object storage
// TODO: support more backend types.
type ObjectStorageSpec struct {
	// TODO: support s3 config
	S3 *S3 `json:"s3,omitempty"`

	Memory bool `json:"memory,omitempty"`

	MinIO *MinIO `json:"minIO,omitempty"`
}

// S3 store the s3 information.
type S3 struct {
	// the name of Provider, default AWS
	Provider string `json:"provider"`

	// the name of s3 bucket, if not set, operator will create new bucket
	Bucket *string `json:"bucket,omitempty"`

	// the secret name of s3 client configure
	SecretName string `json:"secret,omitempty"`
}

// FrontendSpec defines spec of frontend.
type FrontendSpec struct {
	DeployDescriptor `json:",inline"`
}

// DeployDescriptor describe the deploy information.
type DeployDescriptor struct {
	// +optional
	Image *ImageDescriptor `json:"image,omitempty"`

	// +optional
	Replicas *int32 `json:"replicas,omitempty"`

	// +optional
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// +optional
	Ports []corev1.ContainerPort `json:"ports,omitempty"`

	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +optional
	CMD []string `json:"cmd,omitempty"`
}

// ImageDescriptor describe the image information.
type ImageDescriptor struct {
	Repository *string `json:"repository,omitempty"`
	Tag        *string `json:"tag,omitempty"`

	// PullPolicy for the image.
	// Default: IfNotPresent
	PullPolicy *corev1.PullPolicy `json:"pullPolicy,omitempty"`
}

// MinIO defines minIO deploy information.
type MinIO struct {
	DeployDescriptor `json:",inline"`
}

// CloudService defines storage provider.
type CloudService struct {
	Provider string `json:"provider"`
}

// ObjectStorageType defines storage provider.
type ObjectStorageType string

const (
	MinIOType ObjectStorageType = "MinIO"

	S3Type ObjectStorageType = "S3"

	MemoryType ObjectStorageType = "Memory"

	UnknownType ObjectStorageType = "Unknown"
)

type ComponentPhase string

const (
	ComponentUpgrading    ComponentPhase = "Upgrading" // TODO: support component upgrade
	ComponentInitializing ComponentPhase = "Initializing"
	ComponentScaling      ComponentPhase = "Scaling"
	ComponentReady        ComponentPhase = "Ready"
	ComponentFailed       ComponentPhase = "Failed"
	ComponentUnknown      ComponentPhase = "Unknown"
)

// MetaNodeStatus defines status of meta node.
type MetaNodeStatus struct {
	Phase ComponentPhase `json:"phase,omitempty"`

	// Total number of non-terminated pods.
	Replicas int32 `json:"replicas,omitempty"`
}

// ObjectStorageStatus defines status of object storage.
type ObjectStorageStatus struct {
	Phase ComponentPhase `json:"phase,omitempty"`

	MinIOStatus *MinIOStatus `json:"minio,omitempty"`

	S3 bool `json:"s3,omitempty"`

	StorageType ObjectStorageType `json:"type,omitempty"`
}

// MinIOStatus define the status of MinIO storage.
type MinIOStatus struct {
	// Total number of non-terminated pods.
	Replicas int32 `json:"replicas,omitempty"`
}

// ComputeNodeStatus defines status of compute node.
type ComputeNodeStatus struct {
	Phase ComponentPhase `json:"phase,omitempty"`

	// Total number of non-terminated pods.
	Replicas int32 `json:"replicas,omitempty"`
}

// CompactorNodeStatus defines status of compactor node.
type CompactorNodeStatus struct {
	Phase ComponentPhase `json:"phase,omitempty"`

	// Total number of non-terminated pods.
	Replicas int32 `json:"replicas,omitempty"`
}

// FrontendSpecStatus defines status of compute frontend.
type FrontendSpecStatus struct {
	Phase ComponentPhase `json:"phase,omitempty"`

	// Total number of non-terminated pods.
	Replicas int32 `json:"replicas,omitempty"`
}

func init() {
	SchemeBuilder.Register(&RisingWave{}, &RisingWaveList{})
}
