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
	"fmt"
	v1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/thoas/go-funk"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SynType string

func validate(rw *v1alpha1.RisingWave) error {
	if rw.Spec.MetaNode == nil {
		return fmt.Errorf("meta node spec nil")
	}

	if rw.Spec.ObjectStorage == nil {
		return fmt.Errorf("object storage spec nil")
	}

	if rw.Spec.ComputeNode == nil {
		return fmt.Errorf("compute node spec nil")
	}

	if rw.Spec.ComputeNode == nil {
		return fmt.Errorf("frontend spec nil")
	}
	return nil
}

func conditionsToMap(conditions []v1alpha1.RisingWaveCondition) map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition {
	obj := funk.ToMap(conditions, "Type")
	m := obj.(map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition)
	return m
}

func conditionMapToArray(m map[v1alpha1.RisingWaveType]v1alpha1.RisingWaveCondition) []v1alpha1.RisingWaveCondition {
	r := funk.Map(m, func(_ v1alpha1.RisingWaveType, v v1alpha1.RisingWaveCondition) v1alpha1.RisingWaveCondition {
		return v
	})
	return r.([]v1alpha1.RisingWaveCondition)
}

// markRunningCondition will update condition if necessary, and return bool flag
// if true, need update
// if false, no changes, ignore update
func markRunningCondition(rw *v1alpha1.RisingWave) bool {
	m := conditionsToMap(rw.Status.Condition)
	v, e := m[v1alpha1.Running]
	if !e || v.Status == metav1.ConditionFalse {
		m[v1alpha1.Running] = v1alpha1.RisingWaveCondition{
			Type:               v1alpha1.Running,
			LastTransitionTime: metav1.Now(),
			Status:             metav1.ConditionTrue,
		}
		rw.Status.Condition = conditionMapToArray(m)
		return true
	}

	rw.Status.Condition = conditionMapToArray(m)
	return false
}
