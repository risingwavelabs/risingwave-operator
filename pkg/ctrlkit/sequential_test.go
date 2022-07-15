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
	"fmt"
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Sequential_Description(t *testing.T) {
	x, y := NewAction("X", nothingFunc), NewAction("Y", nothingFunc)
	if Sequential(x, y).Description() != "Sequential(X, Y)" {
		t.Fail()
	}
}

func Test_Sequential_Simplify(t *testing.T) {
	testcases := map[string]struct {
		actions    []Action
		expectDesc string
	}{
		"no-actions": {
			actions:    nil,
			expectDesc: Nop.Description(),
		},
		"one-action": {
			actions: []Action{
				NewAction("X", nothingFunc),
			},
			expectDesc: "X",
		},
		"multi-actions": {
			actions: []Action{
				NewAction("X", nothingFunc),
				NewAction("Y", nothingFunc),
			},
			expectDesc: "Sequential(X, Y)",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if Sequential(tc.actions...).Description() != tc.expectDesc {
				t.Fail()
			}
		})
	}
}

func newSequentialCounter(idx int, cnt *int) func() {
	return func() {
		if *cnt == idx {
			*cnt++
		} else {
			panic("never reach here")
		}
	}
}

func newSequentialCountActs(size int, cnt *int) []Action {
	acts := make([]Action, 0, size)
	for i := 0; i < size; i++ {
		counter := newSequentialCounter(i, cnt)
		acts = append(acts, NewAction(fmt.Sprintf("C_%d", i), func(ctx context.Context) (ctrl.Result, error) {
			counter()
			return NoRequeue()
		}))
	}
	return acts
}

func Test_Sequential_Run(t *testing.T) {
	testcases := map[string]struct {
		countSize  int
		exitBefore int // Negative means no exit
		expectCnt  int
	}{
		"count-3-no-exit": {
			countSize:  3,
			exitBefore: -1,
			expectCnt:  3,
		},
		"count-3-exit-at-start": {
			countSize:  3,
			exitBefore: 0,
			expectCnt:  0,
		},
		"count-3-exit-before-1": {
			countSize:  3,
			exitBefore: 1,
			expectCnt:  1,
		},
		"count-5-exit-before-3": {
			countSize:  5,
			exitBefore: 3,
			expectCnt:  3,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cnt := 0
			var sequentialAct Action
			counters := newSequentialCountActs(tc.countSize, &cnt)
			if tc.exitBefore >= 0 {
				acts := append(counters[:tc.exitBefore], NewAction("exit", exitFunc))
				acts = append(acts, counters[tc.exitBefore:]...)
				sequentialAct = Sequential(acts...)
			} else {
				sequentialAct = Sequential(counters...)
			}

			_, err := sequentialAct.Run(context.Background())
			if (tc.exitBefore < 0 && err != nil) || (tc.exitBefore >= 0 && err != ErrExit) {
				t.Fail()
			}

			if cnt != tc.expectCnt {
				t.Fail()
			}
		})
	}
}
