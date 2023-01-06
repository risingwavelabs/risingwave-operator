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

package factory

import (
	"sort"

	corev1 "k8s.io/api/core/v1"

	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

func nonZeroOrDefault[T comparable](v T, defaultVal T) T {
	var zero T
	if v == zero {
		return defaultVal
	}
	return v
}

func sortSlicesInContainer(container *corev1.Container) {
	utils.TopologicalSort(container.Env)
	sort.Sort(utils.VolumeMountSlice(container.VolumeMounts))
	sort.Sort(utils.VolumeDeviceSlice(container.VolumeDevices))
}

func keepPodSpecConsistent(podSpec *corev1.PodSpec) {
	// Sort slices to make sure there's no random order. Currently, these fields are considered:
	//   - Volumes (sorted by name)
	//   - For each container
	//     - Env (sorted by name)
	//     - VolumeMounts (sorted by name)
	//     - VolumeDevices (sorted by name)

	sort.Sort(utils.VolumeSlice(podSpec.Volumes))

	for _, container := range podSpec.InitContainers {
		sortSlicesInContainer(&container)
	}

	for _, container := range podSpec.Containers {
		sortSlicesInContainer(&container)
	}
}
