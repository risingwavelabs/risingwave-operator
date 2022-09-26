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
	"fmt"
	"time"

	ctrl "sigs.k8s.io/controller-runtime"
)

// NeedsRequeue reports if the result and error indicates a requeue.
func NeedsRequeue(result ctrl.Result, err error) bool {
	return err != nil || result.Requeue || result.RequeueAfter > 0
}

// RequeueImmediately returns a result with requeue set to true and a nil.
func RequeueImmediately() (ctrl.Result, error) {
	return ctrl.Result{Requeue: true}, nil
}

// RequeueAfter returns a result with requeue after set to the given duration and a nil.
func RequeueAfter(after time.Duration) (ctrl.Result, error) {
	return ctrl.Result{RequeueAfter: after}, nil
}

// RequeueIfError returns an empty result with the err.
func RequeueIfError(err error) (ctrl.Result, error) {
	return ctrl.Result{}, err
}

// RequeueIfErrorAndWrap returns an empty result with a wrapped err.
func RequeueIfErrorAndWrap(explain string, err error) (ctrl.Result, error) {
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("%s: %w", explain, err)
	}
	return ctrl.Result{}, nil
}

// NoRequeue returns an empty result and a nil.
func NoRequeue() (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

// Continue is an alias of NoRequeue.
var Continue = NoRequeue
