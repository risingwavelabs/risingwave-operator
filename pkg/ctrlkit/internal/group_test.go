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

package internal

import (
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type action struct {
	desc string
}

func (a *action) Description() string {
	return a.desc
}

func (a *action) Run(ctx context.Context) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func Test_DescribeGroup(t *testing.T) {
	testcases := map[string]struct {
		head       string
		actions    []Action
		expectDesc string
	}{
		"single": {
			head: "G",
			actions: []Action{
				&action{desc: "a"},
			},
			expectDesc: "G(a)",
		},
		"multiple": {
			head: "G",
			actions: []Action{
				&action{desc: "a"},
				&action{desc: "b"},
			},
			expectDesc: "G(a, b)",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			if DescribeGroup(tc.head, tc.actions...) != tc.expectDesc {
				t.Fail()
			}
		})
	}
}
