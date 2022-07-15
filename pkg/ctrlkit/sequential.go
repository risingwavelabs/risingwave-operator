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

	"github.com/singularity-data/risingwave-operator/pkg/ctrlkit/internal"
)

var _ internal.Group = &sequentialGroup{}

type sequentialGroup struct {
	actions []Action
}

func (grp *sequentialGroup) Children() []Action {
	return grp.actions
}

func (grp *sequentialGroup) SetChildren(actions []Action) {
	grp.actions = actions
}

func (grp *sequentialGroup) Name() string {
	return "Sequential"
}

func (grp *sequentialGroup) Description() string {
	return internal.DescribeGroup(grp.Name(), grp.actions...)
}

func (grp *sequentialGroup) Run(ctx context.Context) (ctrl.Result, error) {
	// Run actions one-by-one. If one action needs to requeue or requeue after, then the
	// control flow is broken and control is returned to the outer scope.
	for _, act := range grp.actions {
		result, err := act.Run(ctx)
		if NeedsRequeue(result, err) {
			return result, err
		}
	}

	return NoRequeue()
}

// Sequential organizes the actions into a sequential flow.
func Sequential(actions ...Action) Action {
	if len(actions) == 0 {
		return Nop
	}

	// Simply return the first action if there's only one.
	if len(actions) == 1 {
		return actions[0]
	}

	return &sequentialGroup{actions: actions}
}

// SequentialJoin is an alias of JoinOrdered, because it runs the actions in both join and sequential style.
// It is useful to use `SequentialJoin` to declare something that must run after something else. E.g., if
// B must run after A, but you want B to run no matter what A returns, you can use
// SequentialJoin(A, B) instead of Sequential(A, B).
var SequentialJoin = OrderedJoin
