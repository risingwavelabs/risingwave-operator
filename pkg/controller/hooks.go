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
)

type EventHook struct {
	recorder record.EventRecorder
	rw       *risingwavev1alpha1.RisingWave // Immutable
}

func NewEventHook(recorder record.EventRecorder, rw *risingwavev1alpha1.RisingWave) *EventHook {
	return &EventHook{recorder: recorder, rw: rw}
}

func (h *EventHook) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {

}

func (h *EventHook) PostRun(ctx context.Context, logger logr.Logger, action string, result reconcile.Result, err error) {
	rwName := h.rw.Name
	if action == RisingWaveAction_BarrierConditionInitializingIsTrue {
		h.recorder.Eventf(h.rw, corev1.EventTypeNormal, "Initializing", "Initializing RisingWave instance %s", rwName)
	} else if action == RisingWaveAction_MarkConditionRunningAsTrue {
		h.recorder.Eventf(h.rw, corev1.EventTypeNormal, "Running", "RisingWave instance %s is running", rwName)
	} else if action == RisingWaveAction_MarkConditionUpgradingAsTrue {
		h.recorder.Eventf(h.rw, corev1.EventTypeNormal, "Upgrading", "Upgrading RisingWave instance %s", rwName)
	}
}
