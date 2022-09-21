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
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func TestScaleViewLockManager_splitReplicasIntoGroups(t *testing.T) {
	testcases := map[string]struct {
		sv       risingwavev1alpha1.RisingWaveScaleView
		expected map[string]int32
	}{
		"one": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 10,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "",
						},
					},
				},
			},
			expected: map[string]int32{
				"": 10,
			},
		},
		"two-unlimited": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 9,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "a",
						},
						{
							Group: "b",
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 5,
				"b": 4,
			},
		},
		"two-but-one-limited": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 9,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group: "a",
						},
						{
							Group:       "b",
							MaxReplicas: pointer.Int32(2),
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 7,
				"b": 2,
			},
		},
		"three-with-priorities-1": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 9,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    10,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 3,
				"b": 2,
				"c": 4,
			},
		},
		"three-with-priorities-2": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 9,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:    "a",
							Priority: 10,
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 9,
				"b": 0,
				"c": 0,
			},
		},
		"three-with-priorities-3": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 4,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    10,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 3,
				"b": 1,
				"c": 0,
			},
		},
		"three-with-priorities-4": {
			sv: risingwavev1alpha1.RisingWaveScaleView{
				Spec: risingwavev1alpha1.RisingWaveScaleViewSpec{
					Replicas: 4,
					ScalePolicy: []risingwavev1alpha1.RisingWaveScaleViewSpecScalePolicy{
						{
							Group:       "a",
							Priority:    5,
							MaxReplicas: pointer.Int32(3),
						},
						{
							Group:       "b",
							Priority:    5,
							MaxReplicas: pointer.Int32(2),
						},
						{
							Group:    "c",
							Priority: 1,
						},
					},
				},
			},
			expected: map[string]int32{
				"a": 2,
				"b": 2,
				"c": 0,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r := NewScaleViewLockManager(nil).splitReplicasIntoGroups(&tc.sv)
			if !reflect.DeepEqual(r, tc.expected) {
				t.Fatalf("wrong result: %s", testutils.JsonMustPrettyPrint(r))
			}
		})
	}
}

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
