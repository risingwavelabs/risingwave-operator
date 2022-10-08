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

package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveValidatingWebhook_ValidateDelete(t *testing.T) {
	assert.Nil(t, NewRisingWaveValidatingWebhook().ValidateDelete(context.Background(), &risingwavev1alpha1.RisingWave{}))
}

func Test_RisingWaveValidatingWebhook_ValidateCreate(t *testing.T) {
	testcases := map[string]struct {
		patch func(r *risingwavev1alpha1.RisingWave)
		pass  bool
	}{
		"fake-pass": {
			patch: nil,
			pass:  true,
		},
		"invalid-image-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Image = "1234_"
			},
			pass: false,
		},
		"invalid-image-in-compute-group-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
							RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
								Image: "abc@/def:123",
							},
						},
					},
				}
			},
			pass: false,
		},
		"invalid-image-in-component-group-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
							Image: "abc@/def:123",
						},
					},
				}
			},
			pass: false,
		},
		"meta-group-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
			},
			pass: true,
		},
		"frontend-group-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
			},
			pass: true,
		},
		"compactor-group-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
			},
			pass: true,
		},
		"rolling-upgrade-nil-when-create-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				}
			},
			pass: true,
		},
		"rolling-upgrade-not-nil-when-create-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type:          risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{},
				}
			},
			pass: false,
		},
		"rolling-upgrade-nil-when-set-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
				}
			},
			pass: true,
		},
		"rolling-upgrade-not-nil-when-set-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type:          risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{},
				}
			},
			pass: true,
		},
		"etcd-meta-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{
					Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: true,
		},
		"empty-meta-storage-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{}
			},
			pass: false,
		},
		"meta-storage-with-default-values-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{
					Memory: pointer.Bool(false),
					Etcd:   &risingwavev1alpha1.RisingWaveMetaStorageEtcd{},
				}
			},
			pass: false,
		},
		"multiple-meta-storages-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{
					Memory: pointer.Bool(true),
					Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: false,
		},
		"minio-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{
						Secret:   "minio-creds",
						Endpoint: "minio",
						Bucket:   "hummock",
					},
				}
			},
			pass: true,
		},
		"s3-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{
						Secret: "s3-creds",
						Bucket: "hummock",
					},
				}
			},
			pass: true,
		},
		"empty-object-storage-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{}
			},
			pass: false,
		},
		"multiple-object-storages-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					Memory: pointer.Bool(true),
					MinIO:  &risingwavev1alpha1.RisingWaveObjectStorageMinIO{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					Memory: pointer.Bool(true),
					S3:     &risingwavev1alpha1.RisingWaveObjectStorageS3{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-3": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{},
					S3:    &risingwavev1alpha1.RisingWaveObjectStorageS3{},
				}
			},
			pass: false,
		},
		"empty-configuration-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Configuration.ConfigMap = &corev1.ConfigMapKeySelector{}
			},
			pass: false,
		},
		"half-empty-configuration-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Configuration.ConfigMap = &corev1.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "a",
					},
				}
			},
			pass: false,
		},
		"half-empty-configuration-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Configuration.ConfigMap = &corev1.ConfigMapKeySelector{
					Key: "a",
				}
			},
			pass: false,
		},
		"pvc-mounts-match-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.PVCTemplates = []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc1",
						},
					},
				}
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: "pvc1",
								},
							},
						},
					},
				}
			},
			pass: true,
		},
		"pvc-not-mounted-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.PVCTemplates = []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc1",
						},
					},
				}
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
			},
			pass: true,
		},
		"pvc-mounts-not-match-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.PVCTemplates = []corev1.PersistentVolumeClaim{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "pvc1",
						},
					},
				}
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
							VolumeMounts: []corev1.VolumeMount{
								{
									Name: "pvc0",
								},
							},
						},
					},
				}
			},
			pass: false,
		},
	}

	webhook := NewRisingWaveValidatingWebhook()

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := testutils.FakeRisingWave()
			if tc.patch != nil {
				tc.patch(risingwave)
			}

			err := webhook.ValidateCreate(context.Background(), risingwave)
			if tc.pass != (err == nil) {
				t.Fatal(tc.pass, err)
			}
		})
	}
}

