// Copyright 2023 RisingWave Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	UpgradeStrategy RisingWaveNodeGroupUpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Resources of the RisingWave component.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// A map of labels describing the nodes to be scheduled on.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

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
	Metadata PartialObjectMeta `json:"metadata,omitempty"`

	// List of environment variables to set in the container.
	// Cannot be updated.
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,7,rep,name=env"`

	// List of sources to populate environment variables in the container.
	// The keys defined within a source must be a C_IDENTIFIER. All invalid keys
	// will be reported as an event when the container is starting. When a key exists in multiple
	// sources, the value associated with the last source will take precedence.
	// Values defined by an Env with a duplicate key will take precedence.
	// Cannot be updated.
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
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

// RisingWaveComponentMeta is the spec describes the meta component.
type RisingWaveComponentMeta struct {
	// Deprecated
	// The time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Deprecated: Use NodeGroups instead.
	// Groups of Pods of meta component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// RisingWaveComponentFrontend is the spec describes the frontend component.
type RisingWaveComponentFrontend struct {
	// Deprecated
	// The time that the Pods of frontend that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Deprecated: Use NodeGroups instead.
	// Groups of Pods of frontend component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// RisingWaveComponentCompute is the spec describes the compute component.
type RisingWaveComponentCompute struct {
	// Deprecated
	// The time that the Pods of compute that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Deprecated: Use NodeGroups instead.
	// Groups of Pods of compute component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComputeGroup `json:"groups,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// RisingWaveComponentCompactor is the spec describes the compactor component.
type RisingWaveComponentCompactor struct {
	// Deprecated
	// The time that the Pods of compactor that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Deprecated: Use NodeGroups instead.
	// Groups of Pods of compactor component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// RisingWaveComponentConnector is the spec describes the connector component.
type RisingWaveComponentConnector struct {
	// Deprecated
	// The time that the Pods of connector that should be restarted. Setting a value on this
	// field will trigger a recreation of all Pods of this component.
	// +optional
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Deprecated: Use NodeGroups instead.
	// Groups of Pods of compactor component.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	Groups []RisingWaveComponentGroup `json:"groups,omitempty"`

	// NodeGroups of the component deployment.
	// +optional
	// +listType=map
	// +listMapKey=name
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	NodeGroups []RisingWaveNodeGroup `json:"nodeGroups,omitempty"`
}

// RisingWaveComponentsSpec is the spec for RisingWave components.
type RisingWaveComponentsSpec struct {
	// Meta contains configuration of the meta component.
	Meta RisingWaveComponentMeta `json:"meta,omitempty"`

	// Frontend contains configuration of the frontend component.
	Frontend RisingWaveComponentFrontend `json:"frontend,omitempty"`

	// Compute contains configuration of the compute component.
	Compute RisingWaveComponentCompute `json:"compute,omitempty"`

	// Compactor contains configuration of the compactor component.
	Compactor RisingWaveComponentCompactor `json:"compactor,omitempty"`

	// Connector contains configuration of the connector component.
	Connector RisingWaveComponentConnector `json:"connector,omitempty"`
}
