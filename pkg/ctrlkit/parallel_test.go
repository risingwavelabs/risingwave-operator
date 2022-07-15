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
	"testing"

	ctrl "sigs.k8s.io/controller-runtime"
)

func Test_Parallel_Description(t *testing.T) {
	x := NewAction("A", nothingFunc)
	if Parallel(x).Description() != "Parallel(A)" {
		t.Fail()
	}
}

func Test_Parallel_Simplify(t *testing.T) {
	testcases := map[string]struct {
		inner Action
		same  bool
	}{
		"simply": {
			inner: Parallel(Nop),
			same:  true,
		},
		"not-simply-nop": {
			inner: Nop,
			same:  false,
		},
		"not-simply": {
			inner: NewAction("any", nothingFunc),
			same:  false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			act := Parallel(tc.inner)
			if tc.same && act != tc.inner {
				t.Fail()
			}
		})
	}
}

func Test_Parallel_Run(t *testing.T) {
	blockChan := make(chan int)
	x := NewAction("block chan", func(ctx context.Context) (ctrl.Result, error) {
		select {
		case <-blockChan:
			return NoRequeue()
		case <-ctx.Done():
			t.Fail()
			return Exit()
		}
	})

	// Start a sender.
	go func() {
		blockChan <- 1
	}()

	Parallel(x).Run(context.Background())
}
