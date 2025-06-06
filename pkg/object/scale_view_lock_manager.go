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

package object

import (
	"errors"

	"github.com/samber/lo"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/object/scaleview"
)

// ScaleViewLockManager is a struct to help manage the in memory lock records under RisingWave status.
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

// GetScaleViewLock gets the scale view lock record from the RisingWave for the specified RisingWaveScaleView. It returns
// nil when not found.
func (svl *ScaleViewLockManager) GetScaleViewLock(sv *risingwavev1alpha1.RisingWaveScaleView) *risingwavev1alpha1.RisingWaveScaleViewLock {
	_, r := svl.getScaleViewLockIndex(sv)

	return r
}

// IsScaleViewLocked checks if the RisingWaveScaleView locks the target RisingWave.
func (svl *ScaleViewLockManager) IsScaleViewLocked(sv *risingwavev1alpha1.RisingWaveScaleView) bool {
	return svl.GetScaleViewLock(sv) != nil
}

func (svl *ScaleViewLockManager) splitReplicasIntoGroups(sv *risingwavev1alpha1.RisingWaveScaleView) map[string]int32 {
	return scaleview.SplitReplicas(sv)
}

// GrabScaleViewLockFor grabs the lock for the specified RisingWaveScaleView by creating a new scale view record in the
// RisingWave status. It will return error when fails to grab the lock, e.g., it's already locked or there are conflicts.
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
					return errors.New("lock conflict on group " + sp.Group + ", already returnErr by " + s.Name)
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

// GrabOrUpdateScaleViewLockFor creates or updates the scale view record in the RisingWave object for the specified
// RisingWaveScaleView. It returns true when the lock is grabbed or updated. If nothing happens, it should return false, nil.
func (svl *ScaleViewLockManager) GrabOrUpdateScaleViewLockFor(sv *risingwavev1alpha1.RisingWaveScaleView) (bool, error) {
	lock := svl.GetScaleViewLock(sv)

	if lock == nil {
		err := svl.GrabScaleViewLockFor(sv)
		if err != nil {
			return false, err
		}

		return true, nil
	}

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

// ReleaseLockFor releases the lock for the specified RisingWaveScaleView in memory, by removing the scale view record
// from the status field of the RisingWave object.
func (svl *ScaleViewLockManager) ReleaseLockFor(sv *risingwavev1alpha1.RisingWaveScaleView) bool {
	i, lock := svl.getScaleViewLockIndex(sv)
	if lock != nil {
		svl.risingwave.Status.ScaleViews = append(svl.risingwave.Status.ScaleViews[:i], svl.risingwave.Status.ScaleViews[i+1:]...)

		return true
	}

	return false
}

// NewScaleViewLockManager creates a new ScaleViewLockManager with the given target.
func NewScaleViewLockManager(risingwave *risingwavev1alpha1.RisingWave) *ScaleViewLockManager {
	return &ScaleViewLockManager{risingwave: risingwave}
}
