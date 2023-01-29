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

package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// CreateEventFilter is the predicate for CreateEvent only.
	CreateEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// DeleteEventFilter is the predicate for DeleteEvent only.
	DeleteEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// UpdateEventFilter is the predicate for UpdateEvent only.
	UpdateEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return true
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// GenericEventFilter is the predicate for GenericEvent only.
	GenericEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return true
		},
	}
)
