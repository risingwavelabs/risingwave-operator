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

// RisingWaveNodeConfigurationConfigMapSource refers to a ConfigMap where the RisingWave configuration is stored.
type RisingWaveNodeConfigurationConfigMapSource struct {
	// Name determines the ConfigMap to provide the configs RisingWave requests. It will be mounted on the Pods
	// directly. It the ConfigMap isn't provided, the controller will use empty value as the configs.
	// +optional
	Name string `json:"name,omitempty"`

	// Key to the configuration file. Defaults to `risingwave.toml`.
	// +kubebuilder:default=risingwave.toml
	// +optional
	Key string `json:"key,omitempty"`

	// Optional determines if the key must exist in the ConfigMap. Defaults to false.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

// RisingWaveNodeConfigurationSecretSource refers to a Secret where the RisingWave configuration is stored.
type RisingWaveNodeConfigurationSecretSource struct {
	// Name determines the Secret to provide the configs RisingWave requests. It will be mounted on the Pods
	// directly. It the Secret isn't provided, the controller will use empty value as the configs.
	// +optional
	Name string `json:"name,omitempty"`

	// Key to the configuration file. Defaults to `risingwave.toml`.
	// +kubebuilder:default=risingwave.toml
	// +optional
	Key string `json:"key,omitempty"`

	// Optional determines if the key must exist in the Secret. Defaults to false.
	// +optional
	Optional *bool `json:"optional,omitempty"`
}

// RisingWaveNodeConfiguration determines where the configurations are from, either ConfigMap, Secret, or raw string.
type RisingWaveNodeConfiguration struct {
	// ConfigMap where the `risingwave.toml` locates.
	ConfigMap *RisingWaveNodeConfigurationConfigMapSource `json:"configMap,omitempty"`

	// Secret where the `risingwave.toml` locates.
	Secret *RisingWaveNodeConfigurationSecretSource `json:"secret,omitempty"`
}
