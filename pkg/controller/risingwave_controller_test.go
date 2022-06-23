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
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

var schemeForTest = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(schemeForTest)
	_ = risingwavev1alpha1.AddToScheme(schemeForTest)
}

func Test_RisingWaveController_DryRun(t *testing.T) {
	controller := &RisingWaveController{
		Client: fake.NewClientBuilder().
			WithScheme(schemeForTest).
			WithObjects(&risingwavev1alpha1.RisingWave{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "example",
					Namespace: "default",
				},
				Spec: risingwavev1alpha1.RisingWaveSpec{
					ObjectStorage: &risingwavev1alpha1.ObjectStorageSpec{},
				},
				Status: risingwavev1alpha1.RisingWaveStatus{
					Conditions: []risingwavev1alpha1.RisingWaveCondition{
						{
							Type:   risingwavev1alpha1.Initializing,
							Status: metav1.ConditionTrue,
						},
					},
				},
			}).
			Build(),
		DryRun: true,
	}

	_, err := controller.Reconcile(context.Background(), reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "example",
			Namespace: "default",
		},
	})

	if err != nil {
		t.Fatal(err)
	}
}
