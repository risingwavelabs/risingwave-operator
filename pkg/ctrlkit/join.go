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
	"sync"

	"github.com/samber/lo"
	"go.uber.org/multierr"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

// Join the errors with the following rules:
//   - any + nil = any
//   - exit + exit = exit
//   - err + exit = err
//   - err1 + err2 = [err1, err2]
func joinErr(err1, err2 error) error {
	if err1 == nil {
		return err2
	}
	if err2 == nil {
		return err1
	}

	if err1 == ErrExit {
		if err2 != ErrExit {
			return err2
		}
		return ErrExit
	}
	if err2 == ErrExit {
		return err1
	}

	return multierr.Append(err1, err2)
}

// joinResultAndErr joins results by the following rules:
//   - Join the errors with joinErr
//   - If it requires requeue, set the requeue in the global one
//   - If it sets a requeue after, set the requeue after if the global one
//     if there's none or it's longer than the local one
func joinResultAndErr(result ctrl.Result, err error, lresult ctrl.Result, lerr error) (ctrl.Result, error) {
	if lresult.Requeue {
		result.Requeue = true
	}
	if lresult.RequeueAfter > 0 {
		if result.RequeueAfter == 0 || result.RequeueAfter > lresult.RequeueAfter {
			result.RequeueAfter = lresult.RequeueAfter
		}
	}
	return result, joinErr(err, lerr)
}

func runJoinActions(ctx context.Context, actions ...Action) (result ctrl.Result, err error) {
	// Run actions one-by-one and join results.
	for _, act := range actions {
		lr, lerr := act.Run(ctx)
		result, err = joinResultAndErr(result, err, lr, lerr)
	}
	return
}

func runJoinActionsInParallel(ctx context.Context, actions ...Action) (result ctrl.Result, err error) {
	results := make([]ctrl.Result, len(actions))
	errs := make([]error, len(actions))
	panics := make([]any, len(actions))

	// Run each action in a new goroutine and organize with WaitGroup.
	wg := &sync.WaitGroup{}

	for i := range actions {
		act := actions[i]
		resultRef, errRef := &results[i], &errs[i]
		panicRef := &panics[i]
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					*panicRef = r
				}
				wg.Done()
			}()

			*resultRef, *errRef = act.Run(ctx)
		}()
	}

	// Wait should set a memory barrier.
	done := make(chan bool)
	go func() {
		wg.Wait()
		close(done)
	}()
	<-done

	panics = lo.Filter(panics, func(r any, _ int) bool { return r != nil })
	if len(panics) > 0 {
		panic(panics)
	}

	// Join results.
	for i := 0; i < len(actions); i++ {
		result, err = joinResultAndErr(result, err, results[i], errs[i])
	}

	return
}

type joinRunner interface {
	IsParallel() bool
	Run(ctx context.Context, actions ...Action) (ctrl.Result, error)
}

type joinRunFunc func(ctx context.Context, actions ...Action) (ctrl.Result, error)

type parallelJoinRunFunc joinRunFunc

// IsParallel implements joinRunner.
func (r joinRunFunc) IsParallel() bool {
	return false
}

// Run implements joinRunner.
func (r joinRunFunc) Run(ctx context.Context, actions ...Action) (ctrl.Result, error) {
	return r(ctx, actions...)
}

// IsParallel implements joinRunner.
func (r parallelJoinRunFunc) IsParallel() bool {
	return true
}

// Run implements joinRunner.
func (r parallelJoinRunFunc) Run(ctx context.Context, actions ...Action) (ctrl.Result, error) {
	return r(ctx, actions...)
}

var (
	defaultJoinRunner  = joinRunFunc(runJoinActions)
	parallelJoinRunner = parallelJoinRunFunc(runJoinActionsInParallel)
)

var _ internal.Group = &joinGroup{}

type joinGroup struct {
	name    string
	actions []Action
	runner  joinRunner
}

// Children implements the Group.
func (grp *joinGroup) Children() []Action {
	return grp.actions
}

// SetChildren implements the Group.
func (grp *joinGroup) SetChildren(actions []Action) {
	grp.actions = actions
}

// Name implements the Group.
func (grp *joinGroup) Name() string {
	return grp.name
}

// Description implements the Action.
func (grp *joinGroup) Description() string {
	return internal.DescribeGroup(grp.Name(), grp.actions...)
}

// Run implements the Action.
func (grp *joinGroup) Run(ctx context.Context) (ctrl.Result, error) {
	return grp.runner.Run(ctx, grp.actions...)
}

func join(name string, actions []Action, runner joinRunner) Action {
	if len(actions) == 0 {
		return Nop
	}

	if len(actions) == 1 {
		return actions[0]
	}

	return &joinGroup{name: name, actions: actions, runner: runner}
}

// Join organizes the actions in a split-join flow, which doesn't guarantee the execution order.
func Join(actions ...Action) Action {
	return join("Join", lo.Shuffle(actions), defaultJoinRunner)
}

// OrderedJoin organizes the actions in a split-join flow and guarantees the execution order.
func OrderedJoin(actions ...Action) Action {
	return join("Join", actions, defaultJoinRunner)
}

// ParallelJoin organizes the actions in a split-join flow and executes them in parallel.
func ParallelJoin(actions ...Action) Action {
	if len(actions) == 1 {
		return Parallel(actions[0])
	}

	return join("ParallelJoin", actions, parallelJoinRunner)
}
