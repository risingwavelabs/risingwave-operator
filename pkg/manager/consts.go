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

package manager

import "time"

const (
	RisingWaveKey           string = "risingwave-app"
	RisingWaveName          string = "risingwave-name"
	RisingWaveMetaValue     string = "meta-node"
	RisingWaveComputeValue  string = "compute-node"
	RisingWaveFrontendValue string = "frontend"
	RisingWaveMinIOValue    string = "minio"
)

const (
	RetryPeriod  = 1 * time.Second
	RetryTimeout = 20 * time.Second
)

const (
	FrontendContainerName = "frontend-container"
)

const (
	ComputeNodeTomlName = "compute-config"
	ComputeNodeTomlKey  = "risingwave.toml"
	ComputeNodeTomlPath = "risingwave.toml"
)

const (
	MetaNodeName    string = "MetaNode"
	FrontendName    string = "Frontend"
	ComputeNodeName string = "ComputeNode"
	MinIOName       string = "MinIO"
	S3Name          string = "S3"
)

const (
	TemplateFileDir           = "/template"
	ComputeNodeConfigTemplate = "compute-config.yaml"

	AccessKeyID     string = "AccessKeyID"
	SecretAccessKey string = "SecretAccessKey"
	Region          string = "Region"
)
