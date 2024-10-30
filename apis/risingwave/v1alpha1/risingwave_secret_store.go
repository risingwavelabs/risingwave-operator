// Copyright 2024 RisingWave Labs
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

// RisingWaveSecretStorePrivateKeySecretReference is a reference to a secret that contains a private key.
type RisingWaveSecretStorePrivateKeySecretReference struct {
	// Name is the name of the secret.
	Name string `json:"name"`

	// Key is the key in the secret that contains the private key.
	Key string `json:"key"`
}

// RisingWaveSecretStorePrivateKey is a private key that can be stored in a secret or directly in the resource.
type RisingWaveSecretStorePrivateKey struct {
	// Value is the private key. It must be a 128-bit key encoded in hex. If this is set, SecretRef must be nil.
	// When the feature gate RandomSecretStorePrivateKey is enabled and neither is set, the private key will be
	// generated randomly.
	// +kubebuilder:validation:Pattern="^[0-9a-f]{32}$"
	// +optional
	Value *string `json:"value,omitempty"`

	// SecretRef is a reference to a secret that contains the private key. If this is set, Value must be nil.
	// Note that the value in the secret must be a 128-bit key encoded in hex.
	// +optional
	SecretRef *RisingWaveSecretStorePrivateKeySecretReference `json:"secretRef,omitempty"`
}

// RisingWaveSecretStore is the configuration of the secret store.
type RisingWaveSecretStore struct {
	// PrivateKey is the private key used to encrypt and decrypt the secrets.
	PrivateKey RisingWaveSecretStorePrivateKey `json:"privateKey,omitempty"`
}
