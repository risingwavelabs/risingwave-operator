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
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Action(t *testing.T) {
	testcases := map[string]struct {
		desc        string
		f           ActionFunc
		result      ctrl.Result
		err         error
		shouldPanic bool
	}{
		"nil-func-panics": {
			shouldPanic: true,
		},
		"desc-equals": {
			desc: "some desc",
			f: func(ctx context.Context) (ctrl.Result, error) {
				return NoRequeue()
			},
		},
		"result-equals": {
			f: func(ctx context.Context) (ctrl.Result, error) {
				return RequeueImmediately()
			},
			result: ctrl.Result{Requeue: true},
			err:    nil,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (!tc.shouldPanic && r != nil) || (tc.shouldPanic && r == nil) {
					t.Fail()
				}
			}()
			act := NewAction(tc.desc, tc.f)
			if act.Description() != tc.desc {
				t.Fail()
			}
			r, err := act.Run(context.Background())
			if r != tc.result || err != tc.err {
				t.Fail()
			}
		})
	}
}
