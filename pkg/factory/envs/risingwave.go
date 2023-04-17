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

package envs

// cspell:disable

// Environment variables to pass options to the RisingWave process and control the behavior.
const (
	RustBacktrace = "RUST_BACKTRACE"
	JavaOpts      = "JAVA_OPTS"

	RWListenAddr             = "RW_LISTEN_ADDR"
	RWAdvertiseAddr          = "RW_ADVERTISE_ADDR"
	RWDashboardHost          = "RW_DASHBOARD_HOST"
	RWPrometheusHost         = "RW_PROMETHEUS_HOST"
	RWEtcdEndpoints          = "RW_ETCD_ENDPOINTS"
	RWEtcdAuth               = "RW_ETCD_AUTH"
	RWEtcdUsername           = "RW_ETCD_USERNAME"
	RWEtcdPassword           = "RW_ETCD_PASSWORD"
	RWConfigPath             = "RW_CONFIG_PATH"
	RWStateStore             = "RW_STATE_STORE"
	RWDataDirectory          = "RW_DATA_DIRECTORY"
	RWWorkerThreads          = "RW_WORKER_THREADS"
	RWConnectorRPCEndPoint   = "RW_CONNECTOR_RPC_ENDPOINT"
	RWBackend                = "RW_BACKEND"
	RWMetaAddr               = "RW_META_ADDR"
	RWMetaAddrLegacy         = "RW_META_ADDRESS" // Will deprecate soon.
	RWMetricsLevel           = "RW_METRICS_LEVEL"
	RWPrometheusListenerAddr = "RW_PROMETHEUS_LISTENER_ADDR"
	RWParallelism            = "RW_PARALLELISM"
	RWTotalMemoryBytes       = "RW_TOTAL_MEMORY_BYTES"
)

// MinIO.
const (
	MinIOEndpoint = "MINIO_ENDPOINT"
	MinIOBucket   = "MINIO_BUCKET"
	MinIOUsername = "MINIO_USERNAME"
	MinIOPassword = "MINIO_PASSWORD"
)

// etcd.
const (
	EtcdUsernameLegacy = "ETCD_USERNAME"
	EtcdPasswordLegacy = "ETCD_PASSWORD"
)

// AWS S3.
const (
	AWSRegion              = "AWS_REGION"
	AWSAccessKeyID         = "AWS_ACCESS_KEY_ID"
	AWSSecretAccessKey     = "AWS_SECRET_ACCESS_KEY"
	AWSS3Bucket            = "AWS_S3_BUCKET"
	AWSEC2MetadataDisabled = "AWS_EC2_METADATA_DISABLED"
)

// S3 compatible.
const (
	S3CompatibleRegion          = "S3_COMPATIBLE_REGION"
	S3CompatibleBucket          = "S3_COMPATIBLE_BUCKET"
	S3CompatibleAccessKeyID     = "S3_COMPATIBLE_ACCESS_KEY_ID"
	S3CompatibleSecretAccessKey = "S3_COMPATIBLE_SECRET_ACCESS_KEY"
	S3CompatibleEndpoint        = "S3_COMPATIBLE_ENDPOINT"
)

const (
	AzureBlobEndpoint    = "AZBLOB_ENDPOINT"
	AzureBlobAccountName = "AZBLOB_ACCOUNT_NAME"
	AzureBlobAccountKey  = "AZBLOB_ACCOUNT_KEY"
)

const (
	// GoogleApplicationCredentials for GCS service.
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)
