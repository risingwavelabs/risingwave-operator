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
	"k8s.io/apimachinery/pkg/util/intstr"
)

type RisingWaveUpgradeStrategyType string

const (
	RisingWaveUpgradeStrategyTypeRecreate      RisingWaveUpgradeStrategyType = "Recreate"
	RisingWaveUpgradeStrategyTypeRollingUpdate RisingWaveUpgradeStrategyType = "RollingUpdate"
)

type RisingWaveRollingUpdate struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// Defaults to 25%.
	// +optional
	// +kubebuilder:default="25%"
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,1,opt,name=maxUnavailable"`
}

type RisingWaveUpgradeStrategy struct {
	// Type of upgrade. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	// +kubebuilder:default=RollingUpdate
	// +kubebuilder:validation:Enum=Recreate;RollingUpdate
	Type RisingWaveUpgradeStrategyType `json:"type,omitempty"`

	// Rolling update config params. Present only if DeploymentStrategyType =
	// RollingUpdate.
	//---
	// TODO: Update this to follow our convention for oneOf, whatever we decide it
	// to be.
	// +optional
	RollingUpdate *RisingWaveRollingUpdate `json:"rollingUpdate,omitempty"`
}

// RisingWaveComponentGroupTemplate is the common deployment template for groups of each component.
// Currently we use the common template for meta/frontend/compactor.
type RisingWaveComponentGroupTemplate struct {
	// Image is the RisingWave image used.
	// +optional
	Image string `json:"image,omitempty"`

	// ImagePullPolicy is the pull policy of RisingWave image. The default value is the same as the
	// default of Kubernetes.
	// +optional
	// +kubebuilder:default=IfNotPresent
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// ImagePullSecrets are the secrets for pulling RisingWave images.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// UpgradeStrategy is the upgrade strategy for the components. By default it is the same as the
	// workload's default strategy that the component is deployed with.
	// Note: the maxSurge will not take effect for the compute component.
	// +optional
	// +patchStrategy=retainKeys
	UpgradeStrategy RisingWaveUpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Resources are the running resources of the RisingWave component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// NodeSelector is a map of labels describing the nodes to be scheduled on.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// PodTemplate is the base template for Pods of RisingWave. By default there's no such template
	// and the controller will set all unrelated fields to the default value.
	// +optional
	PodTemplate *string `json:"podTemplate,omitempty"`
}

// RisingWaveComponentGroup is the common deployment group of each component. Currently we use
// this group for meta/frontend/compactor.
type RisingWaveComponentGroup struct {
	// Name is the name of the group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Replicas is the replicas of Pods in this group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Template is the component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there're
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	*RisingWaveComponentGroupTemplate `json:",inline"`
}

