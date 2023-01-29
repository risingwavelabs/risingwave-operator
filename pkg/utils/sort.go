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

package utils

import (
	"container/heap"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

func swap[T any](a, b *T) {
	t := *a
	*a = *b
	*b = t
}

type envVarIdxPair struct {
	EnvVarName string
	Idx        int
}

type envVarPriorityQueue []*envVarIdxPair

// Len implements sort.Interface.
func (pq *envVarPriorityQueue) Len() int { return len(*pq) }

// Less implements sort.Interface.
func (pq *envVarPriorityQueue) Less(i, j int) bool {
	return (*pq)[i].EnvVarName < (*pq)[j].EnvVarName
}

// Swap implements sort.Interface.
func (pq *envVarPriorityQueue) Swap(i, j int) {
	swap((*pq)[i], (*pq)[j])
}

// Push implements heap.Interface.
func (pq *envVarPriorityQueue) Push(x any) {
	item := x.(*envVarIdxPair)
	*pq = append(*pq, item)
}

// Pop implements heap.Interface.
func (pq *envVarPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

type envVarSliceByIdx struct {
	EnvVar []corev1.EnvVar
	Idx    []int
}

// Len implements sort.Interface.
func (s envVarSliceByIdx) Len() int {
	return len(s.EnvVar)
}

// Less implements sort.Interface.
func (s envVarSliceByIdx) Less(i, j int) bool {
	return s.Idx[i] < s.Idx[j]
}

// Swap implements sort.Interface.
func (s envVarSliceByIdx) Swap(i, j int) {
	swap(&s.EnvVar[i], &s.EnvVar[j])
}

// DependsOn tells if a depends on b.
func DependsOn(a, b corev1.EnvVar) bool {
	idx := strings.Index(a.Value, "$("+b.Name+")")
	if idx == -1 || (idx > 0 && a.Value[idx-1] == '$') {
		return false
	}
	return true
}

// TopologicalSort runs a topological sort on given env vars.
func TopologicalSort(e []corev1.EnvVar) {
	inDegree := make(map[int]int)
	graph := make(map[int][]int)
	n := len(e)

	// initialize the dependency graph
	for i := 0; i < n; i++ {
		inDegree[i] = 0
		graph[i] = make([]int, 0)
	}

	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			// If a depends on b or b depends on a, then use their input order
			if DependsOn(e[i], e[j]) || DependsOn(e[j], e[i]) {
				var child, parent int
				if i < j {
					child, parent = j, i
				} else {
					child, parent = i, j
				}

				graph[parent] = append(graph[parent], child)
				inDegree[child] = inDegree[child] + 1
			}
		}
	}

	sortedOrder := make([]int, 0)
	sources := make(envVarPriorityQueue, 0)
	heap.Init(&sources)

	for key, val := range inDegree {
		if val == 0 {
			heap.Push(&sources, &envVarIdxPair{EnvVarName: e[key].Name, Idx: key})
		}
	}

	for len(sources) != 0 {
		vertex := heap.Pop(&sources).(*envVarIdxPair)

		sortedOrder = append(sortedOrder, vertex.Idx)
		children := graph[vertex.Idx]

		for _, child := range children {
			inDegree[child] = inDegree[child] - 1
			if inDegree[child] == 0 {
				heap.Push(&sources, &envVarIdxPair{EnvVarName: e[child].Name, Idx: child})
			}
		}
	}

	// sort env by the sorted order
	sort.Sort(envVarSliceByIdx{
		e,
		sortedOrder,
	})
}

// VolumeMountSlice is a wrapper of []corev1.VolumeMount that implements the sort.Interface.
type VolumeMountSlice []corev1.VolumeMount

// Len implements sort.Interface.
func (s VolumeMountSlice) Len() int {
	return len(s)
}

// Less implements sort.Interface.
func (s VolumeMountSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap implements sort.Interface.
func (s VolumeMountSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

// VolumeSlice is a wrapper of []corev1.Volume that implements the sort.Interface.
type VolumeSlice []corev1.Volume

// Len implements sort.Interface.
func (s VolumeSlice) Len() int {
	return len(s)
}

// Less implements sort.Interface.
func (s VolumeSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap implements sort.Interface.
func (s VolumeSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}

// VolumeDeviceSlice is a wrapper of []corev1.VolumeDevice that implements the sort.Interface.
type VolumeDeviceSlice []corev1.VolumeDevice

// Len implements sort.Interface.
func (s VolumeDeviceSlice) Len() int {
	return len(s)
}

// Less implements sort.Interface.
func (s VolumeDeviceSlice) Less(i, j int) bool {
	return s[i].Name < s[j].Name
}

// Swap implements sort.Interface.
func (s VolumeDeviceSlice) Swap(i, j int) {
	swap(&s[i], &s[j])
}
