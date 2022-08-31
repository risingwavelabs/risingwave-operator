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
	"time"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

var _ internal.Decorator = &timeoutAction{}

type timeoutAction struct {
	timeout time.Duration
	inner   Action
}

func (act *timeoutAction) Inner() Action {
	return act.inner
}

func (act *timeoutAction) SetInner(inner Action) {
	act.inner = inner
}

func (act *timeoutAction) Name() string {
	return "Timeout"
}

func (act *timeoutAction) Description() string {
	return fmt.Sprintf("Timeout(%s, timeout=%s)", act.inner.Description(), act.timeout)
}

func (act *timeoutAction) Run(ctx context.Context) (ctrl.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, act.timeout)
	defer cancel()

	return act.inner.Run(ctx)
}

// Timeout wraps the reconcile action with a timeout.
func Timeout(timeout time.Duration, act Action) Action {
	return &timeoutAction{timeout: timeout, inner: act}
}
