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

package ctrlkit

import (
	"context"

	ctrl "sigs.k8s.io/controller-runtime"
)

type sequentialActions struct {
	actions []ReconcileAction
}

func (act *sequentialActions) Description() string {
	return describeGroup("Sequential", act.actions...)
}

func (act *sequentialActions) Run(ctx context.Context) (ctrl.Result, error) {
	// Run actions one-by-one. If one action needs to requeue or requeue after, then the
	// control flow is broken and control is returned to the outer scope.
	for _, act := range act.actions {
		result, err := act.Run(ctx)
		if NeedsRequeue(result, err) {
			return result, err
		}
	}

	return NoRequeue()
}

// Sequential organizes the actions into a sequential flow.
func Sequential(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 0 {
		panic("must provide actions to sequential")
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	return &sequentialActions{actions: actions}
}
