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

import "testing"

func Test_Join(t *testing.T) {
	if Join(Nop, Nop).Description() != "Join(Nop, Nop)" {
		t.Fatal("description of join is not correct")
	}
}

func Test_JoinInParallel(t *testing.T) {
	if JoinInParallel(Nop).Description() != Parallel(Nop).Description() {
		t.Fatal("one join should be optimized")
	}

	if JoinInParallel(Nop, Nop).Description() != "ParallelJoin(Nop, Nop)" {
		t.Fatal("description of parallel join is not correct")
	}
}
