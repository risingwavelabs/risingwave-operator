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

package v1alpha1

import (
	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// RisingWaveUpgradeStrategyType is the type of upgrade strategies used in RisingWave.
type RisingWaveUpgradeStrategyType string

// Valid values of RisingWaveUpgradeStrategyType.
const (
	RisingWaveUpgradeStrategyTypeRecreate          RisingWaveUpgradeStrategyType = "Recreate"
	RisingWaveUpgradeStrategyTypeRollingUpdate     RisingWaveUpgradeStrategyType = "RollingUpdate"
	RisingWaveUpgradeStrategyTypeInPlaceIfPossible RisingWaveUpgradeStrategyType = "InPlaceIfPossible"
	RisingWaveUpgradeStrategyTypeInPlaceOnly       RisingWaveUpgradeStrategyType = "InPlaceOnly"
)

// RisingWaveRollingUpdate is the spec to define rolling update strategies.
type RisingWaveRollingUpdate struct {
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding down.
	// Defaults to 25%.
	// +optional
	// +kubebuilder:default="25%"
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"bytes,1,opt,name=maxUnavailable"`

	// Partition is the desired number of pods in old revisions.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up by default.
	// It means when partition is set during pods updating, (replicas - partition value) number of pods will be updated.
	// Default value is 0.
	// +optional
	// +kubebuilder:default=0
	Partition *intstr.IntOrString `json:"partition,omitempty"`

	// The maximum number of pods that can be scheduled above the desired replicas during update or specified delete.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 0.
	// +optional
	// +kubebuilder:default=0
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
}

// RisingWaveUpgradeStrategy is the spec of upgrade strategy used by RisingWave.
type RisingWaveUpgradeStrategy struct {
	// Type of upgrade. Can be "Recreate" or "RollingUpdate". Default is RollingUpdate.
	// +optional
	// +kubebuilder:default=RollingUpdate
	// +kubebuilder:validation:Enum=Recreate;RollingUpdate;InPlaceIfPossible;InPlaceOnly
	Type RisingWaveUpgradeStrategyType `json:"type,omitempty"`

	// Rolling update config params. Present only if DeploymentStrategyType = RollingUpdate.
	// +optional
	RollingUpdate *RisingWaveRollingUpdate `json:"rollingUpdate,omitempty"`

	// InPlaceUpdateStrategy contains strategies for in-place update.
	// +optional
	InPlaceUpdateStrategy *kruisepubs.InPlaceUpdateStrategy `json:"inPlaceUpdateStrategy,omitempty"`
}

// RisingWaveComponentGroupTemplate is the common deployment template for groups of each component.
// Currently, we use the common template for meta/frontend/compactor.
type RisingWaveComponentGroupTemplate struct {
	// Image for RisingWave component.
	// +optional
	Image string `json:"image,omitempty"`

	// Pull policy of the RisingWave image. The default value is the same as the
	// default of Kubernetes.
	// +optional
	// +kubebuilder:default=IfNotPresent
	// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// Secrets for pulling RisingWave images.
	// +optional
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	// Upgrade strategy for the components. By default, it is the same as the
	// workload's default strategy that the component is deployed with.
	// Note: the maxSurge will not take effect for the compute component.
	// +optional
	// +patchStrategy=retainKeys
	UpgradeStrategy RisingWaveUpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Resources of the RisingWave component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// A map of labels describing the nodes to be scheduled on.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Base template for Pods of RisingWave. By default, there's no such template
	// and the controller will set all unrelated fields to the default value.
	// +optional
	PodTemplate *string `json:"podTemplate,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// If specified, indicates the pod's priority. "system-node-critical" and
	// "system-cluster-critical" are two special keywords which indicate the
	// highest priorities with the former being the highest priority. Any other
	// name must be defined by creating a PriorityClass object with that name.
	// If not specified, the pod priority will be default or zero if there is no
	// default.
	// +optional
	PriorityClassName string `json:"priorityClassName,omitempty"`

	// SecurityContext holds pod-level security attributes and common container settings.
	// Optional: Defaults to empty.  See type description for default values of each field.
	// +optional
	SecurityContext *corev1.PodSecurityContext `json:"securityContext,omitempty"`

	// Specifies the DNS parameters of a pod.
	// Parameters specified here will be merged to the generated DNS
	// configuration based on DNSPolicy.
	// +optional
	DNSConfig *corev1.PodDNSConfig `json:"dnsConfig,omitempty"`

	// Optional duration in seconds the pod needs to terminate gracefully. May be decreased in delete request.
	// Value must be non-negative integer. The value zero indicates stop immediately via
	// the kill signal (no opportunity to shut down).
	// If this value is nil, the default grace period will be used instead.
	// The grace period is the duration in seconds after the processes running in the pod are sent
	// a termination signal and the time when the processes are forcibly halted with a kill signal.
	// Set this value longer than the expected cleanup time for your process.
	// Defaults to 30 seconds.
	// +optional
	TerminationGracePeriodSeconds *int64 `json:"terminationGracePeriodSeconds,omitempty"`

	// metadata of the RisingWave's Pods.
	// +optional
	Metadata RisingWavePodTemplatePartialObjectMeta `json:"metadata,omitempty"`
}

// RisingWaveComponentGroup is the common deployment group of each component. Currently, we use
// this group for meta/frontend/compactor.
type RisingWaveComponentGroup struct {
	// Name of the group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Replicas of Pods in this group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`

	// The component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there are
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	*RisingWaveComponentGroupTemplate `json:",inline"`
}

// RisingWaveComputeGroupTemplate is the group template for component compute, which supports specifying
// the volume mounts on compute Pods. The volumes should be either local or defined in the storages.
type RisingWaveComputeGroupTemplate struct {
	// The component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there're
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	RisingWaveComponentGroupTemplate `json:",inline"`

	// Volumes to be mounted on the Pods.
	// +optional
	// +patchMergeKey=mountPath
	// +patchStrategy=merge
	VolumeMounts []corev1.VolumeMount `json:"volumeMounts,omitempty"`
}

// RisingWaveComputeGroup is the group for component compute.
type RisingWaveComputeGroup struct {
	// Name of the group.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Replicas of Pods in this group.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=0
	Replicas int32 `json:"replicas,omitempty"`

	// The component template describes how it would be deployed. It is an optional
	// field and the Pods are going to be deployed with the template defined in global. If there're
	// values defined in this template, it will be merged into the global template and then be used
	// for deployment.
	// +optional
	*RisingWaveComputeGroupTemplate `json:",inline"`
}

// RisingWaveComponentCommonPorts are the common ports that components need to listen.
type RisingWaveComponentCommonPorts struct {
	// Service port of the component. For each component,
	// the 'service' has different meanings. It's an optional field and if it's left out, a
	// default port (varies among components) will be used.
	// +optional
	// +kubebuilder:validation:Minimum=1
	ServicePort int32 `json:"service,omitempty"`

	// Metrics port of the component. It always serves the metrics in
	// Prometheus format.
	// +optional
	// +kubebuilder:validation:Minimum=1
	MetricsPort int32 `json:"metrics,omitempty"`
}

// RisingWaveComponentMetaPorts are the ports of component meta.
type RisingWaveComponentMetaPorts struct {
	RisingWaveComponentCommonPorts `json:",inline"`

	// Dashboard port of the meta, a default value of 8080 will be
	// used if not specified.
	// +optional
	// +kubebuilder:validation:Minimum=1
	DashboardPort int32 `json:"dashboard,omitempty"`
}

// RisingWaveComponentMeta is the spec describes the meta component.
type RisingWaveComponentMeta struct {
	// The time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports to be listened by the meta Pods.
	// +optional
	Ports RisingWaveComponentMetaPorts `json:"ports,omitempty"`

	// Groups of Pods of meta component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

// RisingWaveComponentFrontend is the spec describes the frontend component.
type RisingWaveComponentFrontend struct {
	// The time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports to be listened by the frontend Pods.
	// +optional
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups of Pods of frontend component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

// RisingWaveComponentCompute is the spec describes the compute component.
type RisingWaveComponentCompute struct {
	// The time that the Pods of compute that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports to be listened by compute Pods.
	// +optional
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups of Pods of compute component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComputeGroup `json:"groups,omitempty"`
}

// RisingWaveComponentCompactor is the spec describes the compactor component.
type RisingWaveComponentCompactor struct {
	// The time that the Pods of compactor that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports to be listened by compactor Pods.
	// +optional
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups of Pods of compactor component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

// RisingWaveComponentConnector is the spec describes the connector component.
type RisingWaveComponentConnector struct {
	// The time that the Pods of connector that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Ports to be listened by compactor Pods.
	// +optional
	Ports RisingWaveComponentCommonPorts `json:"ports,omitempty"`

	// Groups of Pods of compactor component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`
}

// RisingWaveComponentsSpec is the spec describes the components of RisingWave.
type RisingWaveComponentsSpec struct {
	// Meta component spec.
	Meta RisingWaveComponentMeta `json:"meta,omitempty"`

	// Frontend component spec.
	Frontend RisingWaveComponentFrontend `json:"frontend,omitempty"`

	// Compute component spec.
	Compute RisingWaveComponentCompute `json:"compute,omitempty"`

	// Compactor component.
	Compactor RisingWaveComponentCompactor `json:"compactor,omitempty"`

	// Connector component spec.
	Connector RisingWaveComponentConnector `json:"connector,omitempty"`
}

// RisingWaveMetaStorageEtcd is the etcd storage for the meta component.
type RisingWaveMetaStorageEtcd struct {
	// Endpoint of etcd. It must be provided.
	Endpoint string `json:"endpoint"`

	// Secret contains the credentials of access the etcd, it must contain the following keys:
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

	// Endpoint of the etcd service for storing the metadata.
	// +optional
	Etcd *RisingWaveMetaStorageEtcd `json:"etcd,omitempty"`
}

// RisingWaveObjectStorageMinIO is the details of MinIO storage for compute and compactor components.
type RisingWaveObjectStorageMinIO struct {
	// Secret contains the credentials to access the MinIO service. It must contain the following keys:
	//   * username
	//   * password
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Endpoint of the MinIO service.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Bucket of the MinIO service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`
}

// RisingWaveObjectStorageS3 is the details of AWS S3 storage for compute and compactor components.
type RisingWaveObjectStorageS3 struct {
	// Secret contains the credentials to access the AWS S3 service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Bucket of the AWS S3 service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Region of AWS S3 service. It is an optional field that overrides the `Region` key from the secret.
	// Specifying the region here makes a guarantee that it won't be changed anymore.
	Region string `json:"region,omitempty"`

	// Endpoint of the AWS (or other vendor's S3-compatible) service. Leave it empty when using AWS S3 service.
	// You can reference the `REGION` and `BUCKET` variables in the endpoint with `${BUCKET}` and `${REGION}`, e.g.,
	//   s3.${REGION}.amazonaws.com
	//   ${BUCKET}.s3.${REGION}.amazonaws.com
	// +optional
	// +kubebuilder:validation:Pattern="^(?:https://)?(?:[^/.\\s]+\\.)*(?:[^/\\s]+)*$"
	Endpoint string `json:"endpoint,omitempty"`

	// VirtualHostedStyle indicates to use a virtual hosted endpoint when endpoint is specified. The operator automatically
	// adds the bucket prefix for you if this is enabled. Be careful about doubly using the style by specifying an endpoint
	// of virtual hosted style as well as enabling this.
	VirtualHostedStyle bool `json:"virtualHostedStyle,omitempty"`
}

// RisingWaveObjectStorageAliyunOSS is the details of Aliyun OSS storage (S3 compatible) for compute and compactor components.
type RisingWaveObjectStorageAliyunOSS struct {
	// Secret contains the credentials to access the Aliyun OSS service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Region of Aliyun OSS service. It is an optional field that overrides the `Region` key from the secret.
	// Specifying the region here makes a guarantee that it won't be changed anymore.
	Region string `json:"region,omitempty"`

	// Bucket of the Aliyun OSS service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// InternalEndpoint indicates if we use the internal endpoint to access Aliyun OSS, which is
	// only available in the internal network.
	InternalEndpoint bool `json:"internalEndpoint,omitempty"`
}

// RisingWaveObjectStorage is the object storage for compute and compactor components.
type RisingWaveObjectStorage struct {
	// Memory indicates to store the data in memory. It's only for test usage and strongly discouraged to
	// be used in production.
	// +optional
	Memory *bool `json:"memory,omitempty"`

	// MinIO storage spec.
	// +optional
	MinIO *RisingWaveObjectStorageMinIO `json:"minio,omitempty"`

	// S3 storage spec.
	// +optional
	S3 *RisingWaveObjectStorageS3 `json:"s3,omitempty"`

	// AliyunOSS storage spec.
	// +optional
	AliyunOSS *RisingWaveObjectStorageAliyunOSS `json:"aliyunOSS,omitempty"`
}

// RisingWaveStoragesSpec is the storages spec.
type RisingWaveStoragesSpec struct {
	// Storage spec for meta.

	Meta RisingWaveMetaStorage `json:"meta,omitempty"`

	// Storage spec for object storage.
	Object RisingWaveObjectStorage `json:"object,omitempty"`

	// The persistent volume claim templates for the compute component. PVCs declared here
	// can be referenced in the groups of compute component.
	// +optional
	PVCTemplates []corev1.PersistentVolumeClaim `json:"pvcTemplates,omitempty"`
}

// RisingWaveConfigurationSpec is the configuration spec.
type RisingWaveConfigurationSpec struct {
	// The reference to a key in a config map that contains the base config for RisingWave.
	// It's an optional field and can be left out. If not specified, a default config is going to be used.
	// +optional
	ConfigMap *corev1.ConfigMapKeySelector `json:"configmap,omitempty"`
}

// RisingWaveTLSConfigSecret is the secret reference that contains the key and cert for TLS.
type RisingWaveTLSConfigSecret struct {
	// Name of the secret.
	// +optional
	Name string `json:"name,omitempty"`

	// The key of the TLS key. A default value of "tls.key" will be used if not specified.
	// +optional
	// +kubebuilder:default=tls.key
	Key string `json:"key,omitempty"`

	// The key of the TLS cert. A default value of "tls.crt" will be used if not specified.
	// +optional
	// +kubebuilder:default=tls.crt
	Cert string `json:"cert,omitempty"`
}

// RisingWaveTLSConfig is the TLS config of RisingWave.
type RisingWaveTLSConfig struct {
	// Enabled indicates if TLS is enabled on RisingWave.
	Enabled bool `json:"enabled,omitempty"`

	// Secret contains the TLS config. If TLS is enabled, the secret
	// must be provided.
	// +optional.
	Secret RisingWaveTLSConfigSecret `json:"secret,omitempty"`
}

// RisingWaveSecuritySpec is the security spec.
type RisingWaveSecuritySpec struct {
	// TLS config of RisingWave.
	// +optional
	// +patchStrategy=retainKeys
	TLS *RisingWaveTLSConfig `json:"tls,omitempty"`
}

// RisingWaveGlobalReplicas are the replicas of each component, declared in global scope.
type RisingWaveGlobalReplicas struct {
	// Replicas of meta component. Replicas specified here is in a default group (with empty name '').
	// +optional
	// +kubebuilder:validation:Minimum=0
	Meta int32 `json:"meta,omitempty"`

	// Replicas of frontend component. Replicas specified here is in a default group (with empty name '').
	// +optional
	// +kubebuilder:validation:Minimum=0
	Frontend int32 `json:"frontend,omitempty"`

	// Replicas of compute component. Replicas specified here is in a default group (with empty name '').
	// +optional
	// +kubebuilder:validation:Minimum=0
	Compute int32 `json:"compute,omitempty"`

	// Replicas of compactor component. Replicas specified here is in a default group (with empty name '').
	// +optional
	// +kubebuilder:validation:Minimum=0
	Compactor int32 `json:"compactor,omitempty"`

	// Replicas of connector component. Replicas specified here is in a default group (with empty name '').
	// +optional
	// +kubebuilder:validation:Minimum=0
	Connector int32 `json:"connector,omitempty"`
}

// RisingWaveGlobalSpec is the global spec.
type RisingWaveGlobalSpec struct {
	// Global template for RisingWave that all components share.
	RisingWaveComponentGroupTemplate `json:",inline"`

	// Replicas of each component in default groups.
	// +optional
	// +patchStrategy=retainKeys
	Replicas RisingWaveGlobalReplicas `json:"replicas,omitempty"`

	// Service type of the frontend service.
	// +optional
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`

	// Service metadata of the frontend service.
	// +optional
	ServiceMeta RisingWavePodTemplatePartialObjectMeta `json:"serviceMetadata,omitempty"`
}

// RisingWaveSpec is the overall spec.
type RisingWaveSpec struct {
	// The spec of a shared template for components and a global scope of replicas.
	Global RisingWaveGlobalSpec `json:"global,omitempty"`

	// The spec of meta storage, object storage for compute and compactor, and PVC templates for compute.
	Storages RisingWaveStoragesSpec `json:"storages,omitempty"`

	// The spec of ports and some controllers (such as `restartAt`) of each component,
	// as well as an advanced concept called `group` to override the global template and create groups
	// of Pods, e.g., deployment in hybrid-arch cluster.
	Components RisingWaveComponentsSpec `json:"components,omitempty"`

	// The spec of security configurations, such as TLS config.
	Security RisingWaveSecuritySpec `json:"security,omitempty"`

	// The spec of configuration template for RisingWave.
	Configuration RisingWaveConfigurationSpec `json:"configuration,omitempty"`

	// Flag to indicate if OpenKruise should be enabled for components.
	// If enabled, CloneSets will be used for meta/frontend/compactor nodes
	// and Advanced StateFulSets will be used for compute nodes.
	// +optional
	// +kubebuilder:default=false
	EnableOpenKruise *bool `json:"enableOpenKruise,omitempty"`

	// Flag to indicate if a default ServiceMonitor (from Prometheus operator) should be created by the controller.
	// False and an empty value means the ServiceMonitor won't be created automatically. But even if it's set to true,
	// the controller will determine if it can create the resource by checking if the CRDs are installed.
	// +optional
	EnableDefaultServiceMonitor *bool `json:"enableDefaultServiceMonitor,omitempty"`
}

// ComponentGroupReplicasStatus are the running status of Pods in group.
type ComponentGroupReplicasStatus struct {
	// Name of the group.
	Name string `json:"name"`

	// Target replicas of the group.
	Target int32 `json:"target"`

	// Running replicas in the group.
	Running int32 `json:"running"`

	// Existence status of the group.
	Exists bool `json:"exists,omitempty"`
}

// ComponentReplicasStatus are the running status of Pods of the component.
type ComponentReplicasStatus struct {
	// Total target replicas of the component.
	Target int32 `json:"target"`

	// Total running replicas of the component.
	Running int32 `json:"running"`

	// List of running status of each group.
	Groups []ComponentGroupReplicasStatus `json:"groups,omitempty"`
}

// RisingWaveComponentsReplicasStatus is the running status of components.
type RisingWaveComponentsReplicasStatus struct {
	// Running status of meta.
	Meta ComponentReplicasStatus `json:"meta"`

	// Running status of frontend.
	Frontend ComponentReplicasStatus `json:"frontend"`

	// Running status of compute.
	Compute ComponentReplicasStatus `json:"compute"`

	// Running status of compactor.
	Compactor ComponentReplicasStatus `json:"compactor"`

	// Running status of connector.
	Connector ComponentReplicasStatus `json:"connector"`
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
	MetaStorageTypeMemory  MetaStorageType = "Memory"
	MetaStorageTypeEtcd    MetaStorageType = "Etcd"
	MetaStorageTypeUnknown MetaStorageType = "Unknown"
)

// RisingWaveMetaStorageStatus is the status of meta storage.
type RisingWaveMetaStorageStatus struct {
	Type MetaStorageType `json:"type"`
}

// ObjectStorageType is the type name of object storage.
type ObjectStorageType string

// These are valid values of ObjectStorageType.
const (
	ObjectStorageTypeMemory    ObjectStorageType = "Memory"
	ObjectStorageTypeMinIO     ObjectStorageType = "MinIO"
	ObjectStorageTypeS3        ObjectStorageType = "S3"
	ObjectStorageTypeAliyunOSS ObjectStorageType = "AliyunOSS"
	ObjectStorageTypeUnknown   ObjectStorageType = "Unknown"
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

// RisingWaveScaleViewLockGroupLock is the lock record of RisingWaveScaleView.
type RisingWaveScaleViewLockGroupLock struct {
	// Group name.
	Name string `json:"name"`

	// Locked replica value.
	Replicas int32 `json:"replicas,omitempty"`
}

// RisingWaveScaleViewLock is a lock record for RisingWaveScaleViews. For example, if there's a RisingWaveScaleView
// targets the current RisingWave, the controller will try to create a new RisingWaveScaleViewLock with the name, uid,
// target component, generation, and the replicas of targeting groups of the RisingWaveScaleView. After the record is set,
// the validation webhook will reject any updates on the replicas of any targeting group that doesn't equal the
// replicas recorded, which makes it a lock similar thing.
type RisingWaveScaleViewLock struct {
	// Name of the owned RisingWaveScaleView object.
	Name string `json:"name"`

	// UID of the owned RisingWaveScaleView object.
	UID types.UID `json:"uid"`

	// Component of the lock.
	Component string `json:"component"`

	// Generation of the lock.
	Generation int64 `json:"generation"`

	// Group locks.
	// +listType=map
	// +listMapKey=name
	GroupLocks []RisingWaveScaleViewLockGroupLock `json:"groupLocks,omitempty"`
}

// RisingWaveStatus is the status of RisingWave.
type RisingWaveStatus struct {
	// Observed generation by controller. It will be updated
	// when controller observes the changes on the spec and going to sync the subresources.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Replica status of components.
	ComponentReplicas RisingWaveComponentsReplicasStatus `json:"componentReplicas,omitempty"`

	// Conditions of the RisingWave.
	// +listType=map
	// +listMapKey=type
	// +patchMergeKey=type
	// +patchStrategy=merge,retainKeys
	Conditions []RisingWaveCondition `json:"conditions,omitempty"`

	// Status of the external storages.
	Storages RisingWaveStoragesStatus `json:"storages,omitempty"`

	// Scale view locks.
	// +listType=map
	// +listMapKey=name
	ScaleViews []RisingWaveScaleViewLock `json:"scaleViews,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all;streaming
// +kubebuilder:printcolumn:name="RUNNING",type=string,JSONPath=`.status.conditions[?(@.type=="Running")].status`
// +kubebuilder:printcolumn:name="STORAGE(META)",type=string,JSONPath=`.status.storages.meta.type`
// +kubebuilder:printcolumn:name="STORAGE(OBJECT)",type=string,JSONPath=`.status.storages.object.type`
// +kubebuilder:printcolumn:name="AGE",type=date,JSONPath=`.metadata.creationTimestamp`

// RisingWave is the struct for RisingWave object.
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
