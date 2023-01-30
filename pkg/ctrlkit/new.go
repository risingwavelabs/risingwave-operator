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

package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

// ActionFunc is an alias for func(context.Context) (ctrl.Result, error).
type ActionFunc func(context.Context) (ctrl.Result, error)

type action struct {
	description string
	actionFunc  ActionFunc
}

// Description implements the Action.
func (w *action) Description() string {
	return w.description
}

// Run implements the Action.
func (w *action) Run(ctx context.Context) (ctrl.Result, error) {
	return w.actionFunc(ctx)
}

// NewAction wraps the given description and function into an action.
func NewAction(description string, f ActionFunc) Action {
	if f == nil {
		panic("action func must be provided")
	}

	return &action{description: description, actionFunc: f}
}
