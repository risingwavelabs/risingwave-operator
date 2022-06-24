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

import "github.com/samber/lo"

func unwrapParallel(act Action) Action {
	for {
		switch innerAct := act.(type) {
		case *parallelAction:
			act = innerAct.inner
		default:
			return act
		}
	}
}

// OptimizeWorkflow optimizes the workflow by eliminating unnecessary layers:
//   * Nop in Sequential or Join will be removed
//   * Empty Join and Sequential will be omitted
//   * Parallel in Sequential will be unwrapped
//   * Sequential and Join with single child will be simplified by removing them
//   * Sequential(Sequential) will be flattened
//   * Join(Join) will be flattened
//   * Parallel(Parallel) will be simplified with only one Parallel
//   * Timeout(Timeout) will be simplified with the tighter timeout
func OptimizeWorkflow(workflow Action) Action {
	switch workflow := workflow.(type) {
	case *sequentialGroup:
		actions := make([]Action, 0)
		for _, act := range workflow.actions {
			act = OptimizeWorkflow(act)
			act = unwrapParallel(act)
			switch innerAct := act.(type) {
			case *sequentialGroup:
				actions = append(actions, innerAct.actions...)
			default:
				actions = append(actions, act)
			}
		}
		workflow.actions = lo.Filter(actions, func(act Action, _ int) bool {
			return act != Nop
		})
		if len(workflow.actions) == 0 {
			return Nop
		}
		if len(workflow.actions) == 1 {
			return workflow.actions[0]
		}
		return workflow
	case *joinGroup:
		actions := make([]Action, 0)
		for i, act := range workflow.actions {
			workflow.actions[i] = OptimizeWorkflow(act)
			// If they are the same Join type, lift the inner one into the outer one.
			if innerJoin, ok := act.(*joinGroup); ok {
				if innerJoin.runner.IsParallel() == workflow.runner.IsParallel() {
					actions = append(actions, innerJoin.actions...)
				} else {
					actions = append(actions, act)
				}
			} else {
				actions = append(actions, act)
			}
		}
		workflow.actions = lo.Filter(actions, func(act Action, _ int) bool {
			return act != Nop
		})
		if len(workflow.actions) == 0 {
			return Nop
		}
		if len(workflow.actions) == 1 {
			if workflow.runner.IsParallel() {
				return OptimizeWorkflow(Parallel(workflow.actions[0]))
			} else {
				return workflow.actions[0]
			}
		}
		return workflow
	case *parallelAction:
		workflow.inner = OptimizeWorkflow(workflow.inner)
		if workflow.inner == Nop {
			return Nop
		}
		if _, ok := workflow.inner.(*parallelAction); ok {
			return workflow.inner
		} else {
			return workflow
		}
	case *retryAction:
		workflow.inner = OptimizeWorkflow(workflow.inner)
		if workflow.inner == Nop {
			return Nop
		}
		return workflow
	case *timeoutAction:
		workflow.inner = OptimizeWorkflow(workflow.inner)
		if workflow.inner == Nop {
			return Nop
		}
		if innerTimeout, ok := workflow.inner.(*timeoutAction); ok {
			if innerTimeout.timeout > workflow.timeout {
				innerTimeout.timeout = workflow.timeout
			}
			return innerTimeout
		} else {
			return workflow
		}
	default:
		return workflow
	}
}
