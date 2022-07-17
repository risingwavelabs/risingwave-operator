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
	"errors"
	"strings"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_NeedsRequeue(t *testing.T) {
	testcases := map[string]struct {
		result  ctrl.Result
		err     error
		requeue bool
	}{
		"zero-result-nil-err": {
			requeue: false,
		},
		"zero-result-non-nil-err": {
			err:     errors.New(""),
			requeue: true,
		},
		"requeue-result-nil-err": {
			result:  ctrl.Result{Requeue: true},
			requeue: true,
		},
		"requeue-after-result-nil-err": {
			result:  ctrl.Result{RequeueAfter: time.Second},
			requeue: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if NeedsRequeue(tc.result, tc.err) != tc.requeue {
				t.Fail()
			}
		})
	}
}

func Test_RequeueImmediately(t *testing.T) {
	r, err := RequeueImmediately()
	expect := ctrl.Result{Requeue: true}
	if r != expect || err != nil {
		t.Fail()
	}
}

func Test_NoRequeue(t *testing.T) {
	r, err := NoRequeue()
	if r != zero[ctrl.Result]() || err != nil {
		t.Fail()
	}
}

func Test_RequeueAfter(t *testing.T) {
	testcases := map[string]struct {
		after  time.Duration
		result ctrl.Result
		err    error
	}{
		"1-second": {
			after:  time.Second,
			result: ctrl.Result{RequeueAfter: time.Second},
		},
		"10-seconds": {
			after:  10 * time.Second,
			result: ctrl.Result{RequeueAfter: 10 * time.Second},
		},
		"1-hour": {
			after:  time.Hour,
			result: ctrl.Result{RequeueAfter: time.Hour},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := RequeueAfter(tc.after)
			if r != tc.result || err != tc.err {
				t.Fail()
			}
		})
	}
}

func Test_RequeueIfError(t *testing.T) {
	someErr := errors.New("")

	testcases := map[string]struct {
		err       error
		result    ctrl.Result
		expectErr error
	}{
		"nil-err": {
			err:       nil,
			expectErr: nil,
		},
		"some-err": {
			err:       someErr,
			expectErr: someErr,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := RequeueIfError(tc.err)
			if r != tc.result || err != tc.expectErr {
				t.Fail()
			}
		})
	}
}

func Test_RequeueIfErrorAndWrap(t *testing.T) {
	someErr := errors.New("")

	testcases := map[string]struct {
		err         error
		message     string
		result      ctrl.Result
		expectIsErr error
	}{
		"nil-err": {
			err:         nil,
			expectIsErr: nil,
		},
		"some-err": {
			err:         someErr,
			message:     "some message",
			expectIsErr: someErr,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r, err := RequeueIfErrorAndWrap(tc.message, tc.err)
			if r != tc.result {
				t.Fail()
			}
			if tc.expectIsErr == nil {
				if err != nil {
					t.Fail()
				}
			} else {
				if !errors.Is(err, tc.expectIsErr) {
					t.Fail()
				}
				if !strings.HasPrefix(err.Error(), tc.message) {
					t.Fail()
				}
			}
		})
	}
}
