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

// If returns the given action if predicate is true, or an Nop otherwise.
func If(predicate bool, act Action) Action {
	if predicate {
		return act
	}
	return Nop
}

// IfElse returns the action left if predicate is true, or the right otherwise.
func IfElse(predicate bool, left, right Action) Action {
	if predicate {
		return left
	}
	return right
}
