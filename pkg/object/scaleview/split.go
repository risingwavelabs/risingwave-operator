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

package scaleview

import (
	"math"
	"sort"

	"github.com/samber/lo"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

func split(total, n int) int {
	if total%n == 0 {
		return total / n
	} else {
		return total/n + 1
	}
}

func canonizeScalePolicy(p risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy) risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy {
	r := p
	if r.MaxReplicas == nil {
		r.MaxReplicas = pointer.Int32(math.MaxInt32)
	}
	return r
}

// SplitReplicas tries to split the total replicas of .spec.replicas into several groups defined in the .spec.scalePolicy.
// It must be a stable function.
func SplitReplicas(sv *risingwavev1alpha1.RisingWaveScaleView) map[string]int32 {
	// Group groups by priority.
	groupsByPriority := make(map[int32][]risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy)
	for _, p := range sv.Spec.ScalePolicy {
		if _, ok := groupsByPriority[p.Priority]; !ok {
			groupsByPriority[p.Priority] = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
				canonizeScalePolicy(p),
			}
		} else {
			groupsByPriority[p.Priority] = append(groupsByPriority[p.Priority], canonizeScalePolicy(p))
		}
	}

	// Sort groups of each priority by (maxReplicas, group).
	for k := range groupsByPriority {
		groups := groupsByPriority[k]
		sort.Slice(groups, func(i, j int) bool {
			if *groups[i].MaxReplicas != *groups[j].MaxReplicas {
				return *groups[i].MaxReplicas < *groups[j].MaxReplicas
			}
			return groups[i].Group < groups[j].Group
		})
	}

	// Get sorted priorities, descending.
	priorities := lo.Keys(groupsByPriority)
	sort.Slice(priorities, func(i, j int) bool {
		return priorities[i] > priorities[j]
	})

	totalLeft := int(sv.Spec.Replicas)
	replicas := make(map[string]int32)

	for _, priority := range priorities {
		groups := groupsByPriority[priority]

		// Set the replicas of groups to the default zero when there are no left replicas.
		if totalLeft <= 0 {
			for i := 0; i < len(groups); i++ {
				g := groups[i]
				replicas[g.Group] = int32(0)
			}
			continue
		}

		// Since the groups are sorted by maxReplicas, then
		//   - If the maxReplicas * leftGroupSize <= totalLeft, it means each group can at least get maxReplicas replicas,
		//     just take that much.
		//   - Otherwise, it means each group can at most get (totalLeft / leftGroupSize)  + 1 replicas, we use a split function
		//     to help take the replicas.
		for i := 0; i < len(groups); i++ {
			g := groups[i]
			max := int(*g.MaxReplicas)

			if max*(len(groups)-i) <= totalLeft {
				replicas[g.Group] = int32(max)
				totalLeft -= max
			} else {
				taken := split(totalLeft, len(groups)-i)
				replicas[g.Group] = int32(taken)
				totalLeft -= taken
			}
		}
	}

	// Run a check here to ensure it's working as expected.
	sum := int32(0)
	for _, r := range replicas {
		sum += r
	}
	if sum != sv.Spec.Replicas {
		panic("algorithm has bug")
	}

	return replicas
}
