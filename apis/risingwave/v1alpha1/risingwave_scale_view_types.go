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
)

type RisingWaveScaleViewStatusSpecTargetRef struct {
}

type RisingWaveScaleViewSpec struct {
	TargetRef RisingWaveScaleViewStatusSpecTargetRef `json:"targetRef,omitempty"`
}

type RisingWaveScaleViewStatusTargetRef struct {
}

type RisingWaveScaleViewStatus struct {
	TargetRef RisingWaveScaleViewStatusTargetRef `json:"targetRef,omitempty"`
	Replicas  int64                              `json:"replicas,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
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
