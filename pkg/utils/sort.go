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

type EnvVarIdxPair struct {
	EnvVarName string
	Idx        int
}

type EnvVarPriorityQueue []*EnvVarIdxPair

func (pq *EnvVarPriorityQueue) Len() int { return len(*pq) }

func (pq *EnvVarPriorityQueue) Less(i, j int) bool {
	return (*pq)[i].EnvVarName < (*pq)[j].EnvVarName
}

func (pq *EnvVarPriorityQueue) Swap(i, j int) {
	swap((*pq)[i], (*pq)[j])
}

func (pq *EnvVarPriorityQueue) Push(x any) {
	item := x.(*EnvVarIdxPair)
	*pq = append(*pq, item)
}

func (pq *EnvVarPriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

type EnvVarSliceByIdx struct {
	EnvVar []corev1.EnvVar
	Idx    []int
}

func (s EnvVarSliceByIdx) Len() int {
	return len(s.EnvVar)
}

func (s EnvVarSliceByIdx) Less(i, j int) bool {
	return s.Idx[i] < s.Idx[j]
}

func (s EnvVarSliceByIdx) Swap(i, j int) {
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
	sources := make(EnvVarPriorityQueue, 0)
	heap.Init(&sources)

	for key, val := range inDegree {
		if val == 0 {
			heap.Push(&sources, &EnvVarIdxPair{EnvVarName: e[key].Name, Idx: key})
		}
	}

	for len(sources) != 0 {
		vertex := heap.Pop(&sources).(*EnvVarIdxPair)

		sortedOrder = append(sortedOrder, vertex.Idx)
		children := graph[vertex.Idx]

		for _, child := range children {
			inDegree[child] = inDegree[child] - 1
			if inDegree[child] == 0 {
				heap.Push(&sources, &EnvVarIdxPair{EnvVarName: e[child].Name, Idx: child})
			}
		}
	}

	// sort env by the sorted order
	sort.Sort(EnvVarSliceByIdx{
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
