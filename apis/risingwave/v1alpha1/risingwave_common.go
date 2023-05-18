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

// PartialObjectMeta contains partial metadata of an object, including labels and annotations.
type PartialObjectMeta struct {
	// Labels of the object.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations of the object.
	Annotations map[string]string `json:"annotations,omitempty"`
}
