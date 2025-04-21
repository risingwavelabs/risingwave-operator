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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

func TestScaleViewLockManager_GrabOrUpdateScaleViewLockFor(t *testing.T) {
	scaleView := &risingwavev1alpha1.RisingWaveScaleView{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "scale-view",
			UID:        uuid.NewUUID(),
			Generation: 2,
		},
		Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
			ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
				{Group: ""},
			},
		},
	}

	testcases := map[string]struct {
		risingwave       *risingwavev1alpha1.RisingWave
		grabbedOrUpdated bool
		returnErr        bool
	}{
		"not-locked-0": {
			risingwave:       &risingwavev1alpha1.RisingWave{},
			grabbedOrUpdated: true,
			returnErr:        false,
		},
		"locked-no-update": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name:       scaleView.Name,
							UID:        scaleView.UID,
							Generation: 2,
						},
					},
				},
			},
			grabbedOrUpdated: false,
			returnErr:        false,
		},
		"locked-update": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name:       scaleView.Name,
							UID:        scaleView.UID,
							Generation: 1,
						},
					},
				},
			},
			grabbedOrUpdated: true,
			returnErr:        false,
		},
		"not-locked-conflict-0": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			grabbedOrUpdated: false,
			returnErr:        true,
		},
		"not-locked-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			grabbedOrUpdated: true,
			returnErr:        false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lockMgr := NewScaleViewLockManager(tc.risingwave)
			ok, err := lockMgr.GrabOrUpdateScaleViewLockFor(scaleView)
			assert.Equal(t, ok, tc.grabbedOrUpdated, "grab or update status not match")
			assert.Equal(t, tc.returnErr, err != nil, "error status not match: "+fmt.Sprintf("%v", err))

			if ok {
				assert.True(t, lockMgr.IsScaleViewLocked(scaleView), "must be locked")
			}
		})
	}
}

func TestScaleViewLockManager_ReleaseLockFor(t *testing.T) {
	scaleView := &risingwavev1alpha1.RisingWaveScaleView{
		ObjectMeta: metav1.ObjectMeta{
			Name: "scale-view",
			UID:  uuid.NewUUID(),
		},
	}

	testcases := map[string]struct {
		risingwave *risingwavev1alpha1.RisingWave
		released   bool
	}{
		"locked-0": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  scaleView.UID,
						},
					},
				},
			},
			released: true,
		},
		"locked-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
						{
							Name: scaleView.Name,
							UID:  scaleView.UID,
						},
					},
				},
			},
			released: true,
		},
		"no-lock": {
			risingwave: &risingwavev1alpha1.RisingWave{},
			released:   false,
		},
		"not-locked-0": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			released: false,
		},
		"not-locked-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			released: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lockMgr := NewScaleViewLockManager(tc.risingwave)
			assert.Equal(t, tc.released, lockMgr.ReleaseLockFor(scaleView), "lock status not match")
			assert.False(t, lockMgr.IsScaleViewLocked(scaleView), "must be unlocked after release")
		})
	}
}

func TestScaleViewLockManager_GrabScaleViewLockFor(t *testing.T) {
	scaleView := &risingwavev1alpha1.RisingWaveScaleView{
		ObjectMeta: metav1.ObjectMeta{
			Name:       "scale-view",
			UID:        uuid.NewUUID(),
			Generation: 2,
		},
		Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
			ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
				{Group: ""},
			},
		},
	}

	testcases := map[string]struct {
		risingwave *risingwavev1alpha1.RisingWave
		returnErr  bool
	}{
		"not-locked-0": {
			risingwave: &risingwavev1alpha1.RisingWave{},
			returnErr:  false,
		},
		"not-locked-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			returnErr: false,
		},
		"locked-no-update": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name:       scaleView.Name,
							UID:        scaleView.UID,
							Generation: 2,
						},
					},
				},
			},
			returnErr: true,
		},
		"locked-update": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name:       scaleView.Name,
							UID:        scaleView.UID,
							Generation: 1,
						},
					},
				},
			},
			returnErr: true,
		},
		"not-locked-conflict-0": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			returnErr: true,
		},
		"not-locked-conflict-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
							GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
								{Name: ""},
							},
						},
					},
				},
			},
			returnErr: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lockMgr := NewScaleViewLockManager(tc.risingwave)
			err := lockMgr.GrabScaleViewLockFor(scaleView)
			assert.Equal(t, tc.returnErr, err != nil, "error status not match: "+fmt.Sprintf("%v", err))

			if err == nil {
				assert.True(t, lockMgr.IsScaleViewLocked(scaleView), "must be locked")
			}
		})
	}
}

func TestScaleViewLockManager_IsScaleViewLocked(t *testing.T) {
	scaleView := &risingwavev1alpha1.RisingWaveScaleView{
		ObjectMeta: metav1.ObjectMeta{
			Name: "scale-view",
			UID:  uuid.NewUUID(),
		},
	}

	testcases := map[string]struct {
		risingwave *risingwavev1alpha1.RisingWave
		locked     bool
	}{
		"matches-0": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  scaleView.UID,
						},
					},
				},
			},
			locked: true,
		},
		"matches-1": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
						{
							Name: scaleView.Name,
							UID:  scaleView.UID,
						},
					},
				},
			},
			locked: true,
		},
		"no-lock": {
			risingwave: &risingwavev1alpha1.RisingWave{},
			locked:     false,
		},
		"not-match-on-uid": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: scaleView.Name,
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			locked: false,
		},
		"no-matching-lock": {
			risingwave: &risingwavev1alpha1.RisingWave{
				Status: risingwavev1alpha1.RisingWaveStatus{
					ScaleViews: []risingwavev1alpha1.RisingWaveScaleViewLock{
						{
							Name: "random",
							UID:  uuid.NewUUID(),
						},
					},
				},
			},
			locked: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lockMgr := NewScaleViewLockManager(tc.risingwave)
			assert.Equal(t, tc.locked, lockMgr.IsScaleViewLocked(scaleView), "lock status not match")
		})
	}
}
