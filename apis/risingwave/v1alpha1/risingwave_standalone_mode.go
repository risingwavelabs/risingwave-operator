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
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RisingWaveStandaloneComponent contains the spec for standalone component.
type RisingWaveStandaloneComponent struct {
	// LogLevel controls the log level of the running nodes. It can be in any format that the underlying component supports,
	// e.g., in the RUST_LOG format for Rust programs. Defaults to INFO.
	// +kubebuilder:default=INFO
	LogLevel string `json:"logLevel,omitempty"`

	// DisallowPrintStackTraces determines if the stack traces are allowed to print when panic happens. This options applies
	// to both Rust and Java programs. Defaults to false.
	// +optional
	DisallowPrintStackTraces *bool `json:"disallowPrintStackTraces,omitempty"`

	// RestartAt is the time that the Pods under the group should be restarted. Setting a value on this field will
	// trigger a full recreation of the Pods. Defaults to nil.
	RestartAt *metav1.Time `json:"restartAt,omitempty"`

	// Replicas is the number of the standalone Pods. Maximum is 1. Defaults to 1.
	// +kubebuilder:default=1
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=1
	Replicas int32 `json:"replicas,omitempty"`

	// Upgrade strategy for the components. By default, it is the same as the
	// workload's default strategy that the component is deployed with.
	// Note: the maxSurge will not take effect for the compute component.
	// +optional
	// +patchStrategy=retainKeys
	UpgradeStrategy RisingWaveNodeGroupUpgradeStrategy `json:"upgradeStrategy,omitempty"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,9,opt,name=minReadySeconds"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// +optional
	VolumeClaimTemplates []PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent
	// volume claims created from volumeClaimTemplates. By default, all persistent
	// volume claims are created as needed and retained until manually deleted. This
	// policy allows the lifecycle to be altered, for example by deleting persistent
	// volume claims when their stateful set is deleted, or when their pod is scaled
	// down. This requires the StatefulSetAutoDeletePVC feature gate to be enabled,
	// which is alpha.
	// +optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaimRetentionPolicy"`

	// The maximum time in seconds for a deployment to make progress before it
	// is considered to be failed. The deployment controller will continue to
	// process failed deployments and a condition with a ProgressDeadlineExceeded
	// reason will be surfaced in the deployment status. Note that progress will
	// not be estimated during the time a deployment is paused. Defaults to 600s.
	ProgressDeadlineSeconds *int32 `json:"progressDeadlineSeconds,omitempty" protobuf:"varint,9,opt,name=progressDeadlineSeconds"`

	// Template tells how the Pod should be started. It is an optional field. If it's empty, then the pod template in
	// the first-level fields under spec will be used.
	// +optional
	Template RisingWaveNodePodTemplate `json:"template,omitempty"`
}
