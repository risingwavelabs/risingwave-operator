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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// RisingWaveConfigurationSpec is the configuration spec.
type RisingWaveConfigurationSpec struct {
	RisingWaveNodeConfiguration `json:",inline"`
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
}

// RisingWaveComponentsSpec is the spec for RisingWave components.
type RisingWaveComponentsSpec struct {
	// Standalone contains configuration of the standalone component.
	Standalone *RisingWaveStandaloneComponent `json:"standalone,omitempty"`

	// Meta contains configuration of the meta component.
	Meta RisingWaveComponent `json:"meta,omitempty"`

	// Frontend contains configuration of the frontend component.
	Frontend RisingWaveComponent `json:"frontend,omitempty"`

	// Compute contains configuration of the compute component.
	Compute RisingWaveComponent `json:"compute,omitempty"`

	// Compactor contains configuration of the compactor component.
	Compactor RisingWaveComponent `json:"compactor,omitempty"`
}

// RisingWaveSpec is the overall spec.
type RisingWaveSpec struct {
	// The spec of ports and some controllers (such as `restartAt`) of each component,
	// as well as an advanced concept called `group` to override the global template and create groups
	// of Pods, e.g., deployment in hybrid-arch cluster.
	Components RisingWaveComponentsSpec `json:"components,omitempty"`

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

	// Flag to indicate if full kubernetes address should be enabled for components.
	// If enabled, address will be [<pod>.]<service>.<namespace>.svc. Otherwise, it will be [<pod>.]<service>.
	// Enabling this flag on existing RisingWave will cause incompatibility.
	// +optional
	// +kubebuilder:default=false
	EnableFullKubernetesAddr *bool `json:"enableFullKubernetesAddr,omitempty"`

	// Flag to control whether to deploy in standalone mode or distributed mode. If standalone mode is used,
	// spec.components will be ignored. Standalone mode can be turned on/off dynamically.
	// +optional
	// +kubebuilder:default=false
	EnableStandaloneMode *bool `json:"enableStandaloneMode,omitempty"`

	// Flag to control whether to enable embedded serving mode. If enabled, the frontend nodes will be created
	// with embedded serving node enabled, and the compute nodes will serve streaming workload only.
	// +optional
	// +kubebuilder:default=false
	EnableEmbeddedServingMode *bool `json:"enableEmbeddedServingMode,omitempty"`

	// Flag to control whether to enable advertising with IP. If enabled, the meta and compute nodes will be advertised
	// with their IP addresses. This is useful when one wants to avoid the DNS resolution overhead and latency.
	EnableAdvertisingWithIP *bool `json:"enableAdvertisingWithIP,omitempty"`

	// Image for RisingWave component.
	Image string `json:"image"`

	// FrontendServiceType determines the service type of the frontend service. Defaults to ClusterIP.
	// +optional
	// +kubebuilder:default=ClusterIP
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	FrontendServiceType corev1.ServiceType `json:"frontendServiceType,omitempty"`

	// AdditionalFrontendServiceMetadata tells the operator to add the specified metadata onto the frontend Service.
	// Note that the system reserved labels and annotations are not valid and will be rejected by the webhook.
	AdditionalFrontendServiceMetadata PartialObjectMeta `json:"additionalFrontendServiceMetadata,omitempty"`

	// MetaStore determines which backend the meta store will use and the parameters for it. Defaults to memory.
	// But keep in mind that memory backend is not recommended in production.
	// +kubebuilder:default={memory: true}
	MetaStore RisingWaveMetaStoreBackend `json:"metaStore,omitempty"`

	// StateStore determines which backend the state store will use and the parameters for it. Defaults to memory.
	// But keep in mind that memory backend is not recommended in production.
	// +kubebuilder:default={memory: true}
	StateStore RisingWaveStateStoreBackend `json:"stateStore,omitempty"`

	// TLS configures the TLS/SSL certificates for SQL access.
	TLS *RisingWaveTLSConfiguration `json:"tls,omitempty"`

	// StandaloneMode determines which style of command-line args should be used for the standalone mode.
	// 0 - auto detect by image version, 1 - the old standalone mode, 2 - standalone mode V2 (single-node).
	// This is only for backward compatibility and will be deprecated in the future.
	// +kubebuilder:default=0
	StandaloneMode int32 `json:"standaloneMode,omitempty"`
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

	// Running status of standalone component.
	Standalone ComponentReplicasStatus `json:"standalone"`
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

// RisingWaveInternalStatus stores some internal status of RisingWave, such as internal states.
type RisingWaveInternalStatus struct {
	// StateStoreRootPath stores the root path of the state store data directory. It's for compatibility purpose and
	// should not be updated in most cases.
	StateStoreRootPath string `json:"stateStoreRootPath,omitempty"`
}

// RisingWaveStatus is the status of RisingWave.
type RisingWaveStatus struct {
	// Observed generation by controller. It will be updated
	// when controller observes the changes on the spec and going to sync the subresources.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Version of the Global Image
	Version string `json:"version,omitempty"`

	// Replica status of components.
	ComponentReplicas RisingWaveComponentsReplicasStatus `json:"componentReplicas,omitempty"`

	// Conditions of the RisingWave.
	// +listType=map
	// +listMapKey=type
	// +patchMergeKey=type
	// +patchStrategy=merge,retainKeys
	Conditions []RisingWaveCondition `json:"conditions,omitempty"`

	// Scale view locks.
	// +listType=map
	// +listMapKey=name
	ScaleViews []RisingWaveScaleViewLock `json:"scaleViews,omitempty"`

	// Internal status.
	Internal RisingWaveInternalStatus `json:"internal,omitempty"`

	// -----------------------------------v1alpha2 features ------------------------------------------ //

	// Status of the meta store.
	MetaStore RisingWaveMetaStoreStatus `json:"metaStore,omitempty"`

	// Status of the state store.
	StateStore RisingWaveStateStoreStatus `json:"stateStore,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rw,categories=all;streaming
// +kubebuilder:printcolumn:name="META STORE",type=string,JSONPath=`.status.metaStore.backend`
// +kubebuilder:printcolumn:name="STATE STORE",type=string,JSONPath=`.status.stateStore.backend`
// +kubebuilder:printcolumn:name="VERSION",type=string,JSONPath=`.status.version`
// +kubebuilder:printcolumn:name="RUNNING",type=string,JSONPath=`.status.conditions[?(@.type=="Running")].status`
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
