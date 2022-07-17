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
