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

package version

import (
	"fmt"

	"golang.org/x/mod/semver"
)

const (
	MinimumVersion = "v0.2.0"

	DefaultVersion = "v0.2.0"

	LatestVersion = "latest" // TODO: need to convert latest to the version tag https://github.com/risingwavelabs/risingwave-operator/issues/201
)

func ValidateVersion(version string) (bool, error) {
	if version == LatestVersion {
		return true, nil
	}

	if !semver.IsValid(version) {
		return false, fmt.Errorf("%s is not a valid version", version)
	}

	r := semver.Compare(version, MinimumVersion)
	if r == -1 {
		return false, fmt.Errorf("version must be >= %s", MinimumVersion)
	}
	return true, nil
}
