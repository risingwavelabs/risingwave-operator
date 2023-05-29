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
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_NewRisingWaveManager(t *testing.T) {
	mgr := NewRisingWaveManager(nil, testutils.FakeRisingWave(), false)

	if mgr.risingwave == mgr.mutableRisingWave {
		t.Fail()
	}

	if !testutils.DeepEqual(mgr.risingwave, mgr.mutableRisingWave) {
		t.Fail()
	}
}

func Test_RisingWaveManager_UpdateRemote(t *testing.T) {
	risingwave := testutils.FakeRisingWave()

	mgr := NewRisingWaveManager(
		fake.NewClientBuilder().
			WithScheme(testutils.Scheme).
			WithStatusSubresource(&risingwavev1alpha1.RisingWave{}).
			WithObjects(risingwave).
			Build(),
		risingwave,
		false,
	)

	// Should do nothing.
	if mgr.UpdateRemoteRisingWaveStatus(context.Background()) != nil {
		t.Fail()
	}

	// Mutate and check.
	mgr.SyncObservedGeneration()

	if testutils.DeepEqual(mgr.risingwave.Status, mgr.mutableRisingWave.Status) {
		t.Fail()
	}

	// Remote update and check.
	if err := mgr.UpdateRemoteRisingWaveStatus(context.Background()); err != nil {
		t.Fatal(err)
	}

	var currentRisingWave risingwavev1alpha1.RisingWave
	if err := mgr.client.Get(context.Background(), types.NamespacedName{
		Namespace: mgr.risingwave.Namespace,
		Name:      mgr.risingwave.Name,
	}, &currentRisingWave); err != nil {
		t.Fatal(err)
	}

	if currentRisingWave.Status.ObservedGeneration != risingwave.Generation {
		t.Fail()
	}
}

func Test_RisingWaveManager_openKruiseAvailable(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	mgrWithOpenKruiseUnavailable := NewRisingWaveManager(nil, risingwave, false)
	mgrWithOpenKruiseAvailable := NewRisingWaveManager(nil, risingwave, true)
	if mgrWithOpenKruiseUnavailable.IsOpenKruiseAvailable() {
		t.Fail()
	}
	if !mgrWithOpenKruiseAvailable.IsOpenKruiseAvailable() {
		t.Fail()
	}

}

func Test_RisingWaveManager_OpenKruiseEnabled(t *testing.T) {
	testcases := map[string]struct {
		mgr      *RisingWaveManager
		expected bool
	}{
		"open-kruise-not-available": {
			mgr:      NewRisingWaveManager(nil, testutils.FakeRisingWave(), false),
			expected: false,
		},
		"open-kruise-available-disabled": {
			mgr:      NewRisingWaveManager(nil, testutils.FakeRisingWaveOpenKruiseDisabled(), true),
			expected: false,
		},
		"open-kruise-available-enabled": {
			mgr:      NewRisingWaveManager(nil, testutils.FakeRisingWaveOpenKruiseEnabled(), true),
			expected: true,
		},
		"open-kruise-unavailable-enabled": {
			mgr:      NewRisingWaveManager(nil, testutils.FakeRisingWaveOpenKruiseDisabled(), false),
			expected: false},
	}
	for _, tc := range testcases {
		if tc.mgr.IsOpenKruiseEnabled() != tc.expected {
			t.Fail()
		}
	}
}

func Test_RisingWaveManager_UpdateMemoryAndGet(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	mgr := NewRisingWaveManager(nil, risingwave, false)

	// IsObservedGenerationOutdated
	if !mgr.IsObservedGenerationOutdated() {
		t.Fail()
	}

	// SyncObservedGeneration
	mgr.SyncObservedGeneration()
	if mgr.mutableRisingWave.Status.ObservedGeneration != risingwave.Generation {
		t.Fail()
	}

	// UpdateCondition exists
	mgr.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
		Type:   risingwavev1alpha1.RisingWaveConditionRunning,
		Status: metav1.ConditionFalse,
	})
	if mgr.GetCondition(risingwavev1alpha1.RisingWaveConditionRunning).Status != metav1.ConditionTrue {
		t.Fail()
	}
	if mgr.GetCondition(risingwavev1alpha1.RisingWaveConditionFailed) != nil {
		t.Fail()
	}

	// RemoveCondition exists
	mgr.RemoveCondition(risingwavev1alpha1.RisingWaveConditionRunning)
	if mgr.GetCondition(risingwavev1alpha1.RisingWaveConditionRunning) == nil {
		t.Fail()
	}
	if len(mgr.mutableRisingWave.Status.Conditions) != 0 {
		t.Fail()
	}

	// UpdateCondition new
	mgr.UpdateCondition(risingwavev1alpha1.RisingWaveCondition{
		Type:   risingwavev1alpha1.RisingWaveConditionFailed,
		Status: metav1.ConditionTrue,
	})
	if len(mgr.mutableRisingWave.Status.Conditions) == 0 {
		t.Fail()
	}
	if mgr.mutableRisingWave.Status.Conditions[0].Type != risingwavev1alpha1.RisingWaveConditionFailed ||
		mgr.mutableRisingWave.Status.Conditions[0].Status != metav1.ConditionTrue ||
		mgr.mutableRisingWave.Status.Conditions[0].LastTransitionTime.IsZero() {
		t.Fail()
	}

	// UpdateStatus
	mgr.UpdateStatus(func(rws *risingwavev1alpha1.RisingWaveStatus) {
		rws.ComponentReplicas.Meta.Running = 0
	})
	if mgr.mutableRisingWave.Status.ComponentReplicas.Meta.Running != 0 {
		t.Fail()
	}

	// RisingWave
	if !testutils.DeepEqual(mgr.RisingWave(), risingwave) {
		t.Fail()
	}
}
