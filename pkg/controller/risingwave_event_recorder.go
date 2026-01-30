// Copyright 2023 RisingWave Labs
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

package controller

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/events"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/event"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/object"
)

// RisingWaveEventRecorder is an action hook for recording events.
type RisingWaveEventRecorder struct {
	recorder events.EventRecorder
	mgr      *object.RisingWaveManager
	msgStore *event.MessageStore
}

// NewEventHook creates an event hook for the given risingwave.
func NewEventHook(recorder events.EventRecorder, mgr *object.RisingWaveManager, msgStore *event.MessageStore) *RisingWaveEventRecorder {
	return &RisingWaveEventRecorder{recorder: recorder, mgr: mgr, msgStore: msgStore}
}

// PreRun implements the ActionHook interface.
func (h *RisingWaveEventRecorder) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {
}

// isAfterConditionTrueAndChanged is a function for checking if the condition is exactly changed in the current workflow.
func (h *RisingWaveEventRecorder) isAfterConditionTrueAndChanged(before, after *object.RisingWaveReader, eval func(r *object.RisingWaveReader) bool) bool {
	return eval(after) && !eval(before)
}

func (h *RisingWaveEventRecorder) recordEvent(event consts.RisingWaveEventType) {
	h.recorder.Eventf(h.mgr.RisingWave(), nil, event.Type, event.Name, event.Name, h.msgStore.MessageFor(event.Name))
}

func (h *RisingWaveEventRecorder) recordConditionChangingEvents() {
	before, after := &h.mgr.RisingWaveReader, object.NewRisingWaveReader(h.mgr.RisingWaveAfterImage())

	if h.isAfterConditionTrueAndChanged(before, after, func(r *object.RisingWaveReader) bool {
		return r.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionInitializing, true)
	}) {
		h.recordEvent(consts.RisingWaveEventTypeInitializing)
	}

	if h.isAfterConditionTrueAndChanged(before, after, func(r *object.RisingWaveReader) bool {
		return r.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, true)
	}) {
		h.recordEvent(consts.RisingWaveEventTypeRunning)
	}

	// Not initializing && running == false => we're recovering
	if h.isAfterConditionTrueAndChanged(before, after, func(r *object.RisingWaveReader) bool {
		return r.GetCondition(risingwavev1alpha1.RisingWaveConditionInitializing) == nil &&
			r.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionRunning, false)
	}) {
		h.recordEvent(consts.RisingWaveEventTypeRecovering)
	}

	if h.isAfterConditionTrueAndChanged(before, after, func(r *object.RisingWaveReader) bool {
		return r.DoesConditionExistAndEqual(risingwavev1alpha1.RisingWaveConditionUpgrading, true)
	}) {
		h.recordEvent(consts.RisingWaveEventTypeUpgrading)
	}
}

func (h *RisingWaveEventRecorder) recordStatesWarningEvents() {
	warningEvents := []consts.RisingWaveEventType{
		consts.RisingWaveEventTypeUnhealthy,
	}

	for _, ev := range warningEvents {
		if h.msgStore.IsMessageSet(ev.Name) {
			h.recordEvent(ev)
		}
	}
}

// PostRun implements the ActionHook interface.
func (h *RisingWaveEventRecorder) PostRun(ctx context.Context, logger logr.Logger, action string, result reconcile.Result, err error) {
	if action != RisingWaveAction_UpdateRisingWaveStatusViaClient {
		return
	}

	if err != nil {
		return
	}

	h.recordConditionChangingEvents()

	h.recordStatesWarningEvents()
}
