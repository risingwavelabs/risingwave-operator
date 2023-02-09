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
	LabelRisingWaveComponent  = "risingwave/component"
	LabelRisingWaveName       = "risingwave/name"
	LabelRisingWaveGeneration = "risingwave/generation"
	LabelRisingWaveGroup      = "risingwave/group"
)

// =================================================
// Annotations.
// =================================================

// System reserved annotations.
const (
	AnnotationRestartAt          = "risingwave/restart-at"
	AnnotationPauseReconcile     = "risingwave.risingwavelabs.com/pause-reconcile"
	AnnotationInheritLabelPrefix = "risingwave.risingwavelabs.com/inherit-label-prefix"
)

// =================================================
// Finalizers.
// =================================================

// System reserved finalizers.
const (
	// FinalizerScaleView Deprecated.
	FinalizerScaleView = "risingwavescaleview/finalizer"
)

// =================================================
// Consts.
// =================================================

// Special label values of LabelRisingWaveGeneration.
const (
	// NoSync indicates that operator won't sync the resource after it's created.
	NoSync = "nosync"
)

// Label values of LabelRisingWaveComponent.
const (
	ComponentMeta      = "meta"
	ComponentFrontend  = "frontend"
	ComponentCompute   = "compute"
	ComponentCompactor = "compactor"
	ComponentConnector = "connector"
	ComponentConfig    = "config"
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

// Port names of components.
const (
	PortService   string = "service"
	PortMetrics   string = "metrics"
	PortDashboard string = "dashboard"
)

// Default port values of components.
const (
	DefaultMetaServicePort      int32 = 5690
	DefaultMetaDashboardPort    int32 = 5691
	DefaultMetaMetricsPort      int32 = 1250
	DefaultComputeServicePort   int32 = 5688
	DefaultComputeMetricsPort   int32 = 1222
	DefaultFrontendServicePort  int32 = 4567
	DefaultFrontendMetricsPort  int32 = 8080
	DefaultCompactorServicePort int32 = 6660
	DefaultCompactorMetricsPort int32 = 1260
	DefaultConnectorServicePort int32 = 5000
	DefaultConnectorMetricsPort int32 = 5001
)
