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
	"sort"
	"strings"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
)

func swap[T any](a, b *T) {
	t := *a
	*a = *b
	*b = t
}

type EnvVarIdxSlice struct {
	EnvVarName []string
	Idx        []int
}

func (s EnvVarIdxSlice) Len() int {
	return len(s.EnvVarName)
}

func (s EnvVarIdxSlice) Less(i, j int) bool {
	return s.EnvVarName[i] < s.EnvVarName[j]
}

func (s EnvVarIdxSlice) Swap(i, j int) {
	swap(&s.Idx[i], &s.Idx[j])
}

type EnvVarSlice struct {
	EnvVar []corev1.EnvVar
	Idx    []int
}

func (s EnvVarSlice) Len() int {
	return len(s.EnvVar)
}

func (s EnvVarSlice) Less(i, j int) bool {
	return s.Idx[i] < s.Idx[j]
}

func (s EnvVarSlice) Swap(i, j int) {
	swap(&s.EnvVar[i], &s.EnvVar[j])
}

func DependsOn(a, b corev1.EnvVar) bool {
	idx := strings.Index(a.Value, "$("+b.Name+")")
	if idx == -1 || (idx > 0 && a.Value[idx-1] == '$') {
		return false
	}
	return true
}

func TopologicalSort(e []corev1.EnvVar) {
	inDegree := make(map[int]int)
	graph := make(map[int][]int)
	for i := 0; i < len(e); i++ {
		inDegree[i] = 0
		graph[i] = make([]int, 0)
	}

	edges := make([][]int, 0)
	for i := 0; i < len(e); i++ {
		for j := i; j < len(e); j++ {
			// If a depends on b or b depends on a, then use their input order
			if DependsOn(e[i], e[j]) || DependsOn(e[j], e[i]) {
				if i < j {
					edges = append(edges, []int{i, j})
				} else {
					edges = append(edges, []int{j, i})
				}

				child, parent := edges[len(edges)-1][1], edges[len(edges)-1][0]
				graph[parent] = append(graph[parent], child)
				inDegree[child] = inDegree[child] + 1
			}
		}
	}

	sortedOrder := make([]int, 0)
	sources := make([]int, 0)
	for key, val := range inDegree {
		if val == 0 {
			sources = append(sources, key)
		}
	}

	for len(sources) != 0 {
		n := len(sources)
		sort.Sort(EnvVarIdxSlice{
			lo.Map(sources, func(i int, _ int) string {
				return e[i].Name
			}),
			sources,
		})

		for i := 0; i < n; i++ {
			vertex := sources[0]
			sources = sources[1:]

			sortedOrder = append(sortedOrder, vertex)
			children := graph[vertex]

			for _, child := range children {
				inDegree[child] = inDegree[child] - 1
				if inDegree[child] == 0 {
					sources = append(sources, child)
				}
			}
		}
	}

	sort.Sort(EnvVarSlice{
		e,
		sortedOrder,
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
