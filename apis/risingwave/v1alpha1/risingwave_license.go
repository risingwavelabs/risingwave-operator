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

const RisingWaveLicenseKeySecretKey = "licenseKey"

// RisingWaveLicenseKey is the license configuration for RisingWave.
type RisingWaveLicenseKey struct {
	// SecretName that contains the license. The license must be JWT formatted JSON.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// SecretKey to the license in the Secret above. Defaults to `licenseKey`.
	// +kubebuilder:default=licenseKey
	SecretKey string `json:"secretKey,omitempty"`

	// PassAsFile will pass the license as a file to the RisingWave process.
	// The feature is only available in the RisingWave v2.1 and later. See
	// https://github.com/risingwavelabs/risingwave/pull/18768 for more information.
	//
	// It's optional to set this field. If not set, the operator will deduce the value
	// based on the RisingWave version. But when it comes to non-semver versions, it's
	// recommended to set this field explicitly.
	// +optional
	PassAsFile *bool `json:"passAsFile,omitempty"`
}
