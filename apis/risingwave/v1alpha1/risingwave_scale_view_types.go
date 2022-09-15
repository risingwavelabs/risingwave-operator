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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// RisingWaveScaleViewTargetRef is the reference of the target RisingWave.
type RisingWaveScaleViewTargetRef struct {
	// Name of the RisingWave object.
	Name string `json:"name"`

	// Component name. Must be one of meta, frontend, compute, and compactor.
	// +kubebuilder:validation:Enum=meta;frontend;compute;compactor
	Component string `json:"component"`

	// UID of the target RisingWave object. Should be set by the mutating webhook.
	UID types.UID `json:"uid,omitempty"`
}

// RisingWaveScaleViewSpecScalePolicyConstraints is the constraints of replicas in scale policy.
type RisingWaveScaleViewSpecScalePolicyConstraints struct {
	// Maximum value of the replicas.
	// +kubebuilder:validation:Maximum=1000
	Max int32 `json:"max,omitempty"`
}

// RisingWaveScaleViewSpecScalePolicy is the scale policy of a group.
type RisingWaveScaleViewSpecScalePolicy struct {
	// Group name.
	Group string `json:"group"`

	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=10

	// 0-10, optional. The groups will be sorted by the priority and the current replicas.
	// The higher it is, the more replicas of the target group will be considered kept, i.e. scale out first, scale in last.
	// +optional
	Priority int32 `json:"priority,omitempty"`

	// MaxReplicas is the limit of the replicas.
	// +kubebuilder:validation:Maximum=5000
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

// RisingWaveScaleViewSpec is the spec of RisingWaveScaleView.
type RisingWaveScaleViewSpec struct {
	// Reference of the target RisingWave.
	TargetRef RisingWaveScaleViewTargetRef `json:"targetRef,omitempty"`

	// Desired replicas.
	Replicas int32 `json:"replicas,omitempty"`

	// Serialized label selector. Would be set by the webhook.
	LabelSelector string `json:"labelSelector,omitempty"`

	// An array of groups and the policies for scale, optional and empty means the default group with the default policy.
	// +listType=map
	// +listMapKey=group
	ScalePolicy []RisingWaveScaleViewSpecScalePolicy `json:"scalePolicy,omitempty"`
}

// RisingWaveScaleViewStatus is the status of RisingWaveScaleView.
type RisingWaveScaleViewStatus struct {
	// Running replicas.
	Replicas *int32 `json:"replicas,omitempty"`

	// Lock status.
	Locked bool `json:"locked,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector
// +kubebuilder:printcolumn:name="TARGET",type=string,JSONPath=`.spec.targetRef.name`
// +kubebuilder:printcolumn:name="COMPONENT",type=string,JSONPath=`.spec.targetRef.component`
// +kubebuilder:printcolumn:name="READY",type=integer,JSONPath=`.status.replicas`
// +kubebuilder:printcolumn:name="REPLICAS",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="LOCKED",type=boolean,JSONPath=`.status.locked`
// +kubebuilder:resource:shortName=rwsv,categories=all;streaming

type RisingWaveScaleView struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RisingWaveScaleViewSpec   `json:"spec,omitempty"`
	Status RisingWaveScaleViewStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type RisingWaveScaleViewList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []RisingWaveScaleView `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RisingWaveScaleView{}, &RisingWaveScaleViewList{})
}
