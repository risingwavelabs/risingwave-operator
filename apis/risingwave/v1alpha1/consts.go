/*
 * Copyright 2022 Singularity Data
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

package v1alpha1

const (
	MinIOServerPortName = "minio-server"
	MinIOServerPort     = 9301

	MinIOConsolePortName = "minio-console"
	MinIOConsolePort     = 9400

	MetaServerPortName = "meta-server"
	MetaServerPort     = 5690

	MetaDashboardPortName = "meta-dashboard"
	MetaDashboardPort     = 5691

	MetaMetricsPortName = "metrics"
	MetaMetricsPort     = 1250

	ComputeNodePortName = "compute-node"
	ComputeNodePort     = 5688

	ComputeNodeMetricsPortName = "metrics"
	ComputeNodeMetricsPort     = 1222

	FrontendPortName = "frontend"
	FrontendPort     = 4567

	CompactorNodePortName = "compactor-node"
	CompactorNodePort     = 6660

	CompactorNodeMetricsPortName = "metrics"
	CompactorNodeMetricsPort     = 1260
)

const (
	ArchKey = "kubernetes.io/arch"
)

const (
	CloudProviderConfigureSecretName string = "cloud-provider-configure"
)
