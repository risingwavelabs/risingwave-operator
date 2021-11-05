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

package manger

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
	ComputeNodeTomlName  = "compute-config"
	ComputeNodeTomlKey   = "risingwave.toml"
	ComputeNodeTomlPath  = "risingwave.toml"
	ComputeNodeTomlValue = `[ server ]
heartbeat_interval = 1000

[ batch ]
chunk_size = 1024

[ streaming ]
chunk_size = 1024

[ storage ]
sstable_size = 268435456
block_size = 4096
bloom_false_positive = 0.1
data_directory = "hummock_001"
checksum_algo = "crc32c"
async_checkpoint_enabled = true
`
)

const (
	MetaNodeName    string = "MetaNode"
	FrontendName    string = "Frontend"
	ComputeNodeName string = "ComputeNode"
	MinIOName       string = "MinIO"
)

const (
	FrontendTemplateStr = `
    risingwave.pgserver.ip=$(MY_POD_IP)
    risingwave.pgserver.port=%d
    risingwave.leader.clustermode=Distributed

    ## optional metadata service config
    risingwave.catalog.mode=Remote
    risingwave.meta.node=%s:%d`
)
