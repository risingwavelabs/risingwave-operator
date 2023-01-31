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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

func Test_ConvertRisingwave(t *testing.T) {
	fake := FakeRW()
	newWave := ConvertRisingwave(fake)
	assert.Equal(t, len(newWave.Spec.Components.Frontend.Groups), len(fake.Spec.Components.Frontend.Groups)+1)
	assert.Equal(t, len(newWave.Spec.Components.Frontend.Groups), 3)
	var exist bool
	var computeGroup v1alpha1.RisingWaveComputeGroup
	for _, g := range newWave.Spec.Components.Compute.Groups {
		if g.Name == DefaultGroup {
			exist = true
			computeGroup = g
		}
	}
	assert.Equal(t, exist, true)
	assert.Equal(t, computeGroup.Image, "test.image.global")
	var cpu = computeGroup.Resources.Limits[corev1.ResourceCPU]
	assert.Equal(t, cpu, resource.MustParse("1"))
}
