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

package internal

import "bytes"

// Group describes a groups of actions.
type Group interface {
	Children() []Action
	SetChildren([]Action)
	Name() string
}

// DescribeGroup describes actions grouped, must be a slice with at least 1 elements.
func DescribeGroup(head string, actions ...Action) string {
	buf := &bytes.Buffer{}

	buf.WriteString(head)
	buf.WriteString("(")
	for _, act := range actions[:len(actions)-1] {
		buf.WriteString(act.Description())
		buf.WriteString(", ")
	}
	buf.WriteString(actions[len(actions)-1].Description())
	buf.WriteString(")")

	return buf.String()
}
