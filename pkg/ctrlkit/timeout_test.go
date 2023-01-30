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
	"fmt"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Timeout_Decorator(t *testing.T) {
	testDecorator[timeoutAction](t, "Timeout")
}

func newSleepAction(t time.Duration) Action {
	return NewAction("", func(ctx context.Context) (ctrl.Result, error) {
		select {
		case <-ctx.Done():
			return RequeueIfError(ctx.Err())
		case <-time.After(t):
			return Continue()
		}
	})
}

func Test_Timeout_Description(t *testing.T) {
	x := NewAction("A", nothingFunc)
	if Timeout(10*time.Second, x).Description() != fmt.Sprintf("Timeout(%s, timeout=%s)", x.Description(), "10s") {
		t.Fail()
	}
}

func Test_Timeout_Run(t *testing.T) {
	testcases := map[string]struct {
		action      Action
		expectedErr error
	}{
		"timeout-1s-sleep-10s-timeouts": {
			action:      Timeout(time.Second, newSleepAction(10*time.Second)),
			expectedErr: context.DeadlineExceeded,
		},
		"timeout-10s-sleep-1s-passes": {
			action:      Timeout(10*time.Second, newSleepAction(time.Second)),
			expectedErr: nil,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.action.Run(context.Background())
			if err != tc.expectedErr {
				t.Fail()
			}
		})
	}
}
