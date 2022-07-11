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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RisingWaveComponentGroupTemplate struct {
	Image            string                      `json:"image,omitempty"`
	ImagePullPolicy  corev1.PullPolicy           `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets []string                    `json:"imagePullSecrets,omitempty"`
	UpgradeStrategy  appsv1.DeploymentStrategy   `json:"upgradeStrategy,omitempty"`
	Resources        corev1.ResourceRequirements `json:"resources,omitempty"`
	NodeSelector     map[string]string           `json:"nodeSelector,omitempty"`
	PodTemplate      *string                     `json:"podTemplate,omitempty"`
}

type RisingWaveComponentGroup struct {
	Name     string                            `json:"name,omitempty"`
	Replicas int64                             `json:"replicas,omitempty"`
	Template *RisingWaveComponentGroupTemplate `json:",inline"`
}

type RisingWaveComputeGroupTemplate struct {
	RisingWaveComponentGroupTemplate `json:",inline"`
	VolumeMounts                     []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

type RisingWaveComputeGroup struct {
	Name     string                          `json:"name,omitempty"`
	Replicas int64                           `json:"replicas,omitempty"`
	Template *RisingWaveComputeGroupTemplate `json:",inline"`
}

type RisingWaveComponentMetaPorts struct {
	RisingWaveComponentCommonPorts `json:",inline"`
	Dashboard                      int32 `json:"dashboard,omitempty"`
}

type RisingWaveComponentMeta struct {
	Ports  RisingWaveComponentMetaPorts `json:"ports,omitempty"`
	Groups []RisingWaveComponentGroup   `json:"groups,omitempty"`
}

type RisingWaveComponentCommonPorts struct {
	Service int32 `json:"service,omitempty"`
	Metrics int32 `json:"metrics,omitempty"`
}

type RisingWaveComponentFrontend struct {
	Ports  RisingWaveComponentCommonPorts `json:"ports,omitempty"`
	Groups []RisingWaveComponentGroup     `json:"groups,omitempty"`
}

type RisingWaveComponentCompute struct {
	Ports  RisingWaveComponentCommonPorts `json:"ports,omitempty"`
	Groups []RisingWaveComputeGroup       `json:"groups,omitempty"`
}

type RisingWaveComponentCompactor struct {
	Ports  RisingWaveComponentCommonPorts `json:"ports,omitempty"`
	Groups []RisingWaveComponentGroup     `json:"groups,omitempty"`
}

type RisingWaveComponentsSpec struct {
	Meta      RisingWaveComponentMeta      `json:"meta,omitempty"`
	Frontend  RisingWaveComponentFrontend  `json:"frontend,omitempty"`
	Compute   RisingWaveComponentCompute   `json:"compute,omitempty"`
	Compactor RisingWaveComponentCompactor `json:"compactor,omitempty"`
}

type RisingWaveMetaStorage struct {
	Memory *bool   `json:"memory,omitempty"`
	Etcd   *string `json:"etcd,omitempty"`
}

type RisingWaveObjectStorageMinIO struct {
	Endpoint string `json:"endpoint,omitempty"`
	Bucket   string `json:"bucket,omitempty"`
}

type RisingWaveObjectStorageS3 struct {
	Secret string `json:"secret,omitempty"`
}

type RisingWaveObjectStorage struct {
	Memory       *bool                          `json:"memory,omitempty"`
	MinIO        *RisingWaveObjectStorageMinIO  `json:"minio,omitempty"`
	S3           *RisingWaveObjectStorageS3     `json:"s3,omitempty"`
	PVCTemplates []corev1.PersistentVolumeClaim `json:"pvcTemplates,omitempty"`
}

type RisingWaveStoragesSpec struct {
	Meta RisingWaveMetaStorage `json:"meta,omitempty"`
}

type RisingWaveConfigurationSpec struct {
	ConfigMap corev1.ConfigMapKeySelector `json:"configmap,omitempty"`
}

type RisingWaveTLSConfigSecret struct {
	Name string `json:"name,omitempty"`
	Key  string `json:"key,omitempty"`
	Cert string `json:"cert,omitempty"`
}

type RisingWaveTLSConfig struct {
	Enabled bool                       `json:"enabled,omitempty"`
	Secret  *RisingWaveTLSConfigSecret `json:"secret,omitempty"`
}

type RisingWaveSecuritySpec struct {
	TLS *RisingWaveTLSConfig `json:"tls,omitempty"`
}

type RisingWaveGlobalReplicas struct {
	Meta      int64 `json:"meta,omitempty"`
	Frontend  int64 `json:"frontend,omitempty"`
	Compute   int64 `json:"compute,omitempty"`
	Compactor int64 `json:"compactor,omitempty"`
}

type RisingWaveGlobalSpec struct {
	Template RisingWaveComponentGroupTemplate `json:",inline"`
	Replicas RisingWaveGlobalReplicas         `json:"replicas,omitempty"`
}

type RisingWaveSpec struct {
	Global        RisingWaveGlobalSpec        `json:"global,omitempty"`
	Storages      RisingWaveStoragesSpec      `json:"storages,omitempty"`
	Components    RisingWaveComponentsSpec    `json:"components,omitempty"`
	Security      RisingWaveSecuritySpec      `json:"security,omitempty"`
	Configuration RisingWaveConfigurationSpec `json:"configuration,omitempty"`
}

type ComponentGroupReplicas struct {
	Name    string `json:"name"`
	Target  int64  `json:"target"`
	Running int64  `json:"running"`
}

type ComponentReplicas struct {
	Target  int64                    `json:"target"`
	Running int64                    `json:"running"`
	Groups  []ComponentGroupReplicas `json:"groups,omitempty"`
}

type RisingWaveComponentsReplicasStatus struct {
	Meta      ComponentReplicas `json:"meta"`
	Frontend  ComponentReplicas `json:"frontend"`
	Compute   ComponentReplicas `json:"compute"`
	Compactor ComponentReplicas `json:"compactor"`
}

type RisingWaveConditionType string

const (
	RisingWaveConditionRunning      RisingWaveConditionType = "Running"
	RisingWaveConditionInitializing RisingWaveConditionType = "Initializing"
	RisingWaveConditionUpgrading    RisingWaveConditionType = "Upgrading"
	RisingWaveConditionFailed       RisingWaveConditionType = "Failed"
	RisingWaveConditionUnknown      RisingWaveConditionType = "Unknown"
)

type RisingWaveCondition struct {
	// Type of the condition
	Type RisingWaveConditionType `json:"type"`

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

type RisingWaveMetaStorageStatus struct {
	Type string `json:"type"`
}

type RisingWaveObjectStorageStatus struct {
	Type string `json:"type"`
}

type RisingWaveStoragesStatus struct {
	Meta   RisingWaveMetaStorageStatus   `json:"meta"`
	Object RisingWaveObjectStorageStatus `json:"object"`
}

type RisingWaveStatus struct {
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	ComponentReplicas RisingWaveComponentsReplicasStatus `json:"componentReplicas,omitempty"`

	Conditions []RisingWaveCondition `json:"conditions,omitempty"`

	Storages RisingWaveStoragesStatus `json:"storages,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all
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

func init() {
	SchemeBuilder.Register(&RisingWave{}, &RisingWaveList{})
}
