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
	"context"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_NewFakeClient(t *testing.T) {
	c := NewFakeClient()
	ns := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				"test-k": "test-v",
			},
		},
	}
	err := c.Create(context.Background(), &ns)
	assert.Equal(t, err, nil)

	var newNS corev1.Namespace
	err = c.Get(context.Background(), client.ObjectKey{Name: "test"}, &newNS)
	assert.Equal(t, err, nil)
	assert.Equal(t, ns.Name, newNS.Name)
	assert.Equal(t, len(newNS.Annotations), 1)
}
