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
	"sync"

	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/risingwavelabs/risingwave-operator/pkg/ctrlkit/internal"
)

var _ internal.Decorator = &sharedAction{}

type sharedAction struct {
	inner Action

	once   sync.Once
	done   chan bool
	result ctrl.Result
	err    error
	panic  any
}

// Inner implements the Decorator.
func (s *sharedAction) Inner() internal.Action {
	return s.inner
}

// SetInner implements the Decorator.
func (s *sharedAction) SetInner(inner internal.Action) {
	s.inner = inner
}

// Name implements the Decorator.
func (s *sharedAction) Name() string {
	return "Shared"
}

// Description implements the Action.
func (s *sharedAction) Description() string {
	return fmt.Sprintf("%s(%s)", s.Name(), s.inner.Description())
}

// Run implements the Action.
func (s *sharedAction) Run(ctx context.Context) (ctrl.Result, error) {
	// Start a new goroutine to do this.
	go s.once.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				s.panic = r
			}
			close(s.done)
		}()
		s.result, s.err = s.inner.Run(ctx)
	})

	// Wait on the done channel. Memory barrier should also be carried here.
	<-s.done

	if s.panic != nil {
		panic(s.panic)
	} else {
		return s.result, s.err
	}
}

// Shared wraps the action into a shared action. Any executions against this action
// would result in exactly once execution of the inner action and the same result.
func Shared(inner Action) Action {
	if _, ok := inner.(*sharedAction); ok {
		return inner
	}
	return &sharedAction{
		inner: inner,
		done:  make(chan bool),
	}
}
