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

package util

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	IsTerminal = func() bool {
		return !color.NoColor
	}

	Bold = func() func(a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.Bold).SprintFunc()
		}
		return fmt.Sprint
	}()

	Red = func() func(format string, a ...interface{}) string {
		if IsTerminal() {
			return color.New(color.FgRed).SprintfFunc()
		}
		return fmt.Sprintf
	}()
)
