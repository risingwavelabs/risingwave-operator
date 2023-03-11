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
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

var _ internal.Decorator = &retryAction{}

type retryAction struct {
	inner    Action
	limit    int
	interval time.Duration
}

// Inner implements the Decorator.
func (act *retryAction) Inner() Action {
	return act.inner
}

// SetInner implements the Decorator.
func (act *retryAction) SetInner(inner Action) {
	act.inner = inner
}

// Name implements the Decorator.
func (act *retryAction) Name() string {
	return "Retry"
}

// Description implements the Action.
func (act *retryAction) Description() string {
	if act.interval == 0 {
		return fmt.Sprintf("Retry(%s, limit=%d)", act.inner.Description(), act.limit)
	}
	return fmt.Sprintf("Retry(%s, limit=%d, interval=%s)", act.inner.Description(), act.limit, act.interval.String())
}

// Run implements the Action.
func (act *retryAction) Run(ctx context.Context) (result reconcile.Result, err error) {
	for i := 0; i < act.limit; i++ {
		result, err = act.inner.Run(ctx)
		if err == nil || err == ErrExit {
			return
		}

		if act.interval > 0 && i < act.limit-1 {
			select {
			case <-ctx.Done():
				return RequeueIfError(ctx.Err())
			case <-time.After(act.interval):
				continue
			}
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

// RetryInterval wraps an action into a retryable action. It accepts a retry interval to make gaps between retires.
func RetryInterval(limit int, interval time.Duration, act Action) Action {
	if limit < 1 {
		panic("limit must be positive")
	}
	if interval <= 0 {
		panic("interval must be positive")
	}

	return &retryAction{
		limit:    limit,
		inner:    act,
		interval: interval,
	}
}
