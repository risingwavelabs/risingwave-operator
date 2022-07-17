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
		beforeOpt  Action
		expectDesc []string
	}{
		"remove-nop-in-sequential-2": {
			beforeOpt: Sequential(Nop, y),
			expectDesc: []string{
				y.Description(),
			},
		},
		"remove-nop-in-sequential-3": {
			beforeOpt: Sequential(x, Nop, y),
			expectDesc: []string{
				Sequential(x, y).Description(),
			},
		},
		"remove-nop-in-join-2": {
			beforeOpt: Join(x, Nop),
			expectDesc: []string{
				x.Description(),
			},
		},
		"remove-nop-in-join-3": {
			beforeOpt: Join(x, Nop, y),
			expectDesc: []string{
				"Join(x, y)",
				"Join(y, x)",
			},
		},
		"remove-parallel-in-sequential": {
			beforeOpt: Sequential(x, Parallel(y)),
			expectDesc: []string{
				Sequential(x, y).Description(),
			},
		},
		"empty-join-to-nop": {
			beforeOpt:  Join(Nop, Nop),
			expectDesc: []string{Nop.Description()},
		},
		"empty-sequential-to-nop": {
			beforeOpt:  Sequential(Nop, Nop),
			expectDesc: []string{Nop.Description()},
		},
		"single-join-unwrap": {
			beforeOpt:  Join(x),
			expectDesc: []string{x.Description()},
		},
		"single-sequential-unwrap": {
			beforeOpt:  Sequential(y),
			expectDesc: []string{y.Description()},
		},
		"flatten-sequential": {
			beforeOpt: Sequential(x, Sequential(x, y)),
			expectDesc: []string{
				Sequential(x, x, y).Description(),
			},
		},
		"flatten-nested-join": {
			beforeOpt: Join(x, Join(x, y)),
			expectDesc: []string{
				"Join(x, x, y)",
				"Join(y, x, x)",
				"Join(x, y, x)",
			},
		},
		"not-flatten-different-nested-joins-1": {
			beforeOpt: Join(x, ParallelJoin(x, y)),
			expectDesc: []string{
				"Join(x, ParallelJoin(x, y))",
				"Join(ParallelJoin(x, y), x)",
			},
		},
		"not-flatten-different-nested-joins-2": {
			beforeOpt: ParallelJoin(x, Join(x, y)),
			expectDesc: []string{
				"ParallelJoin(x, Join(x, y))",
				"ParallelJoin(x, Join(y, x))",
			},
		},
		"unwrap-parallel-sequence": {
			beforeOpt: Parallel(Parallel(x)),
			expectDesc: []string{
				Parallel(x).Description(),
			},
		},
		"timeout-big-timeout-small": {
			beforeOpt: Timeout(5*time.Second, Timeout(time.Second, x)),
			expectDesc: []string{
				Timeout(time.Second, x).Description(),
			},
		},
		"timeout-small-timeout-big": {
			beforeOpt: Timeout(time.Second, Timeout(5*time.Second, x)),
			expectDesc: []string{
				Timeout(time.Second, x).Description(),
			},
		},
		"join-in-parallel-with-nop": {
			beforeOpt: ParallelJoin(x, Nop),
			expectDesc: []string{
				Parallel(x).Description(),
			},
		},
		"parallel-nop": {
			beforeOpt: Parallel(Nop),
			expectDesc: []string{
				Nop.Description(),
			},
		},
		"timeout-nop": {
			beforeOpt: Timeout(time.Second, Nop),
			expectDesc: []string{
				Nop.Description(),
			},
		},
		"retry-nop": {
			beforeOpt: Retry(2, Nop),
			expectDesc: []string{
				Nop.Description(),
			},
		},
		"retry-some": {
			beforeOpt: Retry(2, x),
			expectDesc: []string{
				Retry(2, x).Description(),
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			desc := OptimizeWorkflow(tc.beforeOpt).Description()
			if !lo.Contains(tc.expectDesc, desc) {
				t.Fail()
			}
		})
	}
}
