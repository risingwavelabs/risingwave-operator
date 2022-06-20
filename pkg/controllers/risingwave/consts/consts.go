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

package consts

// =================================================
// Labels.
// =================================================

const (
	LabelRisingWaveComponent  = "risingwave/component"
	LabelRisingWaveName       = "risingwave/name"
	LabelRisingWaveGeneration = "risingwave/generation"
)

// =================================================
// Annotations.
// =================================================

const ()

// =================================================
// Consts.
// =================================================

// Label values of LabelRisingWaveComponent.
const (
	ComponentMeta      = "meta"
	ComponentFrontend  = "frontend"
	ComponentCompute   = "compute"
	ComponentCompactor = "compactor"
)

// Credential keys for AWS S3.
const (
	AWSS3AccessKeyID     string = "AccessKeyID"
	AWSS3SecretAccessKey string = "SecretAccessKey"
	AWSS3Region          string = "Region"
)
