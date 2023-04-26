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

// RisingWaveMetaStoreBackendType is the type for the meta store backends.
type RisingWaveMetaStoreBackendType string

// All valid meta store backend types.
const (
	RisingWaveMetaStoreBackendTypeMemory  RisingWaveMetaStoreBackendType = "Memory"
	RisingWaveMetaStoreBackendTypeEtcd    RisingWaveMetaStoreBackendType = "Etcd"
	RisingWaveMetaStoreBackendTypeUnknown RisingWaveMetaStoreBackendType = "Unknown"
)

// RisingWaveMetaStoreStatus is the status of the meta store.
type RisingWaveMetaStoreStatus struct {
	// Backend type of the meta store.
	Backend RisingWaveMetaStoreBackendType `json:"backend,omitempty"`
}

// RisingWaveMetaStoreBackendEtcd is the collection of parameters for the etcd backend meta store.
type RisingWaveMetaStoreBackendEtcd struct {
	// Endpoint of etcd. It must be provided.
	Endpoint string `json:"endpoint"`

	// Secret contains the credentials of access the etcd, it must contain the following keys:
	//   * username
	//   * password
	// But it is an optional field. Empty value indicates etcd is available without authentication.
	// +optional
	Secret string `json:"secret,omitempty"`
}

// RisingWaveMetaStoreBackend is the collection of parameters for the meta store that RisingWave uses. Note that one
// and only one of the first-level fields could be set.
type RisingWaveMetaStoreBackend struct {
	// Memory indicates to store the metadata in memory. It is only for test usage and strongly
	// discouraged to be set in production. If one is using the memory storage for meta,
	// replicas will not work because they are not going to share the same metadata and any kinds
	// exit of the process will cause a permanent loss of the data.
	// +optional
	Memory *bool `json:"memory,omitempty"`

	// Endpoint of the etcd service for storing the metadata.
	// +optional
	Etcd *RisingWaveMetaStoreBackendEtcd `json:"etcd,omitempty"`
}
