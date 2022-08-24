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

package consts

import corev1 "k8s.io/api/core/v1"

type RisingWaveEventType struct {
	Name string
	Type string
}

// Valid event types.
var (
	RisingWaveEventTypeInitializing = RisingWaveEventType{Name: "Initializing", Type: corev1.EventTypeNormal}
	RisingWaveEventTypeRunning      = RisingWaveEventType{Name: "Running", Type: corev1.EventTypeNormal}
	RisingWaveEventTypeUnhealthy    = RisingWaveEventType{Name: "Unhealthy", Type: corev1.EventTypeWarning}
	RisingWaveEventTypeRecovering   = RisingWaveEventType{Name: "Recovering", Type: corev1.EventTypeNormal}
	RisingWaveEventTypeUpgrading    = RisingWaveEventType{Name: "Upgrading", Type: corev1.EventTypeNormal}
)
