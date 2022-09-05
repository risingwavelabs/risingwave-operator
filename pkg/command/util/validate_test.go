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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

var groups = []v1alpha1.RisingWaveComputeGroup{
	{
		Name: "test-g-1",
	},
	{
		Name: "test-g-2",
	},
}

func Test_ValidComputeGroup(t *testing.T) {
	var groupName = "test-g"
	r := IsValidComputeGroup(groupName, groups)
	assert.Equal(t, r, false)

	groupName = "test-g-1"
	r = IsValidComputeGroup(groupName, groups)
	assert.Equal(t, r, true)
}

func Test_ValidRWGroup(t *testing.T) {
	var groups = []v1alpha1.RisingWaveComponentGroup{
		{
			Name: "test-g-1",
		},
		{
			Name: "test-g-2",
		},
	}
	var groupName = "test-g"
	r := IsValidRWGroup(groupName, groups)
	assert.Equal(t, r, false)

	groupName = "test-g-1"
	r = IsValidRWGroup(groupName, groups)
	assert.Equal(t, r, true)
}
