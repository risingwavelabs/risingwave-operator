package risingwave

import (
	"context"
	"github.com/go-logr/logr"
	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	if action == RisingWaveAction_BarrierConditionInitializingIsTrue {
		h.recorder.Event(h.rw, corev1.EventTypeNormal, action, "Initializing")
	}
}
