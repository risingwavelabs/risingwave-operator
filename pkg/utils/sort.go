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

package utils

import corev1 "k8s.io/api/core/v1"

func swap[T any](a, b *T) {
	t := *a
	*a = *b
	*b = t
}

type EnvVarSlice []corev1.EnvVar

func (s EnvVarSlice) Len() int {
	return len(s)
}

func (s EnvVarSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s EnvVarSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

type VolumeMountSlice []corev1.VolumeMount

func (s VolumeMountSlice) Len() int {
	return len(s)
}

func (s VolumeMountSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s VolumeMountSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

type VolumeSlice []corev1.Volume

func (s VolumeSlice) Len() int {
	return len(s)
}

func (s VolumeSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s VolumeSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

type VolumeDeviceSlice []corev1.VolumeDevice

func (s VolumeDeviceSlice) Len() int {
	return len(s)
}

func (s VolumeDeviceSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

func (s VolumeDeviceSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}
