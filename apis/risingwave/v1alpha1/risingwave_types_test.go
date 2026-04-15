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

package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func TestRisingWave_DeepCopy_CopiesFrontendStatefulSetFlag(t *testing.T) {
	risingwave := &RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rw",
			Namespace: "default",
		},
		Spec: RisingWaveSpec{
			EnableOpenKruise:          ptr.To(true),
			EnableFrontendStatefulSet: ptr.To(true),
			EnableStandaloneMode:      ptr.To(false),
		},
	}

	copy := risingwave.DeepCopy()

	if copy == risingwave {
		t.Fatal("DeepCopy should return a different object")
	}

	if copy.Spec.EnableFrontendStatefulSet == nil || !*copy.Spec.EnableFrontendStatefulSet {
		t.Fatal("EnableFrontendStatefulSet was not copied")
	}

	if copy.Spec.EnableFrontendStatefulSet == risingwave.Spec.EnableFrontendStatefulSet {
		t.Fatal("EnableFrontendStatefulSet pointer should be deep copied")
	}

	*copy.Spec.EnableFrontendStatefulSet = false
	if !*risingwave.Spec.EnableFrontendStatefulSet {
		t.Fatal("mutating the copied flag should not affect the original")
	}
}
