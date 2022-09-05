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

package controller

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"k8s.io/client-go/tools/record"

	"github.com/fatih/color"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

const defaultRecorderBufferSize = 10

type resultErr struct {
	reconcile.Result
	error
}

func (r *resultErr) Equals(result reconcile.Result, err error) bool {
	return r.Result == result && errors.Is(err, r.error)
}

func newResultErr(r reconcile.Result, err error) resultErr {
	return resultErr{Result: r, error: err}
}

type actionAssertionHook struct {
	t         *testing.T
	asserts   map[string]resultErr
	mustCover bool
}

func (h *actionAssertionHook) PostRun(ctx context.Context, logger logr.Logger, action string, result reconcile.Result, err error) {
	resultErr, ok := h.asserts[action]
	if !ok {
		return
	}

	if !resultErr.Equals(result, err) {
		fmt.Printf("%s\t[%s]\n", color.RedString("FAIL"), action)
		h.t.Fatalf("unexpected result and error: %v,%v, expected: %v,%v, action: %s", result, err, resultErr.Result, resultErr.error, action)
	} else {
		fmt.Printf("%s\t[%s]\n", color.GreenString("PASS"), action)
	}
}

func (h *actionAssertionHook) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {
	if _, ok := h.asserts[action]; !ok {
		if h.mustCover {
			h.t.Fatalf("unexpected action: %s", action)
		}
	}
}

func newActionAsserts(t *testing.T, asserts map[string]resultErr, mustCover bool) ctrlkit.ActionHook {
	return &actionAssertionHook{
		t:         t,
		asserts:   asserts,
		mustCover: mustCover,
	}
}

func takeAndLogEventsFromFakeRecord(t *testing.T, recorder *record.FakeRecorder) {
	t.Logf("events: \n")
	for i := 0; i < len(recorder.Events); i++ {
		event := <-recorder.Events
		t.Logf("  - %s\n", event)
	}
}

func Test_RisingWaveController_New(t *testing.T) {
	risingwave := &risingwavev1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example",
			Namespace: "default",
		},
		Spec:   risingwavev1alpha1.RisingWaveSpec{},
		Status: risingwavev1alpha1.RisingWaveStatus{},
	}

	recorder := record.NewFakeRecorder(defaultRecorderBufferSize)
	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(testutils.Schema).
			WithObjects(risingwave).
			Build(),
		ActionHookFactory: func() ctrlkit.ActionHook {
			return newActionAsserts(t, map[string]resultErr{
				// New => Initializing(true), Running(false)
				RisingWaveAction_BarrierFirstTimeObserved:        newResultErr(ctrlkit.Continue()),
				RisingWaveAction_MarkConditionInitializingAsTrue: newResultErr(ctrlkit.Continue()),
				RisingWaveAction_MarkConditionRunningAsFalse:     newResultErr(ctrlkit.Continue()),

				// Running(false) => stop
				RisingWaveAction_BarrierConditionRunningIsFalse: newResultErr(ctrlkit.Exit()),

				// Initializing(true) => stop
				RisingWaveAction_BarrierConditionInitializingIsTrue: newResultErr(ctrlkit.Exit()),

				// Running(true) => stop
				RisingWaveAction_BarrierConditionRunningIsTrue: newResultErr(ctrlkit.Exit()),

				// Upgrading(true) => stop
				RisingWaveAction_BarrierConditionUpgradingIsTrue: newResultErr(ctrlkit.Exit()),

				// Sync status with running stats
				RisingWaveAction_CollectRunningStatisticsAndSyncStatus: newResultErr(ctrlkit.Continue()),

				// Update status
				RisingWaveAction_UpdateRisingWaveStatusViaClient: newResultErr(ctrlkit.Continue()),
			}, false)
		},
		Recorder: recorder,
	}

	logger := zap.New(zap.UseDevMode(true))
	_, err := controller.Reconcile(log.IntoContext(context.Background(), logger), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example",
			Namespace: "default",
		},
	})

	numEvents := len(recorder.Events)
	numWantedEvents := 1
	if numEvents != numWantedEvents {
		takeAndLogEventsFromFakeRecord(t, recorder)
		t.Errorf("got %d events, wanted %d", numEvents, numWantedEvents)
	}

	if err != nil {
		t.Fatal(err)
	}
}

