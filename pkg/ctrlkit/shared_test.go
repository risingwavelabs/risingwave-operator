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
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Shared_Decorator(t *testing.T) {
	testDecorator[sharedAction](t, "Shared")
}

func newAtomicCounter(cnt *int32, resultErr resultErr) Action {
	return NewAction("AtomicCounter", func(ctx context.Context) (ctrl.Result, error) {
		atomic.AddInt32(cnt, 1)
		return resultErr.result, resultErr.err
	})
}

func Test_Shared_Optimize(t *testing.T) {
	if Shared(Shared(Nop)).Description() != Shared(Nop).Description() {
		t.Fail()
	}
}

func Test_Shared(t *testing.T) {
	testcases := map[string]struct {
		resultErr resultErr
		extraRun  int
	}{
		"extra-run-0": {
			extraRun: 0,
		},
		"extra-run-1": {
			extraRun: 1,
		},
		"extra-run-3": {
			resultErr: resultErr{
				result: ctrl.Result{RequeueAfter: time.Second},
				err:    errors.New("some error"),
			},
			extraRun: 3,
		},
		"extra-run-10": {
			resultErr: resultErr{
				result: ctrl.Result{Requeue: true},
				err:    nil,
			},
			extraRun: 10,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			cnt := int32(0)
			shared := Shared(newAtomicCounter(&cnt, tc.resultErr))
			var r ctrl.Result
			var err error
			firstRunDone := make(chan bool)
			fatal := false
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				r, err = shared.Run(context.Background())
				close(firstRunDone)
			}()
			for i := 0; i < tc.extraRun; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					extraR, extraErr := shared.Run(context.Background())

					// Wait for the first result and compare.
					<-firstRunDone
					if r != extraR || err != extraErr {
						fatal = true
					}
				}()
			}
			wg.Wait()

			if fatal {
				t.Fatal("result not the same")
			}

			if cnt != 1 {
				t.Fatal("run count is not 1")
			}
		})
	}
}
