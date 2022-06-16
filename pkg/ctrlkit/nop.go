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

	ctrl "sigs.k8s.io/controller-runtime"
)

type nopAction struct{}

func (act *nopAction) Description() string {
	return "Nop"
}

func (act *nopAction) Run(ctx context.Context) (ctrl.Result, error) {
	return NoRequeue()
}

// Nop is a special action that does nothing.
var Nop = &nopAction{}