func Test_RisingWaveController_Deleted(t *testing.T) {
	risingwave := &risingwavev1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "example",
			Namespace:         "default",
			DeletionTimestamp: &metav1.Time{Time: time.Now()},
		},
		Spec:   risingwavev1alpha1.RisingWaveSpec{},
		Status: risingwavev1alpha1.RisingWaveStatus{},
	}

	recorder := record.NewFakeRecorder(defaultRecorderBufferSize)
	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(testutils.Schema).
			WithObjects(risingwave).
			Build(),
		ActionHookFactory: func() ctrlkit.ActionHook {
			return newActionAsserts(t, nil, true)
		},
	}

	logger := zap.New(zap.UseDevMode(true))
	_, err := controller.Reconcile(log.IntoContext(context.Background(), logger), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example",
			Namespace: "default",
		},
	})

	numEvents := len(recorder.Events)
	numWantedEvents := 0
	if numEvents != numWantedEvents {
		takeAndLogEventsFromFakeRecord(t, recorder)
		t.Errorf("got %d events, wanted %d", numEvents, numWantedEvents)
	}

	if err != nil {
		t.Fatal(err)
	}
}

func Test_RisingWaveController_Initializing(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	risingwave.Status = risingwavev1alpha1.RisingWaveStatus{
		Conditions: []risingwavev1alpha1.RisingWaveCondition{
			{
				Type:   risingwavev1alpha1.RisingWaveConditionInitializing,
				Status: metav1.ConditionTrue,
			},
		},
	}

	recorder := record.NewFakeRecorder(defaultRecorderBufferSize)
	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(testutils.Schema).
			WithObjects(risingwave).
			Build(),
		Recorder: recorder,
		ActionHookFactory: func() ctrlkit.ActionHook {
			return newActionAsserts(t, map[string]resultErr{
				RisingWaveAction_BarrierConditionInitializingIsTrue:  newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncMetaService:                     newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncComputeService:                  newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncCompactorService:                newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncFrontendService:                 newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncConfigConfigMap:                 newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncMetaDeployments:                 newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncFrontendDeployments:             newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncCompactorDeployments:            newResultErr(ctrlkit.Continue()),
				RisingWaveAction_SyncComputeStatefulSets:             newResultErr(ctrlkit.Continue()),
				RisingWaveAction_WaitBeforeMetaDeploymentsReady:      newResultErr(ctrlkit.Exit()),
				RisingWaveAction_WaitBeforeFrontendDeploymentsReady:  newResultErr(ctrlkit.Exit()),
				RisingWaveAction_WaitBeforeComputeStatefulSetsReady:  newResultErr(ctrlkit.Exit()),
				RisingWaveAction_WaitBeforeCompactorDeploymentsReady: newResultErr(ctrlkit.Exit()),
				RisingWaveAction_SyncObservedGeneration:              newResultErr(ctrlkit.Continue()),
			}, false)
		},
	}

	logger := zap.New(zap.UseDevMode(true))
	_, err := controller.Reconcile(log.IntoContext(context.Background(), logger), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: risingwave.Name, Namespace: risingwave.Namespace},
	})

	numEvents := len(recorder.Events)
	numWantedEvents := 0
	if numEvents != numWantedEvents {
		takeAndLogEventsFromFakeRecord(t, recorder)
		t.Errorf("got %d events, wanted %d", numEvents, numWantedEvents)
	}

	if err != nil {
		t.Fatal(err)
	}
}

func Test_RisingWaveController_Recovery(t *testing.T) {
	risingwave := testutils.FakeRisingWave()
	risingwave.Status.ObservedGeneration = risingwave.Generation

	recorder := record.NewFakeRecorder(defaultRecorderBufferSize)
	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(testutils.Schema).
			WithObjects(risingwave).
			Build(),
		Recorder: recorder,
		ActionHookFactory: func() ctrlkit.ActionHook {
			return newActionAsserts(t, map[string]resultErr{
				RisingWaveAction_MarkConditionUpgradingAsTrue: newResultErr(ctrlkit.Continue()),
			}, false)
		},
	}

	logger := zap.New(zap.UseDevMode(true))
	_, err := controller.Reconcile(log.IntoContext(context.Background(), logger), reconcile.Request{
		NamespacedName: types.NamespacedName{Name: risingwave.Name, Namespace: risingwave.Namespace},
	})

	var currentRisingwave risingwavev1alpha1.RisingWave
	if err := controller.Client.Get(context.Background(), types.NamespacedName{
		Name:      risingwave.Name,
		Namespace: risingwave.Namespace,
	}, &currentRisingwave); err != nil {
		t.Fatal(err)
	}

	risingwaveManager := object.NewRisingWaveManager(nil, &currentRisingwave)
	runningCondition := risingwaveManager.GetCondition(risingwavev1alpha1.RisingWaveConditionRunning)
	if runningCondition == nil || runningCondition.Status != metav1.ConditionFalse {
		t.Logf("condition: %s", testutils.JsonMustPrettyPrint(runningCondition))
		t.Fatal("Running condition not false")
	}

	numEvents := len(recorder.Events)
	numWantedEvents := 2
	if numEvents != numWantedEvents {
		takeAndLogEventsFromFakeRecord(t, recorder)
		t.Errorf("got %d events, wanted %d", numEvents, numWantedEvents)
	}

	if err != nil {
		t.Fatal(err)
	}
}
