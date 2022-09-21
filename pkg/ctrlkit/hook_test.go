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

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type pingHook struct {
	ping bool
}

func (h *pingHook) PreRun(ctx context.Context, logger logr.Logger, action string, states map[string]runtime.Object) {
	h.ping = true
}

func (h *pingHook) PostRun(ctx context.Context, logger logr.Logger, action string, result ctrl.Result, err error) {
	if !h.ping {
		panic("ping lost")
	}
}

func TestChainActionHooks(t *testing.T) {
	hook := ChainActionHooks(&pingHook{}, &pingHook{})
	hook.Add(&pingHook{})

	ctx := context.Background()
	logger := logr.Discard()

	defer hook.PostRun(ctx, logger, "", ctrl.Result{}, nil)
	hook.PreRun(ctx, logger, "", nil)
}
