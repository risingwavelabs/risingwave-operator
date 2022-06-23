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

	ctrl "sigs.k8s.io/controller-runtime"
)

type parallelAction struct {
	inner ReconcileAction
}

func (act *parallelAction) Description() string {
	return fmt.Sprintf("Parallel(%s)", act.inner.Description())
}

func (act *parallelAction) Run(ctx context.Context) (result ctrl.Result, err error) {
	done := make(chan bool)
	go func() {
		result, err = act.inner.Run(ctx)
		done <- true
	}()
	<-done

	return
}

// Parallel wraps the action and runs it in parallel.
func Parallel(act ReconcileAction) ReconcileAction {
	switch act := act.(type) {
	case *parallelAction:
		return act
	default:
		return &parallelAction{inner: act}
	}
}
