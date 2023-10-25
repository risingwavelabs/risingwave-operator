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

package consts

// =================================================
// Labels.
// =================================================

// System reserved labels.
const (
	LabelRisingWaveComponent       = "risingwave/component"
	LabelRisingWaveName            = "risingwave/name"
	LabelRisingWaveGeneration      = "risingwave/generation"
	LabelRisingWaveGroup           = "risingwave/group"
	LabelRisingWaveMetaRole        = "risingwave/meta-role"
	LabelRisingWaveOperatorVersion = "risingwave/operator-version"
)

// =================================================
// Annotations.
// =================================================

// System reserved annotations.
const (
	AnnotationRestartAt               = "risingwave/restart-at"
	AnnotationPauseReconcile          = "risingwave.risingwavelabs.com/pause-reconcile"
	AnnotationBypassValidatingWebhook = "risingwave.risingwavelabs.com/bypass-validating-webhook"
	AnnotationInheritLabelPrefix      = "risingwave.risingwavelabs.com/inherit-label-prefix"
)

// =================================================
// Consts.
// =================================================

// Label values of LabelRisingWaveMetaRole.
const (
	MetaRoleLeader   = "leader"
	MetaRoleFollower = "follower"
	MetaRoleUnknown  = "unknown"
)

// Special label values of LabelRisingWaveGeneration.
const (
	// NoSync indicates that operator won't sync the resource after it's created.
	NoSync = "nosync"
)

// Label values of LabelRisingWaveComponent.
const (
	ComponentMeta       = "meta"
	ComponentFrontend   = "frontend"
	ComponentCompute    = "compute"
	ComponentCompactor  = "compactor"
	ComponentConnector  = "connector"
	ComponentStandalone = "standalone"
	ComponentConfig     = "config"
)

// Credential keys for MinIO.
const (
	SecretKeyMinIOUsername string = "username"
	SecretKeyMinIOPassword string = "password"
)

// Credential keys for etcd.
const (
	SecretKeyEtcdUsername string = "username"
	SecretKeyEtcdPassword string = "password"
)

// Credential keys for AWS S3.
const (
	SecretKeyAWSS3AccessKeyID     string = "AccessKeyID"
	SecretKeyAWSS3SecretAccessKey string = "SecretAccessKey"
	SecretKeyAWSS3Region          string = "Region"
)

// Credential keys for Azure Blob.
const (
	SecretKeyAzureBlobAccountName string = "AccountName"
	SecretKeyAzureBlobAccountKey  string = "AccountKey"
)

// Credential keys for AliyunOSS.
const (
	SecretKeyAliyunOSSAccessKeyID     string = "AccessKeyID"
	SecretKeyAliyunOSSAccessKeySecret string = "AccessKeySecret"
)

// Credentials for GCS.
const (
	SecretKeyGCSServiceAccountCredentials string = "ServiceAccountCredentials"
)

// Port names of components.
const (
	PortService   string = "service"
	PortMetrics   string = "metrics"
	PortDashboard string = "dashboard"
)

// Port numbers of components.
const (
	MetaServicePort      int32 = 5690
	MetaDashboardPort    int32 = 5691
	MetaMetricsPort      int32 = 1250
	ComputeServicePort   int32 = 5688
	ComputeMetricsPort   int32 = 1222
	FrontendServicePort  int32 = 4567
	FrontendMetricsPort  int32 = 8080
	CompactorServicePort int32 = 6660
	CompactorMetricsPort int32 = 1260
	ConnectorServicePort int32 = 50051
	ConnectorMetricsPort int32 = 50052
)
