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
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/samber/lo"
	"go.uber.org/multierr"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Join_Group(t *testing.T) {
	testGroup[joinGroup](t, "")
}

func Test_Join_Description(t *testing.T) {
	if Join(Nop, Nop).Description() != "Join(Nop, Nop)" {
		t.Fatal("description of join is not correct")
	}
}

func Test_JoinInParallel_Description(t *testing.T) {
	if ParallelJoin(Nop).Description() != Parallel(Nop).Description() {
		t.Fatal("one join should be optimized")
	}

	if ParallelJoin(Nop, Nop).Description() != "ParallelJoin(Nop, Nop)" {
		t.Fatal("description of parallel join is not correct")
	}
}

func Test_joinResultAndErr(t *testing.T) {
	leftErr, rightErr := errors.New("x"), errors.New("y")
	testcases := map[string]struct {
		lr           ctrl.Result
		lerr         error
		rr           ctrl.Result
		rerr         error
		expectResult ctrl.Result
		expectErr    string
		nilErr       bool
	}{
		"err-any-nil": {
			lerr:      leftErr,
			rerr:      nil,
			expectErr: "x",
		},
		"err-nil-any": {
			lerr:      nil,
			rerr:      rightErr,
			expectErr: "y",
		},
		"err-exit-exit": {
			lerr:      ErrExit,
			rerr:      ErrExit,
			expectErr: ErrExit.Error(),
		},
		"err-exit-any": {
			lerr:      ErrExit,
			rerr:      rightErr,
			expectErr: "y",
		},
		"err-any-exit": {
			lerr:      leftErr,
			rerr:      ErrExit,
			expectErr: "x",
		},
		"err-merge-errs": {
			lerr:      leftErr,
			rerr:      rightErr,
			expectErr: multierr.Append(leftErr, rightErr).Error(),
		},
		"requeue-and-not-requeue": {
			lr:           ctrl.Result{Requeue: true},
			expectResult: ctrl.Result{Requeue: true},
			nilErr:       true,
		},
		"not-requeue-and-requeue": {
			rr:           ctrl.Result{Requeue: true},
			expectResult: ctrl.Result{Requeue: true},
			nilErr:       true,
		},
		"requeue-after-and-not-requeue-after": {
			lr:           ctrl.Result{RequeueAfter: time.Second},
			expectResult: ctrl.Result{RequeueAfter: time.Second},
			nilErr:       true,
		},
		"not-requeue-after-and-requeue-after": {
			rr:           ctrl.Result{RequeueAfter: time.Second},
			expectResult: ctrl.Result{RequeueAfter: time.Second},
			nilErr:       true,
		},
		"requeue-after-1s-and-requeue-after-2s": {
			lr:           ctrl.Result{RequeueAfter: time.Second},
			rr:           ctrl.Result{RequeueAfter: 2 * time.Second},
			expectResult: ctrl.Result{RequeueAfter: time.Second},
			nilErr:       true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := joinResultAndErr(tc.lr, tc.lerr, tc.rr, tc.rerr)
			if r != tc.expectResult {
				t.Fail()
			}
			if tc.nilErr && err != nil {
				t.Fail()
			}
			if !tc.nilErr && (err == nil || err.Error() != tc.expectErr) {
				t.Fail()
			}
		})
	}
}

func Test_Join_Simplify(t *testing.T) {
	testcases := map[string]struct {
		actions    []Action
		expectDesc []string
	}{
		"no-actions": {
			actions: nil,
			expectDesc: []string{
				Nop.Description(),
			},
		},
		"single-action": {
			actions: []Action{
				NewAction("x", nothingFunc),
			},
			expectDesc: []string{
				"x",
			},
		},
		"multiple-actions": {
			actions: []Action{
				NewAction("x", nothingFunc),
				NewAction("y", nothingFunc),
			},
			expectDesc: []string{
				"Join(x, y)",
				"Join(y, x)",
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			desc := Join(tc.actions...).Description()
			if !lo.Contains(tc.expectDesc, desc) {
				t.Fail()
			}
		})
	}
}

func Test_JoinInParallel_Simplify(t *testing.T) {
	x := NewAction("X", nothingFunc)
	if ParallelJoin(x).Description() != "Parallel(X)" {
		t.Fail()
	}
}

func newCounterActions(size int, cnt *int) []Action {
	actions := make([]Action, 0, size)
	for i := 0; i < size; i++ {
		actions = append(actions, NewAction(fmt.Sprintf("C_%d", i), func(ctx context.Context) (ctrl.Result, error) {
			*cnt++
			return NoRequeue()
		}))
	}
	return actions
}

func Test_Join_Run(t *testing.T) {
	testcases := map[string]struct {
		actions []Action
		count   int
		result  ctrl.Result
		err     string
	}{
		"count-10-with-exit": {
			actions: []Action{
				NewAction("exit", exitFunc),
			},
			count:  10,
			result: ctrl.Result{},
			err:    ErrExit.Error(),
		},
		"count-10-with-requeue-and-exit": {
			actions: []Action{
				NewAction("requeue", requeueFunc),
				NewAction("exit", exitFunc),
			},
			count:  10,
			result: ctrl.Result{Requeue: true},
			err:    ErrExit.Error(),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cnt := 0
			actions := append(newCounterActions(tc.count, &cnt), tc.actions...)

			r, err := Join(actions...).Run(context.Background())
			if r != tc.result {
				t.Fail()
			}
			if err != nil && err.Error() != tc.err {
				t.Fail()
			}
			if err == nil && len(tc.err) != 0 {
				t.Fail()
			}
			if cnt != tc.count {
				t.Fail()
			}
		})
	}
}

func Test_JoinInParallel_Run(t *testing.T) {
	pingChan := make(chan int, 1)
	pongChan := make(chan int, 1)
	ping := NewAction("ping", func(ctx context.Context) (ctrl.Result, error) {
		pingChan <- 1
		<-pongChan
		return NoRequeue()
	})
	pong := NewAction("pong", func(ctx context.Context) (ctrl.Result, error) {
		pongChan <- 1
		<-pingChan
		return NoRequeue()
	})

	_, err := ParallelJoin(ping, pong).Run(context.Background())
	if err == context.DeadlineExceeded {
		t.Fail()
	}
}

func Test_JoinInOrder_Run(t *testing.T) {
	// Should work only in order. Otherwise the count actions would panic.
	cnt := 0
	OrderedJoin(newSequentialCountActs(10, &cnt)...).Run(context.Background())
}
