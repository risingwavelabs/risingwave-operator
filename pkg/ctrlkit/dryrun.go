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

import "fmt"

// DryRun reports actions with a structured output without run them.
func DryRun(act ReconcileAction) {
	dryRun(act, "")
}

func dryRun(act ReconcileAction, indent string) {
	switch act := act.(type) {
	case *sequentialAction:
		fmt.Println("Sequential(")

		for _, act := range act.actions {
			fmt.Print(indent + "  ")
			dryRun(act, indent+"  ")
		}

		fmt.Print(indent)
		fmt.Println(")")
	case *joinAction:
		fmt.Println("Join(")

		for _, act := range act.actions {
			fmt.Print(indent + "  ")
			dryRun(act, indent+"  ")
		}

		fmt.Print(indent)
		fmt.Println(")")
	case *retryAction:
		fmt.Printf("retry=%d  ", act.limit)
		dryRun(act.inner, indent)
	case *parallelAction:
		fmt.Printf("parallel  ")
		dryRun(act.inner, indent)
	case *timeoutAction:
		fmt.Printf("timeout=%s  ", act.timeout)
		dryRun(act.inner, indent)
	default:
		fmt.Println(act.Description())
	}
}
