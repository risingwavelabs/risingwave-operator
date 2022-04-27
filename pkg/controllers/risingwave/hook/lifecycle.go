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
// TODO: support upgrade.
func GenLifeCycleEvent(phase v1alpha1.ComponentPhase, targetRS, currentRS int32) LifeCycleEvent {
	if len(phase) == 0 {
		return LifeCycleEvent{
			Type: CreateType,
		}
	}

	switch phase {
	case v1alpha1.ComponentInitializing, v1alpha1.ComponentScaling, v1alpha1.ComponentUpgrading:
		return LifeCycleEvent{
			Type: HealthCheckType,
		}
	case v1alpha1.ComponentFailed:
		return LifeCycleEvent{
			Type: HealthCheckType,
		}
	case v1alpha1.ComponentReady:
		// if phase == Ready, but spec.rs != status.rs, means "Scale"
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
	default:
		return LifeCycleEvent{
			Type: SkipType,
		}

	}
}
