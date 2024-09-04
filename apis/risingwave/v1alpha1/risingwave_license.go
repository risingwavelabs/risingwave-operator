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

const RisingWaveLicenseLicenseSecretKey = "license"

// RisingWaveLicense is the license configuration for RisingWave.
type RisingWaveLicense struct {
	// SecretName that contains the license. The license must be JWT formatted JSON and stored with key `license`.
	// +optional
	SecretName string `json:"secretName,omitempty"`
}
