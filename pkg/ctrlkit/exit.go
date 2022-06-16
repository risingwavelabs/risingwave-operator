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
	"errors"

	ctrl "sigs.k8s.io/controller-runtime"
)

// ErrExit exits the workflow early.
var ErrExit = errors.New("exit")

// Exit returns an empty result and an ErrExit.
func Exit() (ctrl.Result, error) {
	return ctrl.Result{}, ErrExit
}

// IgnoreExit keeps the result but returns a nil when err == ErrExit.
func IgnoreExit(r ctrl.Result, err error) (ctrl.Result, error) {
	// If it's ErrExit, ignore it.
	if err == ErrExit {
		return r, nil
	}

	// Otherwise, it might be nil or a multi error.
	return r, err
}
