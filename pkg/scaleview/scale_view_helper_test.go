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

package scaleview

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func TestRisingWaveScaleViewHelper_GetGroupIndex(t *testing.T) {
	for _, component := range []string{
		consts.ComponentFrontend, consts.ComponentMeta, consts.ComponentCompactor, consts.ComponentCompute,
	} {
		helper := NewRisingWaveScaleViewHelper(testutils.FakeRisingWave(), component)
		index, ok := helper.GetGroupIndex("")
		assert.True(t, ok, "empty group must be found")
		assert.Equal(t, 0, index, "empty group must be with index 0")

		_, ok = helper.GetGroupIndex("a")
		assert.False(t, ok, "not existing group")
	}
}

func TestRisingWaveScaleViewHelper_ListComponentGroups(t *testing.T) {
	for _, component := range []string{
		consts.ComponentFrontend, consts.ComponentMeta, consts.ComponentCompactor, consts.ComponentCompute,
	} {
		helper := NewRisingWaveScaleViewHelper(testutils.FakeRisingWaveComponentOnly(), component)
		groups := helper.ListComponentGroups()
		assert.Equal(t, []string{testutils.GetGroupName(0)}, groups, "component group names should be listed")
	}
}

func TestRisingWaveScaleViewHelper_ReadReplicas(t *testing.T) {
	for _, component := range []string{
		consts.ComponentFrontend, consts.ComponentMeta, consts.ComponentCompactor, consts.ComponentCompute,
	} {
		helper := NewRisingWaveScaleViewHelper(testutils.FakeRisingWaveComponentOnly(), component)
		replicas, ok := helper.ReadReplicas(testutils.GetGroupName(0))
		assert.True(t, ok, "group must be found")
		assert.Equal(t, int32(1), replicas, "replicas must be 1")

		_, ok = helper.ReadReplicas(testutils.GetGroupName(1))
		assert.False(t, ok, "group of index 1 doesn't exist")
	}
}

func TestRisingWaveScaleViewHelper_WriteReplicas(t *testing.T) {
	for _, component := range []string{
		consts.ComponentFrontend, consts.ComponentMeta, consts.ComponentCompactor, consts.ComponentCompute,
	} {
		helper := NewRisingWaveScaleViewHelper(testutils.FakeRisingWaveComponentOnly(), component)

		group := testutils.GetGroupName(0)
		replicas, _ := helper.ReadReplicas(group)
		assert.Equal(t, int32(1), replicas, "replicas must be 1")

		ok := helper.WriteReplicas(group, 2)
		assert.True(t, ok)

		replicas, _ = helper.ReadReplicas(group)
		assert.Equal(t, int32(2), replicas, "replicas must be 2")

		ok = helper.WriteReplicas(group, 2)
		assert.False(t, ok)
	}
}