func Test_RisingWaveValidatingWebhook_ValidateUpdate(t *testing.T) {
	testcases := map[string]struct {
		patch func(r *risingwavev1alpha1.RisingWave)
		pass  bool
	}{
		"storages-unchanged-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Replicas.Meta = 1
			},
			pass: true,
		},
		"illegal-changes-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{}
			},
			pass: false,
		},
		"meta-storage-changed-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Meta = risingwavev1alpha1.RisingWaveMetaStorage{
					Etcd: &risingwavev1alpha1.RisingWaveMetaStorageEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: false,
		},
		"object-storage-changed-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					MinIO: &risingwavev1alpha1.RisingWaveObjectStorageMinIO{},
				}
			},
			pass: false,
		},
		"object-storage-changed-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Storages.Object = risingwavev1alpha1.RisingWaveObjectStorage{
					S3: &risingwavev1alpha1.RisingWaveObjectStorageS3{},
				}
			},
			pass: false,
		},
		"empty-global-image-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Image = ""
			},
			pass: false,
		},
		"empty-global-image-and-empty-component-images-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Image = ""
				r.Spec.Global.Replicas = risingwavev1alpha1.RisingWaveGlobalReplicas{}
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
				r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
				r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
					},
				}
			},
			pass: false,
		},
		"empty-global-image-but-component-images-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Image = ""
				r.Spec.Global.Replicas = risingwavev1alpha1.RisingWaveGlobalReplicas{}
				r.Spec.Components.Meta.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
							Image: "ghcr.io/risingwavelabs/risingwave:latest",
						},
					},
				}
				r.Spec.Components.Frontend.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
							Image: "ghcr.io/risingwavelabs/risingwave:latest",
						},
					},
				}
				r.Spec.Components.Compactor.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
							Image: "ghcr.io/risingwavelabs/risingwave:latest",
						},
					},
				}
				r.Spec.Components.Compute.Groups = []risingwavev1alpha1.RisingWaveComputeGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
							RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
								Image: "ghcr.io/risingwavelabs/risingwave:latest",
							},
						},
					},
				}
			},
			pass: true,
		},
	}

	webhook := NewRisingWaveValidatingWebhook()

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := testutils.FakeRisingWave()
			tc.patch(risingwave)

			err := webhook.ValidateUpdate(context.Background(), testutils.FakeRisingWave(), risingwave)
			if tc.pass != (err == nil) {
				t.Fatal(tc.pass, err)
			}
		})
	}
}

func Test_RisingWaveValidatingWebhook_ValidateUpdate_ScaleViews(t *testing.T) {
	testcases := map[string]struct {
		origin      *risingwavev1alpha1.RisingWave
		mutate      func(wave *risingwavev1alpha1.RisingWave)
		statusPatch func(status *risingwavev1alpha1.RisingWaveStatus)
		pass        bool
	}{
		"empty-scale-views": {
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Global.Replicas.Frontend++
			},
			pass: true,
		},
		"scale-views-on-frontend": {
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Global.Replicas.Frontend++
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentFrontend,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     "",
							Replicas: 0,
						},
					},
				})
			},
			pass: false,
		},
		"scale-views-on-compactor-but-update-frontend": {
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Global.Replicas.Frontend++
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentCompactor,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     "",
							Replicas: 0,
						},
					},
				})
			},
			pass: true,
		},
		"delete-locked-group": {
			origin: testutils.FakeRisingWaveComponentOnly(),
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Components.Frontend.Groups = nil
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentFrontend,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     testutils.GetGroupName(0),
							Replicas: 0,
						},
					},
				})
			},
			pass: false,
		},
		"multiple-locked-groups": {
			origin: testutils.FakeRisingWaveComponentOnly(),
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Components.Frontend.Groups[0].Replicas = 2
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentFrontend,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     testutils.GetGroupName(0),
							Replicas: 2,
						},
					},
				}, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentCompactor,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     testutils.GetGroupName(0),
							Replicas: 1,
						},
					},
				})
			},
			pass: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			obj := tc.origin
			if obj == nil {
				obj = testutils.FakeRisingWave()
			}

			if tc.statusPatch != nil {
				tc.statusPatch(&obj.Status)
			}

			newObj := obj.DeepCopy()
			tc.mutate(newObj)

			err := NewRisingWaveValidatingWebhook().ValidateUpdate(context.Background(), obj, newObj)
			if tc.pass {
				assert.Nil(t, err, "unexpected error")
			} else {
				assert.NotNil(t, err, "should be nil")
			}
		})
	}
}
