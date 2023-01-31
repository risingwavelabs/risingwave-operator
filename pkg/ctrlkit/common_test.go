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

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

func zero[T any]() T {
	var t T
	return t
}

var nothingFunc ActionFunc = func(ctx context.Context) (ctrl.Result, error) {
	return NoRequeue()
}

var requeueFunc ActionFunc = func(ctx context.Context) (ctrl.Result, error) {
	return RequeueImmediately()
}

var exitFunc ActionFunc = func(ctx context.Context) (ctrl.Result, error) {
	return Exit()
}

type DecoratorPtr[D any] interface {
	internal.Decorator
	*D
}

func testDecorator[D any, DP DecoratorPtr[D]](t *testing.T, name string) {
	x := NewAction("x", nothingFunc)
	var d D
	dp := DP(&d)
	if dp.Inner() != nil {
		t.Fail()
	}
	dp.SetInner(x)
	if dp.Inner() != x {
		t.Fail()
	}
	if dp.Name() != name {
		t.Fail()
	}
}

func sliceEquals[E any](a, b []E, equals func(x, y E) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !equals(a[i], b[i]) {
			return false
		}
	}
	return true
}

type GroupPtr[G any] interface {
	internal.Group
	*G
}

func testGroup[G any, GP GroupPtr[G]](t *testing.T, name string) {
	x := NewAction("x", nothingFunc)
	y := NewAction("y", nothingFunc)
	actions := []Action{x, y}
	var g G
	gp := GP(&g)
	if gp.Children() != nil {
		t.Fail()
	}
	gp.SetChildren(actions)
	if !sliceEquals(gp.Children(), actions, func(x, y Action) bool { return x == y }) {
		t.Fail()
	}
	if gp.Name() != name {
		t.Fail()
	}
}
