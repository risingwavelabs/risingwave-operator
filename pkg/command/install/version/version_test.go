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

package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateVersion(t *testing.T) {
	var version = "1.0.1"
	ok, err := ValidateVersion(version)
	assert.Equal(t, ok, false)
	assert.Equal(t, strings.Contains(err.Error(), "is not a valid version"), true)

	version = "v2.0.1-beta"
	ok, err = ValidateVersion(version)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)

	version = "v0.1.1-beta"
	ok, err = ValidateVersion(version)
	assert.Equal(t, ok, false)
	assert.Equal(t, strings.Contains(err.Error(), "version must be >="), true)

	version = "v0.2.0"
	ok, err = ValidateVersion(version)
	assert.Equal(t, ok, true)
	assert.Equal(t, err, nil)

	version = "v0.2.0-alpha"
	ok, err = ValidateVersion(version)
	assert.Equal(t, ok, false)
	assert.Equal(t, strings.Contains(err.Error(), "version must be >="), true)

	version = "latest"
	ok, _ = ValidateVersion(version)
	assert.Equal(t, ok, true)
}
