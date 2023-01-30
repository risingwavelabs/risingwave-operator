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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

// ActionHook provides hooks for actions implementations in controller manager.
type ActionHook interface {
	// PreRun is a hook that runs before the action. It's embedded by the ctrlkit-gen tool and should
	// be provided with an option. One can get the name and the states of the action from this hook.
	PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object)

	// PostRun is a hook than runs after the action. It's embedded by the ctrlkit-gen tool. One can get
	// the result of the action from this hook.
	PostRun(ctx context.Context, logger logr.Logger, action string, result ctrl.Result, err error)
}

// ChainedActionHooks chains a sequence of hooks into one.
type ChainedActionHooks struct {
	hooks []ActionHook
}

// PreRun implements the ActionHook interface.
func (c *ChainedActionHooks) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {
	// Run hooks one by one.
	for _, h := range c.hooks {
		h.PreRun(ctx, logger, action, states)
	}
}

// PostRun implements the ActionHook interface.
func (c *ChainedActionHooks) PostRun(ctx context.Context, logger logr.Logger, action string, result ctrl.Result, err error) {
	// Run hooks in the reversed order.
	for i := len(c.hooks) - 1; i >= 0; i-- {
		h := c.hooks[i]
		h.PostRun(ctx, logger, action, result, err)
	}
}

// Add adds hook to this.
func (c *ChainedActionHooks) Add(hook ActionHook) {
	c.hooks = append(c.hooks, hook)
}

// ChainActionHooks creates a ChainedActionHooks with the given hooks.
func ChainActionHooks(hooks ...ActionHook) *ChainedActionHooks {
	return &ChainedActionHooks{hooks: hooks}
}
