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
	"errors"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Exit(t *testing.T) {
	r, err := Exit()

	if err != ErrExit || r != zero[ctrl.Result]() {
		t.Fail()
	}
}

func Test_ExitIf(t *testing.T) {
	testcases := map[string]struct {
		cond   bool
		result ctrl.Result
		err    error
	}{
		"true": {
			cond:   true,
			result: zero[ctrl.Result](),
			err:    ErrExit,
		},
		"false": {
			cond:   false,
			result: zero[ctrl.Result](),
			err:    nil,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := ExitIf(tc.cond)
			if r != tc.result || err != tc.err {
				t.Fail()
			}
		})
	}
}

func Test_IgnoreExit(t *testing.T) {
	nonExitErr := errors.New("")

	testcases := map[string]struct {
		result       ctrl.Result
		err          error
		expectResult ctrl.Result
		expectErr    error
	}{
		"exit-err": {
			result:       zero[ctrl.Result](),
			err:          ErrExit,
			expectResult: zero[ctrl.Result](),
			expectErr:    nil,
		},
		"non-exit-err": {
			result:       zero[ctrl.Result](),
			err:          nonExitErr,
			expectResult: zero[ctrl.Result](),
			expectErr:    nonExitErr,
		},
		"exit-err-requeue": {
			result:       ctrl.Result{Requeue: true},
			err:          ErrExit,
			expectResult: ctrl.Result{Requeue: true},
			expectErr:    nil,
		},
		"exit-err-requeue-after": {
			result:       ctrl.Result{RequeueAfter: time.Second},
			err:          ErrExit,
			expectResult: ctrl.Result{RequeueAfter: time.Second},
			expectErr:    nil,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := IgnoreExit(tc.result, tc.err)
			if r != tc.expectResult || err != tc.expectErr {
				t.Fail()
			}
		})
	}
}
