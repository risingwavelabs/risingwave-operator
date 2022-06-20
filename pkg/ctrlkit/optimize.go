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

// OptimizeWorkflow optimizes the workflow by eliminating unnecessary layers:
//   * Nop in Sequential or Join will be removed
//   * Empty Join and Sequential will be omitted
//   * Sequentail and Join with single child will be simpilified by removing them
//   * Sequential(Sequential) will be flattened
//   * Join(Join) will be flattened
//   * Parallel(Parallel) will be simplified with only one Parallel
//   * Timeout(Timeout) will be simplified with the tighter timeout
func OptimizeWorkflow(workflow ReconcileAction) ReconcileAction {
	switch workflow := workflow.(type) {
	case *sequentialAction:
		actions := make([]ReconcileAction, 0)
		for _, act := range workflow.actions {
			act = OptimizeWorkflow(act)
			if innerSeq, ok := act.(*sequentialAction); ok {
				actions = append(actions, innerSeq.actions...)
			} else {
				actions = append(actions, act)
			}
		}
		workflow.actions = lo.Filter(actions, func(act ReconcileAction, _ int) bool {
			return act != Nop
		})
		if len(workflow.actions) == 0 {
			return Nop
		}
		if len(workflow.actions) == 1 {
			return workflow.actions[0]
		}
		return workflow
	case *joinAction:
		for i, act := range workflow.actions {
			workflow.actions[i] = OptimizeWorkflow(act)
		}
		workflow.actions = lo.Filter(workflow.actions, func(act ReconcileAction, _ int) bool {
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
		if _, ok := workflow.inner.(*parallelAction); ok {
			return workflow.inner
		} else {
			return workflow
		}
	case *retryAction:
		workflow.inner = OptimizeWorkflow(workflow.inner)
		return workflow
	case *timeoutAction:
		workflow.inner = OptimizeWorkflow(workflow.inner)
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
