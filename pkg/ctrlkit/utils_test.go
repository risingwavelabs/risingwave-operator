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

package ctrlkit

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Test_ValidateOwnership(t *testing.T) {
	uid := uuid.NewUUID()

	testcases := map[string]struct {
		owner client.Object
		ownee client.Object
		owns  bool
	}{
		"nil-not-owns": {
			owner: nil,
			ownee: &corev1.Service{},
		},
		"not-owns-nil": {
			owner: &corev1.Service{},
			ownee: nil,
		},
		"owns-by-controller-ref": {
			owner: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: uid,
				},
			},
			ownee: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							UID:        uid,
							Controller: pointer.Bool(true),
						},
					},
				},
			},
			owns: true,
		},
		"not-owns-without-controller-ref": {
			owner: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: uid,
				},
			},
			ownee: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							UID: uid,
						},
					},
				},
			},
		},
		"not-owns-by-different-controller": {
			owner: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					UID: uid,
				},
			},
			ownee: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					OwnerReferences: []metav1.OwnerReference{
						{
							UID:        uuid.NewUUID(),
							Controller: pointer.Bool(true),
						},
					},
				},
			},
			owns: false,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			r := ValidateOwnership(tc.ownee, tc.owner)
			if r != tc.owns {
				t.Fail()
			}
		})
	}
}
