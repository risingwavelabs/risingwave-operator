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
	"sync"

	multierr "github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Join the errors with the following rules:
//   * any + nil = any
//   * exit + exit = exit
//   * err + exit = err
//   * err1 + err2 = [err1, err2]
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
//   * Join the errors with joinErr
//   * If it requires requeue, set the requeue in the global one
//   * If it sets a requeue after, set the requeue after if the global one
//     if there's none or it's longer than the local one
func joinResultAndErr(result ctrl.Result, err error, lresult ctrl.Result, lerr error) (ctrl.Result, error) {
	if lerr != nil {
		err = joinErr(err, lerr)
	}
	if lresult.Requeue {
		result.Requeue = true
	}
	if lresult.RequeueAfter > 0 {
		if result.RequeueAfter == 0 || result.RequeueAfter > lresult.RequeueAfter {
			result.RequeueAfter = lresult.RequeueAfter
		}
	}
	return result, err
}

func runJoinActions(ctx context.Context, actions ...ReconcileAction) (result ctrl.Result, err error) {
	// Run actions one-by-one and join results.
	for _, act := range actions {
		lr, lerr := act.Run(ctx)
		result, err = joinResultAndErr(result, err, lr, lerr)
	}
	return
}

func runJoinActionsInParallel(ctx context.Context, actions ...ReconcileAction) (result ctrl.Result, err error) {
	lresults := make([]ctrl.Result, len(actions))
	lerrs := make([]error, len(actions))

	// Run each action in a new goroutine and organize with WaitGroup.
	wg := &sync.WaitGroup{}

	for i := range actions {
		act := actions[i]
		lresult, lerr := &lresults[i], &lerrs[i]
		wg.Add(1)
		go func() {
			defer wg.Done()

			*lresult, *lerr = act.Run(ctx)
		}()
	}

	// Wait should set a memory barrier.
	wg.Wait()

	// Join results.
	for i := 0; i < len(actions); i++ {
		result, err = joinResultAndErr(result, err, lresults[i], lerrs[i])
	}

	return
}

type joinRunner interface {
	IsParallel() bool
	Run(ctx context.Context, actions ...ReconcileAction) (ctrl.Result, error)
}

type joinRunFunc func(ctx context.Context, actions ...ReconcileAction) (ctrl.Result, error)

type parallelJoinRunFunc joinRunFunc

func (r joinRunFunc) IsParallel() bool {
	return false
}

func (r joinRunFunc) Run(ctx context.Context, actions ...ReconcileAction) (ctrl.Result, error) {
	return r(ctx, actions...)
}

func (r parallelJoinRunFunc) IsParallel() bool {
	return true
}

func (r parallelJoinRunFunc) Run(ctx context.Context, actions ...ReconcileAction) (ctrl.Result, error) {
	return r(ctx, actions...)
}

var (
	defaultJoinRunner  = joinRunFunc(runJoinActions)
	parallelJoinRunner = parallelJoinRunFunc(runJoinActionsInParallel)
)

type joinAction struct {
	actions []ReconcileAction
	runner  joinRunner
}

func (act *joinAction) Description() string {
	if act.runner.IsParallel() {
		return describeGroup("ParallelJoin", act.actions...)
	} else {
		return describeGroup("Join", act.actions...)
	}
}

func (act *joinAction) Run(ctx context.Context) (ctrl.Result, error) {
	return act.runner.Run(ctx, act.actions...)
}

func join(actions []ReconcileAction, runner joinRunner) ReconcileAction {
	if len(actions) == 0 {
		return Nop
	}

	if len(actions) == 1 {
		return actions[0]
	}

	return &joinAction{actions: actions, runner: runner}
}

// Join organizes the actions in a split-join flow, which doesn't guarantee the execution order.
func Join(actions ...ReconcileAction) ReconcileAction {
	return join(lo.Shuffle(actions), defaultJoinRunner)
}

// JoinOrdered organizes the actions in a split-join flow and guarantees the execution order.
func JoinOrdered(actions ...ReconcileAction) ReconcileAction {
	return join(actions, defaultJoinRunner)
}

// JoinInParallel organizes the actions in a split-join flow and executes them in parallel.
func JoinInParallel(actions ...ReconcileAction) ReconcileAction {
	if len(actions) == 1 {
		return Parallel(actions[0])
	}

	return join(actions, parallelJoinRunner)
}
