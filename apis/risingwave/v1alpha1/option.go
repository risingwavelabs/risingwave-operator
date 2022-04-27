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

import (
	corev1 "k8s.io/api/core/v1"
)

var defaultOption RisingWaveOptions

type RisingWaveOptions struct {
	Arch Arch

	MetaNode    BaseOptions
	ComputeNode BaseOptions
	MinIO       BaseOptions
	Frontend    BaseOptions
}

// ImageOptions TODO: remove this map after all images support docker multi platform.
type ImageOptions struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type BaseOptions struct {
	Image map[Arch]ImageOptions

	PullPolicy corev1.PullPolicy
	Replicas   int32
	Resources  corev1.ResourceRequirements
}

func SetDefaultOption(opt RisingWaveOptions) {
	defaultOption = opt
}
