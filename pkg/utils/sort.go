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

import (
	"strings"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
)

func swap[T any](a, b *T) {
	t := *a
	*a = *b
	*b = t
}

type EnvVarIdxPair struct {
	EnvVar corev1.EnvVar
	Idx    int
}

type EnvVarSlice []EnvVarIdxPair

func (s EnvVarSlice) Len() int {
	return len(s)
}

func (s EnvVarSlice) Less(i, j int) bool {
	// If a depends on b or b depends on a, then use their input order
	if s[i].DependsOn(s[j]) || s[j].DependsOn(s[i]) {
		return s[i].Idx < s[j].Idx
	}

	// Otherwise, compare the name of a and b in alphabetical order
	return s[i].EnvVar.Name < s[j].EnvVar.Name
}

func (s EnvVarSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

func (a EnvVarIdxPair) DependsOn(b EnvVarIdxPair) bool {
	idx := strings.Index(a.EnvVar.Value, "$("+b.EnvVar.Name+")")
	if idx == -1 || (idx > 0 && a.EnvVar.Value[idx-1] == '$') {
		return false
	}
	return true
}

func ToEnvVarSlice(e []corev1.EnvVar) EnvVarSlice {
	return lo.Map(e, func(env corev1.EnvVar, idx int) EnvVarIdxPair {
		return EnvVarIdxPair{
			env,
			idx,
		}
	})
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
