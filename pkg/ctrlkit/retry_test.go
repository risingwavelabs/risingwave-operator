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
	"errors"
	"strconv"
	"testing"
	"time"

	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Retry_Decorator(t *testing.T) {
	testDecorator[retryAction](t, "Retry")
}

func newFailUntilAction(failCnt int, cnt *int) Action {
	return NewAction("FailCnt-"+strconv.Itoa(failCnt), func(ctx context.Context) (ctrl.Result, error) {
		if *cnt < failCnt {
			*cnt++
			return RequeueIfError(errors.New("fail"))
		}
		*cnt++
		return Continue()
	})
}

func newExitCountAction(cnt *int) Action {
	return NewAction("ExitCount", func(ctx context.Context) (ctrl.Result, error) {
		*cnt++
		return Exit()
	})
}

func Test_Retry_Description(t *testing.T) {
	x := NewAction("A", nothingFunc)
	if Retry(3, x).Description() != "Retry(A, limit=3)" {
		t.Fail()
	}

	if RetryInterval(3, 10*time.Millisecond, x).Description() != "Retry(A, limit=3, interval=10ms)" {
		t.Fail()
	}
}

func Test_Retry_LimitShouldGreaterThanZero(t *testing.T) {
	testcases := map[string]struct {
		shouldPanic bool
		limit       int
	}{
		"minus-one-panics": {
			shouldPanic: true,
			limit:       -1,
		},
		"zero-panics": {
			shouldPanic: true,
			limit:       0,
		},
		"one-not-panics": {
			shouldPanic: false,
			limit:       1,
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

			Retry(tc.limit, Nop)
		})
	}
}

func Test_RetryInterval_ShouldPanic(t *testing.T) {
	testcases := map[string]struct {
		shouldPanic bool
		limit       int
		interval    time.Duration
	}{
		"minus-one-panics": {
			shouldPanic: true,
			limit:       -1,
		},
		"zero-panics": {
			shouldPanic: true,
			limit:       0,
		},
		"negative-duration-panics": {
			shouldPanic: true,
			limit:       1,
			interval:    -1 * time.Second,
		},
		"zero-duration-panics": {
			shouldPanic: true,
			limit:       1,
			interval:    0,
		},
		"positive-duration-not-panics": {
			shouldPanic: false,
			limit:       1,
			interval:    time.Second,
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

			RetryInterval(tc.limit, tc.interval, Nop)
		})
	}
}

func Test_Retry(t *testing.T) {
	cnt := 0

	testcases := map[string]struct {
		action      Action
		expectedCnt int
		expectedErr string
	}{
		"fail-cnt-3-retry-2-failed": {
			action:      Retry(2, newFailUntilAction(3, &cnt)),
			expectedCnt: 2,
			expectedErr: "fail",
		},
		"fail-cnt-3-retry-3-failed": {
			action:      Retry(3, newFailUntilAction(3, &cnt)),
			expectedCnt: 3,
			expectedErr: "fail",
		},
		"fail-cnt-4-retry-3-succeeded": {
			action:      Retry(4, newFailUntilAction(3, &cnt)),
			expectedCnt: 4,
			expectedErr: "",
		},
		"exit-will-not-retry": {
			action:      Retry(3, newExitCountAction(&cnt)),
			expectedCnt: 1,
			expectedErr: ErrExit.Error(),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cnt = 0
			_, err := tc.action.Run(context.Background())
			if err != nil && err.Error() != tc.expectedErr {
				t.Fail()
			}
			if err == nil && tc.expectedErr != "" {
				t.Fail()
			}
			if tc.expectedCnt != cnt {
				t.Fail()
			}
		})
	}
}

func Test_RetryInterval(t *testing.T) {
	testcases := map[string]struct {
		action      Action
		limit       int
		interval    time.Duration
		timeout     time.Duration
		expectedErr string
	}{
		"fail": {
			action:      newFailUntilAction(3, pointer.Int(0)),
			limit:       4,
			interval:    10 * time.Millisecond,
			expectedErr: "",
		},
		"timeout": {
			action:      newFailUntilAction(3, pointer.Int(0)),
			limit:       10,
			interval:    10 * time.Second,
			timeout:     10 * time.Millisecond,
			expectedErr: context.DeadlineExceeded.Error(),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			act := RetryInterval(tc.limit, tc.interval, tc.action)
			ctx := context.Background()
			if tc.timeout > 0 {
				tCtx, cancel := context.WithTimeout(ctx, tc.timeout)
				defer cancel()
				ctx = tCtx
			}
			_, err := act.Run(ctx)
			if err != nil && err.Error() != tc.expectedErr {
				t.Fail()
			}
			if err == nil && tc.expectedErr != "" {
				t.Fail()
			}
		})
	}
}
