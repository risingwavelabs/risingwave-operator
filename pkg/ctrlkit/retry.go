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
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

var _ internal.Decorator = &retryAction{}

type retryAction struct {
	inner Action
	limit int
}

func (act *retryAction) Inner() Action {
	return act.inner
}

func (act *retryAction) SetInner(inner Action) {
	act.inner = inner
}

func (act *retryAction) Name() string {
	return "Retry"
}

func (act *retryAction) Description() string {
	return fmt.Sprintf("Retry(%s, limit=%d)", act.inner.Description(), act.limit)
}

func (act *retryAction) Run(ctx context.Context) (result reconcile.Result, err error) {
	for i := 0; i < act.limit; i++ {
		result, err = act.inner.Run(ctx)
		if err == nil || err == ErrExit {
			return
		}
	}
	return
}

// Retry wraps an action into a retryable action.
// It reruns the action iff. there is a non-exit error.
func Retry(limit int, act Action) Action {
	if limit < 1 {
		panic("limit must be positive")
	}

	return &retryAction{
		limit: limit,
		inner: act,
	}
}
