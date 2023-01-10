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

// RisingWavePodTemplatePartialObjectMeta is the spec for metadata templates.
type RisingWavePodTemplatePartialObjectMeta struct {
	// Labels of the object.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations of the object.
	Annotations map[string]string `json:"annotations,omitempty"`
}

// RisingWavePodTemplateSpec is the spec of RisingWavePodTemplate.
type RisingWavePodTemplateSpec struct {
	RisingWavePodTemplatePartialObjectMeta `json:"metadata,omitempty"`

	Spec corev1.PodSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=rwpt,categories=all;streaming

// RisingWavePodTemplate is the struct for RisingWavePodTemplate object.
type RisingWavePodTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Template          RisingWavePodTemplateSpec `json:"template,omitempty"`
}

// +kubebuilder:object:root=true

// RisingWavePodTemplateList contains a list of RisingWavePodTemplate.
type RisingWavePodTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RisingWavePodTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RisingWavePodTemplate{}, &RisingWavePodTemplateList{})
}
