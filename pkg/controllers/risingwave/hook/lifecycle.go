package hook

import "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"

type LifeCycleType string

const (
	SkipType        LifeCycleType = "skip"
	CreateType      LifeCycleType = "create"
	UpgradeType     LifeCycleType = "upgrade"
	ScaleUpType     LifeCycleType = "scale-up"
	ScaleDownType   LifeCycleType = "scale-down"
	HealthCheckType LifeCycleType = "health-check"
)

type LifeCycleEvent struct {
	Type LifeCycleType
}

type LifeCycleHook func() error

type LifeCycleOption struct {
	PreUpdateFunc LifeCycleHook

	PostUpdateFunc LifeCycleHook

	PostReadyFunc LifeCycleHook
}

// GenLifeCycleEvent will gen lifeCycleEvent according to phase and rs.
// TODO: support upgrade
func GenLifeCycleEvent(phase v1alpha1.ComponentPhase, targetRS, currentRS int32) LifeCycleEvent {
	if len(phase) == 0 {
		return LifeCycleEvent{
			Type: CreateType,
		}
	}

	if phase == v1alpha1.ComponentInitializing {
		return LifeCycleEvent{
			Type: HealthCheckType,
		}
	}

	if phase == v1alpha1.ComponentFailed {
		return LifeCycleEvent{
			Type: HealthCheckType,
		}
	}

	// if phase == Ready, but spec.rs != status.rs, means "Scale"
	if phase == v1alpha1.ComponentReady {
		if targetRS == currentRS {
			return LifeCycleEvent{
				Type: SkipType,
			}
		}
		if targetRS < currentRS {
			return LifeCycleEvent{
				Type: ScaleDownType,
			}
		}
		return LifeCycleEvent{
			Type: ScaleUpType,
		}
	}

	return LifeCycleEvent{
		Type: SkipType,
	}
}
