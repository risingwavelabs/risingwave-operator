/*
 * Copyright 2024 RisingWave Labs
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

package utils

import (
	"strings"
	"time"
)

// IsNightlyVersionAfter checks if the version is a nightly version and is after the given date.
func IsNightlyVersionAfter(version string, format string, date time.Time) bool {
	if strings.HasPrefix(version, "nightly-") {
		version = strings.TrimPrefix(version, "nightly-")
		t, err := time.Parse(format, version)
		if err != nil {
			return false
		}
		return t.After(date)
	}
	return false
}

// IsRisingWaveNightlyVersionAfter checks if the version is a nightly version and is after the given date.
func IsRisingWaveNightlyVersionAfter(version string, date time.Time) bool {
	return IsNightlyVersionAfter(version, "2006-01-02", date)
}
