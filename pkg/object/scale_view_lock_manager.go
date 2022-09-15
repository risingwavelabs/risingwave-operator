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

package object

import (
	"errors"
	"math"
	"sort"

	"github.com/samber/lo"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

type ScaleViewLockManager struct {
	risingwave *risingwavev1alpha1.RisingWave
}

func (svl *ScaleViewLockManager) getScaleViewLockIndex(sv *risingwavev1alpha1.RisingWaveScaleView) (int, *risingwavev1alpha1.RisingWaveScaleViewLock) {
	scaleViews := svl.risingwave.Status.ScaleViews
	for i, s := range scaleViews {
		if s.Name == sv.Name && s.UID == sv.UID {
			return i, &scaleViews[i]
		}
	}
	return 0, nil
}

func (svl *ScaleViewLockManager) GetScaleViewLock(sv *risingwavev1alpha1.RisingWaveScaleView) *risingwavev1alpha1.RisingWaveScaleViewLock {
	_, r := svl.getScaleViewLockIndex(sv)
	return r
}

func (svl *ScaleViewLockManager) IsScaleViewLocked(sv *risingwavev1alpha1.RisingWaveScaleView) bool {
	return svl.GetScaleViewLock(sv) != nil
}

func canonizeScalePolicy(p risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy) risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy {
	r := p
	if r.MaxReplicas == nil {
		r.MaxReplicas = pointer.Int32(math.MaxInt32)
	}
	return r
}

func (svl *ScaleViewLockManager) split(total, n int) int {
	if total%n == 0 {
		return total / n
	} else {
		return total/n + 1
	}
}

func (svl *ScaleViewLockManager) splitReplicasIntoGroups(sv *risingwavev1alpha1.RisingWaveScaleView) map[string]int32 {
	// Must be a stable algorithm.

	groupsGroupByPriority := make(map[int32][]risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy)
	for _, p := range sv.Spec.ScalePolicy {
		if _, ok := groupsGroupByPriority[p.Priority]; !ok {
			groupsGroupByPriority[p.Priority] = []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
				canonizeScalePolicy(p),
			}
		} else {
			groupsGroupByPriority[p.Priority] = append(groupsGroupByPriority[p.Priority], canonizeScalePolicy(p))
		}
	}

	for k := range groupsGroupByPriority {
		groups := groupsGroupByPriority[k]
		sort.Slice(groups, func(i, j int) bool {
			if *groups[i].MaxReplicas != *groups[j].MaxReplicas {
				return *groups[i].MaxReplicas < *groups[j].MaxReplicas
			}
			return groups[i].Group < groups[j].Group
		})
	}

	totalLeft := int(sv.Spec.Replicas)
	replicas := make(map[string]int32)

	for priority := int32(10); priority >= 0; priority-- {
		groups, ok := groupsGroupByPriority[priority]
		if !ok {
			continue
		}
		if totalLeft <= 0 {
			for i := 0; i < len(groups); i++ {
				g := groups[i]
				replicas[g.Group] = int32(0)
			}
		} else {
			base := 0
			for i := 0; i < len(groups); i++ {
				g := groups[i]
				max := int(*g.MaxReplicas)
				if max == math.MaxInt32 {
					taken := svl.split(totalLeft, len(groups)-i)
					replicas[g.Group] = int32(taken + base)
					totalLeft -= taken
				} else {
					if totalLeft >= (max-base)*(len(groups)-i) {
						replicas[g.Group] = int32(max)
						totalLeft -= (max - base) * (len(groups) - i)
						base = max
					} else {
						for j := i; j < len(groups); j++ {
							g := groups[j]
							taken := svl.split(totalLeft, len(groups)-j)
							replicas[g.Group] = int32(taken + base)
							totalLeft -= taken
						}
						break
					}
				}
			}
		}
	}

	sum := int32(0)
	for _, r := range replicas {
		sum += r
	}
	if sum != sv.Spec.Replicas {
		panic("algorithm has bug")
	}

	return replicas
}

func (svl *ScaleViewLockManager) GrabScaleViewLockFor(sv *risingwavev1alpha1.RisingWaveScaleView) error {
	groupReplicas := svl.splitReplicasIntoGroups(sv)

	for _, s := range svl.risingwave.Status.ScaleViews {
		if s.Name == sv.Name && s.UID != sv.UID {
			return errors.New("scale view found but uid mismatch")
		}
		if s.Name == sv.Name && s.UID == sv.UID {
			return errors.New("already grabbed")
		}

		if s.Component == sv.Spec.TargetRef.Component {
			lockedGroups := lo.Map(s.GroupLocks, func(t risingwavev1alpha1.RisingWaveScaleViewLockGroupLock, _ int) string { return t.Name })
			for _, sp := range sv.Spec.ScalePolicy {
				if lo.Contains(lockedGroups, sp.Group) {
					return errors.New("lock conflict on group " + sp.Group + ", already locked by " + s.Name)
				}
			}
		}
	}

	svl.risingwave.Status.ScaleViews = append(svl.risingwave.Status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
		Name:       sv.Name,
		UID:        sv.UID,
		Component:  sv.Spec.TargetRef.Component,
		Generation: sv.Generation,
		GroupLocks: lo.Map(sv.Spec.ScalePolicy, func(t risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy, _ int) risingwavev1alpha1.RisingWaveScaleViewLockGroupLock {
			return risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
				Name:     t.Group,
				Replicas: groupReplicas[t.Group],
			}
		}),
	})

	return nil
}

func (svl *ScaleViewLockManager) GrabOrUpdateScaleViewLockFor(sv *risingwavev1alpha1.RisingWaveScaleView) (bool, error) {
	lock := svl.GetScaleViewLock(sv)
	if lock == nil {
		err := svl.GrabScaleViewLockFor(sv)
		if err != nil {
			return false, err
		}
		return true, nil
	} else {
		if lock.Generation == sv.Generation {
			return false, nil
		}

		groupReplicas := svl.splitReplicasIntoGroups(sv)
		lock.Generation = sv.Generation
		lock.GroupLocks = lo.Map(sv.Spec.ScalePolicy, func(t risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy, _ int) risingwavev1alpha1.RisingWaveScaleViewLockGroupLock {
			return risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
				Name:     t.Group,
				Replicas: groupReplicas[t.Group],
			}
		})
		return true, nil
	}
}

func (svl *ScaleViewLockManager) ReleaseLockFor(sv *risingwavev1alpha1.RisingWaveScaleView) bool {
	i, lock := svl.getScaleViewLockIndex(sv)
	if lock != nil {
		svl.risingwave.Status.ScaleViews = append(svl.risingwave.Status.ScaleViews[:i], svl.risingwave.Status.ScaleViews[i+1:]...)
		return true
	}
	return false
}

func NewScaleViewLockManager(risingwave *risingwavev1alpha1.RisingWave) *ScaleViewLockManager {
	return &ScaleViewLockManager{risingwave: risingwave}
}
