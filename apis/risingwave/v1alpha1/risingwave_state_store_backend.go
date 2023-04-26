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

// RisingWaveStateStoreBackendType is the type for the state store backends.
type RisingWaveStateStoreBackendType string

// All valid state store backend types.
const (
	RisingWaveStateStoreBackendTypeMemory    RisingWaveStateStoreBackendType = "Memory"
	RisingWaveStateStoreBackendTypeMinIO     RisingWaveStateStoreBackendType = "MinIO"
	RisingWaveStateStoreBackendTypeS3        RisingWaveStateStoreBackendType = "S3"
	RisingWaveStateStoreBackendTypeHDFS      RisingWaveStateStoreBackendType = "HDFS"
	RisingWaveStateStoreBackendTypeWebHDFS   RisingWaveStateStoreBackendType = "WebHDFS"
	RisingWaveStateStoreBackendTypeGCS       RisingWaveStateStoreBackendType = "GCS"
	RisingWaveStateStoreBackendTypeAliyunOSS RisingWaveStateStoreBackendType = "AliyunOSS"
	RisingWaveStateStoreBackendTypeAzureBlob RisingWaveStateStoreBackendType = "AzureBlob"
	RisingWaveStateStoreBackendTypeUnknown   RisingWaveStateStoreBackendType = "Unknown"
)

// RisingWaveStateStoreStatus is the status of the state store.
type RisingWaveStateStoreStatus struct {
	// Backend type of the state store.
	Backend RisingWaveStateStoreBackendType `json:"backend,omitempty"`
}

// RisingWaveStateStoreBackendMinIO is the collection of parameters for the MinIO backend state store.
type RisingWaveStateStoreBackendMinIO struct {
	// Secret contains the credentials to access the MinIO service. It must contain the following keys:
	//   * username
	//   * password
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Endpoint of the MinIO service.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Bucket of the MinIO service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`
}

// RisingWaveStateStoreBackendS3 is the collection of parameters for the S3 backend state store.
type RisingWaveStateStoreBackendS3 struct {
	// Secret contains the credentials to access the AWS S3 service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Bucket of the AWS S3 service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Region of AWS S3 service. It is an optional field that overrides the `Region` key from the secret.
	// Specifying the region here makes a guarantee that it won't be changed anymore.
	Region string `json:"region,omitempty"`

	// Endpoint of the AWS (or other vendor's S3-compatible) service. Leave it empty when using AWS S3 service.
	// You can reference the `REGION` and `BUCKET` variables in the endpoint with `${BUCKET}` and `${REGION}`, e.g.,
	//   s3.${REGION}.amazonaws.com
	//   ${BUCKET}.s3.${REGION}.amazonaws.com
	// +optional
	// +kubebuilder:validation:Pattern="^(?:https://)?(?:[^/.\\s]+\\.)*(?:[^/\\s]+)*$"
	Endpoint string `json:"endpoint,omitempty"`

	// VirtualHostedStyle indicates to use a virtual hosted endpoint when endpoint is specified. The operator automatically
	// adds the bucket prefix for you if this is enabled. Be careful about doubly using the style by specifying an endpoint
	// of virtual hosted style as well as enabling this.
	VirtualHostedStyle bool `json:"virtualHostedStyle,omitempty"`
}

// RisingWaveStateStoreBackendGCS is the collection of parameters for the GCS backend state store.
type RisingWaveStateStoreBackendGCS struct {
	// UseWorkloadIdentity indicates to use workload identity to access the GCS service. If this is enabled, secret is not required, and ADC is used.
	// +kubebuilder:validation:Required
	UseWorkloadIdentity bool `json:"useWorkloadIdentity"`

	// Secret contains the credentials to access the GCS service. It must contain the following keys:
	//   * ServiceAccountCredentials
	// +kubebuilder:validation:Optional
	Secret string `json:"secret,omitempty"`

	// Bucket of the GCS bucket service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Working directory root of the GCS bucket
	// +kubebuilder:validation:Required
	Root string `json:"root"`
}

// RisingWaveStateStoreBackendAliyunOSS is the details of Aliyun OSS storage (S3 compatible) for compute and compactor components.
type RisingWaveStateStoreBackendAliyunOSS struct {
	// Secret contains the credentials to access the Aliyun OSS service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Region of Aliyun OSS service. It is an optional field that overrides the `Region` key from the secret.
	// Specifying the region here makes a guarantee that it won't be changed anymore.
	Region string `json:"region,omitempty"`

	// Bucket of the Aliyun OSS service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// InternalEndpoint indicates if we use the internal endpoint to access Aliyun OSS, which is
	// only available in the internal network.
	InternalEndpoint bool `json:"internalEndpoint,omitempty"`
}

// RisingWaveStateStoreBackendAzureBlob is the details of Azure blob storage (S3 compatible) for compute and compactor components.
type RisingWaveStateStoreBackendAzureBlob struct {
	// Secret contains the credentials to access the Azure Blob service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	// +kubebuilder:validation:Required
	Secret string `json:"secret"`

	// Container Name of the Azure Blob service.
	// +kubebuilder:validation:Required
	Container string `json:"container"`

	// Working directory root of the Azure Blob service.
	// +kubebuilder:validation:Required
	Root string `json:"root"`

	// Endpoint of the Azure Blob service.
	// e.g. https://yufantest.blob.core.windows.net
	// +kubebuilder:validation:Pattern="^(?:https://)?(?:[^/.\\s]+\\.)*(?:[^/\\s]+)*$"
	Endpoint string `json:"endpoint"`
}

// RisingWaveStateStoreBackendHDFS is the details of HDFS storage (S3 compatible) for compute and compactor components.
type RisingWaveStateStoreBackendHDFS struct {
	// Name node of the HDFS
	// +kubebuilder:validation:Required
	NameNode string `json:"nameNode"`

	// Working directory root of the HDFS
	// +kubebuilder:validation:Required
	Root string `json:"root"`
}

// RisingWaveStateStoreBackend is the collection of parameters for the state store that RisingWave uses. Note that one
// and only one of the first-level fields could be set.
type RisingWaveStateStoreBackend struct {
	// DataDirectory is the directory to store the data in the object storage. It is an optional field.
	DataDirectory string `json:"dataDirectory,omitempty"`

	// Memory indicates to store the data in memory. It's only for test usage and strongly discouraged to
	// be used in production.
	// +optional
	Memory *bool `json:"memory,omitempty"`

	// MinIO storage spec.
	// +optional
	MinIO *RisingWaveStateStoreBackendMinIO `json:"minio,omitempty"`

	// S3 storage spec.
	// +optional
	S3 *RisingWaveStateStoreBackendS3 `json:"s3,omitempty"`

	// GCS storage spec.
	// +optional
	GCS *RisingWaveStateStoreBackendGCS `json:"GCS,omitempty"`

	// AliyunOSS storage spec.
	// +optional
	AliyunOSS *RisingWaveStateStoreBackendAliyunOSS `json:"aliyunOSS,omitempty"`

	// Azure Blob storage spec.
	// +optional
	AzureBlob *RisingWaveStateStoreBackendAzureBlob `json:"azureBlob,omitempty"`

	// HDFS storage spec.
	// +optional
	HDFS *RisingWaveStateStoreBackendHDFS `json:"hdfs,omitempty"`

	// WebHDFS storage spec.
	// +optional
	WebHDFS *RisingWaveStateStoreBackendHDFS `json:"webhdfs,omitempty"`
}
