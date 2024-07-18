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
	RustLog       = "RUST_LOG"
	RustMinStack  = "RUST_MIN_STACK"
	JavaOpts      = "JAVA_OPTS"

	RWListenAddr             = "RW_LISTEN_ADDR"
	RWAdvertiseAddr          = "RW_ADVERTISE_ADDR"
	RWDashboardHost          = "RW_DASHBOARD_HOST"
	RWPrometheusHost         = "RW_PROMETHEUS_HOST"
	RWEtcdEndpoints          = "RW_ETCD_ENDPOINTS"
	RWSQLEndpoint            = "RW_SQL_ENDPOINT"
	RWSQLDatabase            = "RW_SQL_DATABASE"
	RWEtcdAuth               = "RW_ETCD_AUTH"
	RWEtcdUsername           = "RW_ETCD_USERNAME"
	RWEtcdPassword           = "RW_ETCD_PASSWORD"
	RWSQLUsername            = "RW_SQL_USERNAME"
	RWSQLPassword            = "RW_SQL_PASSWORD"
	RWMySQLUsername          = "RW_MYSQL_USERNAME"
	RWMySQLPassword          = "RW_MYSQL_PASSWORD"
	RWPostgresUsername       = "RW_POSTGRES_USERNAME"
	RWPostgresPassword       = "RW_POSTGRES_PASSWORD"
	RWConfigPath             = "RW_CONFIG_PATH"
	RWStateStore             = "RW_STATE_STORE"
	RWDataDirectory          = "RW_DATA_DIRECTORY"
	RWWorkerThreads          = "RW_WORKER_THREADS"
	RWBackend                = "RW_BACKEND"
	RWMetaAddr               = "RW_META_ADDR"
	RWMetaAddrLegacy         = "RW_META_ADDRESS" // Will deprecate soon.
	RWPrometheusListenerAddr = "RW_PROMETHEUS_LISTENER_ADDR"
	RWParallelism            = "RW_PARALLELISM"
	RWTotalMemoryBytes       = "RW_TOTAL_MEMORY_BYTES"
	RWSslCert                = "RW_SSL_CERT"
	RWSslKey                 = "RW_SSL_KEY"
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
	S3CompatibleRegion          = "AWS_REGION"
	S3CompatibleBucket          = "AWS_S3_BUCKET"
	S3CompatibleAccessKeyID     = "AWS_ACCESS_KEY_ID"
	S3CompatibleSecretAccessKey = "AWS_SECRET_ACCESS_KEY"
	S3CompatibleEndpoint        = "RW_S3_ENDPOINT"
)

// Azure blob.
const (
	AzureBlobEndpoint    = "AZBLOB_ENDPOINT"
	AzureBlobAccountName = "AZBLOB_ACCOUNT_NAME"
	AzureBlobAccountKey  = "AZBLOB_ACCOUNT_KEY"
)

// AliyunOSS.
const (
	AliyunOSSRegion          = "OSS_REGION"
	AliyunOSSBucket          = "OSS_S3_BUCKET"
	AliyunOSSEndpoint        = "OSS_ENDPOINT"
	AliyunOSSAccessKeyID     = "OSS_ACCESS_KEY_ID"
	AliyunOSSSecretAccessKey = "OSS_ACCESS_KEY_SECRET"
)

// HuaweiCloudOBS.
const (
	HuaweiCloudOBSRegion          = "OBS_REGION"
	HuaweiCloudOBSEndpoint        = "OBS_ENDPOINT"
	HuaweiCloudOBSAccessKeyID     = "OBS_ACCESS_KEY_ID"
	HuaweiCloudOBSSecretAccessKey = "OBS_SECRET_ACCESS_KEY"
)

const (
	// GoogleApplicationCredentials for GCS service.
	GoogleApplicationCredentials = "GOOGLE_APPLICATION_CREDENTIALS"
)