// RisingWaveComputeGroupTemplate is the group template for component compute, which supports specifying
// the volume mounts on the compute Pods. The volumes should be either local or defined in the storages.
type RisingWaveComputeGroupTemplate struct {
	// Template is the component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there're
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	RisingWaveComponentGroupTemplate `json:",inline"`

	// VolumeMounts are the volumes to be mounted on the Pods.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// RisingWaveComputeGroup is the group for component compute.
type RisingWaveComputeGroup struct {
	// Name is the name of the group.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Replicas is the replicas of Pods in this group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Template is the component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there're
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	*RisingWaveComputeGroupTemplate `json:",inline"`
}

// RisingWaveComponentCommonPorts are the common ports that components need to listen.
type RisingWaveComponentCommonPorts struct {
	// ServicePort is the service port of the component. For each component,
	// the 'service' has different meanings. It's an optional field and if it's left out, a
	// default port (varies among components) will be used.
	// +optional
	// +kubebuilder:validation:Minimum=1
	ServicePort int32 `json:"service,omitempty"`

	// MetricsPort is the metrics port of the component. It always serves the metrics in
	// Prometheus format.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MetricsPort int32 `json:"metrics,omitempty"`
}

// RisingWaveComponentMetaPorts are the ports of component meta.
type RisingWaveComponentMetaPorts struct {
	RisingWaveComponentCommonPorts `json:",inline"`

	// DashboardPort is the dashboard port of the meta, a default value of 8080 will be
	// used if not specified.
	// +optional
	// +kubebuilder:validation:Minimum=1
	DashboardPort int32 `json:"dashboard,omitempty"`
}

type RisingWaveComponentMeta struct {
	// RestartAt is the time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports are the ports to be listened by the meta Pods.
	// +optional
	// +kubebuilder:default={service: 5690, metrics: 1250, dashboard: 5691}
	Ports RisingWaveComponentMetaPorts `json:"ports,omitempty"`

	// Groups are the groups of Pods of the meta component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

type RisingWaveComponentFrontend struct {
	// RestartAt is the time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports are the ports to be listened by the frontend Pods.
	// +optional
	// +kubebuilder:default={service: 4567, metrics: 8080}
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups are the groups of Pods of the frontend component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

type RisingWaveComponentCompute struct {
	// RestartAt is the time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports are the ports to be listened by the compute Pods.
	// +optional
	// +kubebuilder:default={service: 5688, metrics: 1222}
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups are the groups of Pods of the compute component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComputeGroup `json:"groups,omitempty"`
}

type RisingWaveComponentCompactor struct {
	// RestartAt is the time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports are the ports to be listened by the compactor Pods.
	// +optional
	// +kubebuilder:default={service: 6660, metrics: 1260}
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups are the groups of Pods of the compactor component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

// RisingWaveComponentsSpec is the spec describes the components of RisingWave.
type RisingWaveComponentsSpec struct {
	// Meta is for the meta component.
	Meta RisingWaveComponentMeta `json:"meta,omitempty"`

	// Frontend is for the frontend component.
	Frontend RisingWaveComponentFrontend `json:"frontend,omitempty"`

	// Compute is for the compute component.
	Compute RisingWaveComponentCompute `json:"compute,omitempty"`

	// Compactor is for the compactor component.
	Compactor RisingWaveComponentCompactor `json:"compactor,omitempty"`
}

// RisingWaveMetaStorageEtcd is the etcd storage for the meta component.
type RisingWaveMetaStorageEtcd struct {
	// Endpoint is the endpoint of etcd. It must be provided.
	Endpoint string `json:"endpoint"`

	// Secret is the secret contains credentials of etcd, it must contains the following keys:
	//   * username
	//   * password
	// But it is an optional field. Empty value indicates etcd is available without authentication.
	// +optional
	Secret string `json:"secret,omitempty"`
}

// RisingWaveMetaStorage is the storage for the meta component.
type RisingWaveMetaStorage struct {
	// Memory indicates to store the metadata in memory. It is only for test usage and strongly
	// discouraged to be set in production. If one is using the memory storage for meta,
	// replicas will not work because they are not going to share the same metadata and any kinds
	// exit of the process will cause a permanent loss of the data.
	// +optional
	Memory *bool `json:"memory,omitempty"`

	// Etcd is the endpoint of the etcd service for storing the metadata.
	// +optional
	Etcd *RisingWaveMetaStorageEtcd `json:"etcd,omitempty"`
}

// RisingWaveObjectStorageMinIO is the details of MinIO storage for the compute and compactor components.
type RisingWaveObjectStorageMinIO struct {
	// Secret is the secret containing the credentials to access the MinIO service. It must contain the
	// the following keys:
	//   * username
	//   * password
	// +kubebuilder:validation:Required
	Secret string `json:"secret,omitempty"`

	// Endpoint is the endpoint of the MinIO service.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint,omitempty"`

	// Bucket is the bucket of the MinIO service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket,omitempty"`
}

// RisingWaveObjectStorageS3 is the details of AWS S3 storage for the compute and compactor components.
type RisingWaveObjectStorageS3 struct {
	// Secret is the secret containing the credentials to access the AWS S3 service. It must contain the
	// the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region
	// +kubebuilder:validation:Required
	Secret string `json:"secret,omitempty"`

	// Bucket is the bucket of the AWS S3 service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket,omitempty"`
}

// RisingWaveObjectStorage is the object storage for the compute and compactor components.
type RisingWaveObjectStorage struct {
	// Memory indicates to store the data in memory. It's only for test usage and strongly discouraged to
	// be used in production.
	// +optional
	Memory *bool `json:"memory,omitempty"`

	// MinIO indicates to use the MinIO storage.
	// +optional
	MinIO *RisingWaveObjectStorageMinIO `json:"minio,omitempty"`

	// S3 indicates to use the S3 storage.
	// +optional
	S3 *RisingWaveObjectStorageS3 `json:"s3,omitempty"`
}

// RisingWaveStoragesSpec is the storages spec.
type RisingWaveStoragesSpec struct {
	// Meta is the storage declaration of meta.

	Meta RisingWaveMetaStorage `json:"meta,omitempty"`

	// Object is the storage declaration of compute and compactor.
	Object RisingWaveObjectStorage `json:"object,omitempty"`

	// PVCTemplates are the persistent volume claim templates for the compute component. PVCs declared here
	// can be referenced in the groups of compute component.
	// +optional
	PVCTemplates []corev1.PersistentVolumeClaim `json:"pvcTemplates,omitempty"`
}

// RisingWaveConfigurationSpec is the configuration spec.
type RisingWaveConfigurationSpec struct {
	// ConfigMap is the reference to a key in a config map that contains the base config for RisingWave.
	// It's an optional field and can be left out. If not specified, a default config is going to be used.
	// +optional
	ConfigMap *corev1.ConfigMapKeySelector `json:"configmap,omitempty"`
}

// RisingWaveTLSConfigSecret is the secret reference that contains the key and cert for TLS.
type RisingWaveTLSConfigSecret struct {
	// Name is the name of the secret.
	// +optional
	Name string `json:"name,omitempty"`

	// Key is the key of the TLS key. A default value of "tls.key" will be used if not specified.
	// +optional
	// +kubebuilder:default=tls.key
	Key string `json:"key,omitempty"`

	// Cert is the cert of the TLS cert. A default value of "tls.crt" will be used if not specified.
	// +optional
	// +kubebuilder:default=tls.crt
	Cert string `json:"cert,omitempty"`
}

// RisingWaveTLSConfig is the TLS config of RisingWave.
type RisingWaveTLSConfig struct {
	// Enabled indicates if TLS is enabled on RisingWave.
	Enabled bool `json:"enabled,omitempty"`

	// Secret is the secret that contains the TLS config. If TLS is enabled, the secret
	// must be provided.
	// +optional.
	Secret RisingWaveTLSConfigSecret `json:"secret,omitempty"`
}

// RisingWaveSecuritySpec is the security spec.
type RisingWaveSecuritySpec struct {
	// TLS is the TLS config of RisingWave.
	// +optional
	// +patchStrategy=retainKeys
	TLS *RisingWaveTLSConfig `json:"tls,omitempty"`
}

// RisingWaveGlobalReplicas are the replicas of each component, declared in global scope.
type RisingWaveGlobalReplicas struct {
	// Meta is the replicas of meta component.
	// +optional
	Meta int32 `json:"meta,omitempty"`

	// Frontend is the replicas of frontend component.
	// +optional
	Frontend int32 `json:"frontend,omitempty"`

	// Compute is the replicas of compute component.
	// +optional
	Compute int32 `json:"compute,omitempty"`

	// Compactor is the replicas of compactor component.
	// +optional
	Compactor int32 `json:"compactor,omitempty"`
}

// RisingWaveGlobalSpec is the global spec.
type RisingWaveGlobalSpec struct {
	// Template is the global template for RisingWave that all components share.
	RisingWaveComponentGroupTemplate `json:",inline"`

	// Replicas is the replicas of each component, in global scope.
	// +optional
	// +patchStrategy=retainKeys
	Replicas RisingWaveGlobalReplicas `json:"replicas,omitempty"`

	// ServiceType is the service type of the frontend service.
	// +optional
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
}

// RisingWaveSpec is the overall spec.
type RisingWaveSpec struct {
	// Global contains a shared template for components and a global scope of replicas.
	Global RisingWaveGlobalSpec `json:"global,omitempty"`

	// Storages contains the storage for meta, the object storage for compute and compactor, and
	// the PVC template for compute.
	Storages RisingWaveStoragesSpec `json:"storages,omitempty"`

	// Components contains the ports and some controllers (such as `restartAt`) of each component,
	// as well as an advanced concept called `group` to override the global template and create groups
	// of Pods, e.g., deployment in hybrid-arch cluster.
	Components RisingWaveComponentsSpec `json:"components,omitempty"`

	// Security contains the TLS configuration.
	Security RisingWaveSecuritySpec `json:"security,omitempty"`

	// Configuration contains the config template configuration.
	Configuration RisingWaveConfigurationSpec `json:"configuration,omitempty"`
}

// ComponentGroupReplicasStatus are the running status of Pods in group.
type ComponentGroupReplicasStatus struct {
	// Name is the name of the group.
	Name string `json:"name"`

	// Target is the target replicas of the group.
	Target int32 `json:"target"`

	// Running is the running replicas in the group.
	Running int32 `json:"running"`
}

// ComponentReplicasStatus are the running status of Pods of the component.
type ComponentReplicasStatus struct {
	// Target is the total target replicas of the component.
	Target int32 `json:"target"`

	// Running is the total running replicas of the component.
	Running int32 `json:"running"`

	// Groups are a list of running status of each group.
	Groups []ComponentGroupReplicasStatus `json:"groups,omitempty"`
}

// RisingWaveComponentsReplicasStatus is the running status of components.
type RisingWaveComponentsReplicasStatus struct {
	// Meta is the running status of meta.
	Meta ComponentReplicasStatus `json:"meta"`

	// Frontend is the running status of frontend.
	Frontend ComponentReplicasStatus `json:"frontend"`

	// Compute is the running status of compute.
	Compute ComponentReplicasStatus `json:"compute"`

	// Compactor is the running status of compactor.
	Compactor ComponentReplicasStatus `json:"compactor"`
}

// RisingWaveConditionType is the condition type of RisingWave.
type RisingWaveConditionType string

// These are valid value of RisingWaveConditionType.
const (
	RisingWaveConditionRunning      RisingWaveConditionType = "Running"
	RisingWaveConditionInitializing RisingWaveConditionType = "Initializing"
	RisingWaveConditionUpgrading    RisingWaveConditionType = "Upgrading"
	RisingWaveConditionFailed       RisingWaveConditionType = "Failed"
	RisingWaveConditionUnknown      RisingWaveConditionType = "Unknown"
)

// RisingWaveCondition indicates a condition of RisingWave.
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

// MetaStorageType is the type name of meta storage.
type MetaStorageType string

// These are valid values of MetaStorageType.
const (
	MetaStorageTypeMemory  MetaStorageType = "memory"
	MetaStorageTypeEtcd    MetaStorageType = "etcd"
	MetaStorageTypeUnknown MetaStorageType = "unknown"
)

// RisingWaveMetaStorageStatus is the status of meta storage.
type RisingWaveMetaStorageStatus struct {
	Type MetaStorageType `json:"type"`
}

// ObjectStorageType is the type name of object storage.
type ObjectStorageType string

// These are valid values of ObjectStorageType.
const (
	ObjectStorageTypeMemory  ObjectStorageType = "memory"
	ObjectStorageTypeMinIO   ObjectStorageType = "minio"
	ObjectStorageTypeS3      ObjectStorageType = "s3"
	ObjectStorageTypeUnknown ObjectStorageType = "unknown"
)

// RisingWaveObjectStorageStatus is the status of object storage.
type RisingWaveObjectStorageStatus struct {
	Type ObjectStorageType `json:"type"`
}

// RisingWaveStoragesStatus is the status of external storages.
type RisingWaveStoragesStatus struct {
	Meta   RisingWaveMetaStorageStatus   `json:"meta"`
	Object RisingWaveObjectStorageStatus `json:"object"`
}

// RisingWaveStatus is the status of RisingWave.
type RisingWaveStatus struct {
	// ObservedGeneration is the observed generation by controller. It will be updated
	// when controller observes the changes on the spec and going to sync the subresources.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ComponentReplicas is the replica/running status of components.
	ComponentReplicas RisingWaveComponentsReplicasStatus `json:"componentReplicas,omitempty"`

	// Conditions are the conditions of the RisingWave.
	// +listType=map
	// +listMapKey=type
	// +patchMergeKey=type
	// +patchStrategy=merge,retainKeys
	Conditions []RisingWaveCondition `json:"conditions,omitempty"`

	// Storages are the status of the external storages.
	Storages RisingWaveStoragesStatus `json:"storages,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all;streaming
// +kubebuilder:printcolumn:name="RUNNING",type=string,JSONPath=`.status.conditions[?(@.type=="Running")].status`
// +kubebuilder:printcolumn:name="STORAGE(META)",type=string,JSONPath=`.status.storages.meta.type`
// +kubebuilder:printcolumn:name="STORAGE(OBJECT)",type=string,JSONPath=`.status.storages.object.type`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`
type RisingWave struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RisingWaveSpec   `json:"spec,omitempty"`
	Status RisingWaveStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RisingWaveList contains a list of RisingWave.
type RisingWaveList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RisingWave `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RisingWave{}, &RisingWaveList{})
}
