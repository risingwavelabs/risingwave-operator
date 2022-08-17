// Copyright 2022 Singularity Data
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package risingwave

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/object"
)

// EventHook is an action hook for recording events.
type EventHook struct {
	recorder record.EventRecorder
	mgr      *object.RisingWaveManager
}

// NewEventHook creates an event hook for the given risingwave.
func NewEventHook(recorder record.EventRecorder, mgr *object.RisingWaveManager) *EventHook {
	return &EventHook{recorder: recorder, mgr: mgr}
}

// PreRun implements the ActionHook interface.
func (h *EventHook) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {
}

// PostRun implements the ActionHook interface.
func (h *EventHook) PostRun(ctx context.Context, logger logr.Logger, action string, result reconcile.Result, err error) {
	if action != RisingWaveAction_UpdateRisingWaveStatusViaClient {
		return
	}

	if err != nil {
		return
	}

	reader := object.NewRisingWaveReader(h.mgr.RisingWaveAfterImage())

	if reader.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionInitializing, true) {
		h.recorder.Event(h.mgr.RisingWave(), corev1.EventTypeNormal, "Initializing", "Initializing")
	}

	if reader.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true) {
		h.recorder.Event(h.mgr.RisingWave(), corev1.EventTypeNormal, "Initializing", "Initializing")
	}

	// Not initializing && running == false => we're recovering
	if reader.GetCondition(risingwavev1alpha1.RisingWaveConditionInitializing) == nil &&
		reader.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, false) {
		h.recorder.Event(h.mgr.RisingWave(), corev1.EventTypeWarning, "Recovering", "Recovering")
	}

	if reader.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionUpgrading, true) {
		h.recorder.Event(h.mgr.RisingWave(), corev1.EventTypeNormal, "Upgrading", "Upgrading")
	}
}
