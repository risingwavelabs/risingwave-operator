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

// RisingWaveMinIOCredentials is the reference and keys selector to the MinIO access credentials stored in a local secret.
type RisingWaveMinIOCredentials struct {
	// The name of the secret in the pod's namespace to select from.
	SecretName string `json:"secretName"`

	// UsernameKeyRef is the key of the secret to be the username. Must be a valid secret key.
	// Defaults to "username".
	// +kubebuilder:default=username
	UsernameKeyRef string `json:"usernameKeyRef,omitempty"`

	// PasswordKeyRef is the key of the secret to be the password. Must be a valid secret key.
	// Defaults to "password".
	// +kubebuilder:default=password
	PasswordKeyRef string `json:"passwordKeyRef,omitempty"`
}

// RisingWaveStateStoreBackendMinIO is the collection of parameters for the MinIO backend state store.
type RisingWaveStateStoreBackendMinIO struct {
	// RisingWaveMinIOCredentials is the credentials provider from a Secret.
	RisingWaveMinIOCredentials `json:"credentials"`

	// Secret contains the credentials to access the MinIO service. It must contain the following keys:
	//   * username
	//   * password
	// Deprecated: Please use "credentials" field instead. The "Secret" field will be removed in a future release.
	Secret string `json:"secret,omitempty"`

	// Endpoint of the MinIO service.
	// +kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`

	// Bucket of the MinIO service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`
}

// RisingWaveS3Credentials is the reference and keys selector to the AWS access credentials stored in a local secret.
type RisingWaveS3Credentials struct {
	// UseServiceAccount indicates whether to use the service account token mounted in the pod. It only works when using
	// the AWS S3. If this is enabled, secret and keys are ignored. Defaults to false.
	// +optional
	UseServiceAccount *bool `json:"useProjectedServiceAccountToken,omitempty"`

	// The name of the secret in the pod's namespace to select from.
	SecretName string `json:"secretName,omitempty"`

	// AccessKeyRef is the key of the secret to be the access key. Must be a valid secret key.
	// Defaults to "AccessKeyID".
	// +kubebuilder:default=AccessKeyID
	AccessKeyRef string `json:"accessKeyRef,omitempty"`

	// SecretAccessKeyRef is the key of the secret to be the secret access key. Must be a valid secret key.
	// Defaults to "SecretAccessKey".
	// +kubebuilder:default=SecretAccessKey
	SecretAccessKeyRef string `json:"secretAccessKeyRef,omitempty"`
}

// RisingWaveStateStoreBackendS3 is the collection of parameters for the S3 backend state store.
type RisingWaveStateStoreBackendS3 struct {
	// RisingWaveS3Credentials is the credentials provider from a Secret.
	RisingWaveS3Credentials `json:"credentials"`

	// Secret contains the credentials to access the AWS S3 service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// Deprecated: Please use "credentials" field instead. The "Secret" field will be removed in a future release.
	Secret string `json:"secret,omitempty"`

	// Bucket of the AWS S3 service.
	// +kubebuilder:validation:Required
	Bucket string `json:"bucket"`

	// Region of AWS S3 service. Defaults to "us-east-1".
	// +kubebuilder:validation:Required
	// +kubebuilder:default=us-east-1
	Region string `json:"region"`

	// Endpoint of the AWS (or other vendor's S3-compatible) service. Leave it empty when using AWS S3 service.
	// You can reference the `REGION` and `BUCKET` variables in the endpoint with `${BUCKET}` and `${REGION}`, e.g.,
	//   s3.${REGION}.amazonaws.com
	//   ${BUCKET}.s3.${REGION}.amazonaws.com
	// +optional
	// +kubebuilder:validation:Pattern="^(?:https://)?(?:[^/.\\s]+\\.)*(?:[^/\\s]+)*$"
	Endpoint string `json:"endpoint,omitempty"`
}

// RisingWaveGCSCredentials is the reference and keys selector to the GCS access credentials stored in a local secret.
type RisingWaveGCSCredentials struct {
	// UseWorkloadIdentity indicates to use workload identity to access the GCS service.
	// If this is enabled, secret is not required, and ADC is used.
	// +kubebuilder:validation:Required
	UseWorkloadIdentity *bool `json:"useWorkloadIdentity,omitempty"`

	// The name of the secret in the pod's namespace to select from.
	// +optional
	SecretName string `json:"secretName,omitempty"`

	// ServiceAccountCredentialsKeyRef is the key of the secret to be the service account credentials. Must be a valid secret key.
	// Defaults to "ServiceAccountCredentials".
	// +kubebuilder:default=ServiceAccountCredentials
	// +optional
	ServiceAccountCredentialsKeyRef string `json:"serviceAccountCredentialsKeyRef,omitempty"`
}

// RisingWaveStateStoreBackendGCS is the collection of parameters for the GCS backend state store.
type RisingWaveStateStoreBackendGCS struct {
	// RisingWaveGCSCredentials is the credentials provider from a Secret.
	RisingWaveGCSCredentials `json:"credentials,omitempty"`

	// Secret contains the credentials to access the GCS service. It must contain the following keys:
	//   * ServiceAccountCredentials
	// +kubebuilder:validation:Optional
	// Deprecated: Please use "credentials" field instead. The "Secret" field will be removed in a future release.
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
	// RisingWaveS3Credentials is the credentials provider from a Secret.
	RisingWaveS3Credentials `json:"credentials"`

	// Secret contains the credentials to access the Aliyun OSS service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	//   * Region (optional if region is specified in the field.)
	// Deprecated: Please use "credentials" field instead. The "Secret" field will be removed in a future release.
	Secret string `json:"secret,omitempty"`

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

// RisingWaveAzureBlobCredentials is the reference and keys selector to the AzureBlob access credentials stored in a local secret.
type RisingWaveAzureBlobCredentials struct {
	// The name of the secret in the pod's namespace to select from.
	SecretName string `json:"secretName"`

	// AccountNameKeyRef is the key of the secret to be the account name. Must be a valid secret key.
	// Defaults to "AccountName".
	// +kubebuilder:default=AccountName
	AccountNameRef string `json:"accountNameRef,omitempty"`

	// AccountKeyRef is the key of the secret to be the secret account key. Must be a valid secret key.
	// Defaults to "AccountKey".
	// +kubebuilder:default=AccountKey
	AccountKeyRef string `json:"AccountKeyRef,omitempty"`
}

// RisingWaveStateStoreBackendAzureBlob is the details of Azure blob storage (S3 compatible) for compute and compactor components.
type RisingWaveStateStoreBackendAzureBlob struct {
	// RisingWaveAzureBlobCredentials is the credentials provider from a Secret.
	RisingWaveAzureBlobCredentials `json:"credentials"`

	// Secret contains the credentials to access the Azure Blob service. It must contain the following keys:
	//   * AccessKeyID
	//   * SecretAccessKey
	// Deprecated: Please use "credentials" field instead. The "Secret" field will be removed in a future release.
	Secret string `json:"secret,omitempty"`

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
	GCS *RisingWaveStateStoreBackendGCS `json:"gcs,omitempty"`

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
