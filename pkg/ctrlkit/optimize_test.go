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
	"testing"
	"time"

	"github.com/samber/lo"
)

func Test_UnwrapParallel(t *testing.T) {
	x := NewAction("X", nothingFunc)
	testcases := map[string]struct {
		action Action
		desc   string
	}{
		"no-parallel": {
			action: x,
			desc:   x.Description(),
		},
		"1-parallel": {
			action: Parallel(x),
			desc:   x.Description(),
		},
		"2-parallels": {
			action: Parallel(Parallel(x)),
			desc:   x.Description(),
		},
		"2-parallels-and-timeout-between": {
			action: Parallel(Timeout(5*time.Second, Parallel(x))),
			desc:   Timeout(5*time.Second, Parallel(x)).Description(),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if unwrapParallel(tc.action).Description() != tc.desc {
				t.Fail()
			}
		})
	}
}

func Test_OptimizeWorkflow(t *testing.T) {
	x := NewAction("x", nothingFunc)
	y := NewAction("y", nothingFunc)
	testcases := map[string]struct {
		action Action
		desc   []string
	}{
		"remove-nop-in-sequential": {
			action: Sequential(x, Nop, y),
			desc: []string{
				Sequential(x, y).Description(),
			},
		},
		"remove-nop-in-join": {
			action: Join(x, Nop, y),
			desc: []string{
				"Join(x, y)",
				"Join(y, x)",
			},
		},
		"remove-parallel-in-sequential": {
			action: Sequential(x, Parallel(y)),
			desc: []string{
				Sequential(x, y).Description(),
			},
		},
		"empty-join-to-nop": {
			action: Join(Nop, Nop),
			desc:   []string{Nop.Description()},
		},
		"empty-sequential-to-nop": {
			action: Sequential(Nop, Nop),
			desc:   []string{Nop.Description()},
		},
		"single-join-unwrap": {
			action: Join(x),
			desc:   []string{x.Description()},
		},
		"single-sequential-unwrap": {
			action: Sequential(y),
			desc:   []string{y.Description()},
		},
		"flatten-sequential": {
			action: Sequential(x, Sequential(x, y)),
			desc: []string{
				Sequential(x, x, y).Description(),
			},
		},
		"flatten-join": {
			action: Join(x, Join(x, y)),
			desc: []string{
				"Join(x, x, y)",
				"Join(y, x, x)",
				"Join(x, y, x)",
			},
		},
		"unwrap-parallel-sequence": {
			action: Parallel(Parallel(x)),
			desc: []string{
				Parallel(x).Description(),
			},
		},
		"timeout-big-timeout-small": {
			action: Timeout(5*time.Second, Timeout(time.Second, x)),
			desc: []string{
				Timeout(time.Second, x).Description(),
			},
		},
		"timeout-small-timeout-big": {
			action: Timeout(time.Second, Timeout(5*time.Second, x)),
			desc: []string{
				Timeout(time.Second, x).Description(),
			},
		},
		"join-in-parallel-with-nop": {
			action: JoinInParallel(x, Nop),
			desc: []string{
				Parallel(x).Description(),
			},
		},
		"parallel-nop": {
			action: Parallel(Nop),
			desc: []string{
				Nop.Description(),
			},
		},
		"timeout-nop": {
			action: Timeout(time.Second, Nop),
			desc: []string{
				Nop.Description(),
			},
		},
		"retry-nop": {
			action: Retry(2, Nop),
			desc: []string{
				Nop.Description(),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			desc := OptimizeWorkflow(tc.action).Description()
			if !lo.Contains(tc.desc, desc) {
				t.Fail()
			}
		})
	}
}
