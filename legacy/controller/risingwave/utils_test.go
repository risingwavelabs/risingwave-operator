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

package risingwave

import (
	"testing"

	"gotest.tools/v3/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

var conditions = []v1alpha1.RisingWaveCondition{
	{
		Type:   v1alpha1.Initializing,
		Status: metav1.ConditionTrue,
	},
	{
		Type:   v1alpha1.Running,
		Status: metav1.ConditionFalse,
	},
}

func Test_conditionsToMap(t *testing.T) {
	m := conditionsToMap(conditions)
	assert.Equal(t, len(m), 2)
	v, e := m[v1alpha1.Running]
	assert.Equal(t, e, true)
	assert.Equal(t, v.Status, metav1.ConditionFalse)
}

func Test_conditionMapToArray(t *testing.T) {
	m := conditionsToMap(conditions)
	arr := conditionMapToArray(m)
	assert.Equal(t, len(arr), 2)
}
