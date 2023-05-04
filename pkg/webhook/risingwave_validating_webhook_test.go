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

package webhook

import (
	"context"
	"testing"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/pointer"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
)

func Test_RisingWaveValidatingWebhook_ValidateDelete(t *testing.T) {
	assert.Nil(t, NewRisingWaveValidatingWebhook(false).ValidateDelete(context.Background(), &risingwavev1alpha1.RisingWave{}))
}

func Test_RisingWaveValidatingWebhook_ValidateCreate(t *testing.T) {
	testcases := map[string]struct {
		patch               func(r *risingwavev1alpha1.RisingWave)
		pass                bool
		openKruiseAvailable bool
	}{
		"fake-pass": {
			patch: nil,
			pass:  true,
		},
		"openKruise-enabled-and-available": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"invalid-enable-openKruise-not-available": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
			},
			pass: false,
		},
		"invalid-image-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Image = "1234_"
			},
			pass: false,
		},
		"service-meta-labels-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					ServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1": "value1",
							"key2": "value2",
						},
					},
				}
			},
			pass: true,
		},
		"service-meta-labels-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					ServiceMeta: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
						Labels: map[string]string{
							"key1":            "value1",
							"risingwave/key2": "value2",
						},
					},
				}
			},
			pass: false,
		},
		"invalid-upgrade-strategy-type-InPlaceIfPossible-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.RisingWaveComponentGroupTemplate.UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"invalid-upgrade-strategy-type-InPlaceOnly-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.RisingWaveComponentGroupTemplate.UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"invalid-partition-str-val": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.RisingWaveComponentGroupTemplate.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "test-string",
						},
					},
				}
			},
			pass:                false,
			openKruiseAvailable: true,
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
				r.Spec.Global.Replicas.Meta = 0
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
		"upgrade-strategy-partition-valid-string": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveRollingUpdate{
						Partition: &intstr.IntOrString{
							Type:   intstr.String,
							StrVal: "50%",
						},
					},
				}
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"upgrade-strategy-InPlaceOnly-openKruise-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
				}
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"upgrade-strategy-InPlaceIfPossible-openKruise-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
				}
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"inPlace-strategy-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(false)
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type:                  risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{},
				}
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"inPlace-strategy-Recreate-openKruise-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
				r.Spec.Global.UpgradeStrategy = risingwavev1alpha1.RisingWaveUpgradeStrategy{
					Type:                  risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{},
				}
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"etcd-meta-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: true,
		},
		"empty-meta-storage-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{}
			},
			pass: false,
		},
		"meta-storage-with-default-values-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Memory: pointer.Bool(false),
					Etcd:   &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{},
				}
			},
			pass: false,
		},
		"multiple-meta-storages-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Memory: pointer.Bool(true),
					Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: false,
		},
		"minio-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{
						Endpoint: "minio",
						Bucket:   "hummock",
						RisingWaveMinIOCredentials: risingwavev1alpha1.RisingWaveMinIOCredentials{
							SecretName: "minio-creds",
						},
					},
				}
			},
			pass: true,
		},
		"s3-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket: "hummock",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							SecretName: "s3-creds",
						},
					},
				}
			},
			pass: true,
		},
		"gcs-object-storage-workload-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						UseWorkloadIdentity: true,
						Bucket:              "gcs-bucket",
						Root:                "gcs-root",
					},
				}
			},
			pass: true,
		},
		"gcs-object-storage-secret-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						UseWorkloadIdentity: false,
						Bucket:              "gcs-bucket",
						Root:                "gcs-root",
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							SecretName: "gcs-creds",
						},
					},
				}
			},
			pass: true,
		},
		"gcs-object-storage-both-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						UseWorkloadIdentity: true,
						Bucket:              "gcs-bucket",
						Root:                "gcs-root",
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							SecretName: "gcs-creds",
						},
					},
				}
			},
			pass: false,
		},
		"gcs-object-storage-none-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						UseWorkloadIdentity: false,
						Bucket:              "gcs-bucket",
						Root:                "gcs-root",
					},
				}
			},
			pass: false,
		},
		"gcs-object-storage-edge-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						UseWorkloadIdentity: false,
						Bucket:              "gcs-bucket",
						Root:                "gcs-root",
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							SecretName: "",
						},
					},
				}
			},
			pass: false,
		},
		"aliyun-oss-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					AliyunOSS: &risingwavev1alpha1.RisingWaveStateStoreBackendAliyunOSS{
						Bucket: "hummock",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							SecretName: "aliyun-oss-creds",
						},
					},
				}
			},
			pass: true,
		},
		"azure-blob-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{
						Container: "hummock",
						Root:      "azure-blob-root",
						Endpoint:  "https://accountName.blob.core.windows.net",
						RisingWaveAzureBlobCredentials: risingwavev1alpha1.RisingWaveAzureBlobCredentials{
							SecretName: "azure-blob-creds",
						},
					},
				}
			},
			pass: true,
		},
		"hdfs-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					HDFS: &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{
						NameNode: "test",
						Root:     "test",
					},
				}
			},
			pass: true,
		},
		"webhdfs-object-storage-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					WebHDFS: &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{
						NameNode: "test",
						Root:     "test",
					},
				}
			},
			pass: true,
		},
		"empty-object-storage-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{}
			},
			pass: false,
		},
		"multiple-object-storages-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					Memory: pointer.Bool(true),
					MinIO:  &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					Memory: pointer.Bool(true),
					S3:     &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-3": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					S3:    &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-4": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					HDFS:  &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-5": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO:     &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{},
				}
			},
			pass: false,
		},
		"multiple-object-storages-fail-6": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					HDFS:    &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{},
					WebHDFS: &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{},
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
				r.Spec.Storages.PVCTemplates = []risingwavev1alpha1.PersistentVolumeClaim{
					{
						PersistentVolumeClaimPartialObjectMeta: risingwavev1alpha1.PersistentVolumeClaimPartialObjectMeta{
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
				r.Spec.Storages.PVCTemplates = []risingwavev1alpha1.PersistentVolumeClaim{
					{
						PersistentVolumeClaimPartialObjectMeta: risingwavev1alpha1.PersistentVolumeClaimPartialObjectMeta{
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
				r.Spec.Storages.PVCTemplates = []risingwavev1alpha1.PersistentVolumeClaim{
					{
						PersistentVolumeClaimPartialObjectMeta: risingwavev1alpha1.PersistentVolumeClaimPartialObjectMeta{
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
		"insufficient-resources-cpu-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": resource.MustParse("250m"),
							},
							Requests: corev1.ResourceList{
								"cpu": resource.MustParse("1000m"),
							},
						},
					},
				}
			},
			pass: false,
		},
		"insufficient-resources-memory-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": resource.MustParse("100Mi"),
							},
							Requests: corev1.ResourceList{
								"memory": resource.MustParse("1Gi"),
							},
						},
					},
				}
			},
			pass: false,
		},
		"pods-meta-labels-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Labels: map[string]string{
								"key1": "value1",
								"key2": "value2",
							},
						},
					},
				}
			},
			pass: true,
		},
		"pods-meta-labels-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Metadata: risingwavev1alpha1.RisingWavePodTemplatePartialObjectMeta{
							Labels: map[string]string{
								"key1":            "value1",
								"risingwave/key2": "value2",
							},
						},
					},
				}
			},
			pass: false,
		},
		"limit-not-exist-pass-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("1"),
								"memory": resource.MustParse("100Mi"),
							},
						},
					},
				}
			},
			pass: true,
		},
		"limit-not-exist-pass-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": resource.MustParse("1"),
							},
							Requests: corev1.ResourceList{
								"memory": resource.MustParse("100Mi"),
							},
						},
					},
				}
			},
			pass: true,
		},
		"limit-zero-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu": resource.MustParse("0"),
							},
							Requests: corev1.ResourceList{
								"cpu": resource.MustParse("1"),
							},
						},
					},
				}
			},
			pass: false,
		},
		"limit-zero-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"memory": resource.MustParse("0"),
							},
							Requests: corev1.ResourceList{
								"memory": resource.MustParse("100Mi"),
							},
						},
					},
				}
			},
			pass: false,
		},
		"limit-exist-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global = risingwavev1alpha1.RisingWaveGlobalSpec{
					RisingWaveComponentGroupTemplate: risingwavev1alpha1.RisingWaveComponentGroupTemplate{
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								"cpu":    resource.MustParse("1"),
								"memory": resource.MustParse("1Gi"),
							},
							Requests: corev1.ResourceList{
								"cpu":    resource.MustParse("100m"),
								"memory": resource.MustParse("100Mi"),
							},
						},
					},
				}
			},
			pass: true,
		},
		"multi-memory-meta-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Replicas.Meta = 2
			},
			pass: false,
		},
		"multi-etcd-meta-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Replicas.Meta = 2
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
						Endpoint: "etcd",
						RisingWaveEtcdCredentials: &risingwavev1alpha1.RisingWaveEtcdCredentials{
							SecretName: "etcd-credentials",
						},
					},
				}
			},
			pass: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			risingwave := testutils.FakeRisingWave()
			if tc.patch != nil {
				tc.patch(risingwave)
			}

			// we let webhook take on open kruise availability specified in Test case.
			webhook := NewRisingWaveValidatingWebhook(tc.openKruiseAvailable)

			err := webhook.ValidateCreate(context.Background(), risingwave)
			if tc.pass != (err == nil) {
				t.Fatal(tc.pass, err)
			}
		})
	}
}

func Test_RisingWaveValidatingWebhook_ValidateUpdate(t *testing.T) {
	testcases := map[string]struct {
		patch               func(r *risingwavev1alpha1.RisingWave)
		pass                bool
		openKruiseAvailable bool
		oldObjMutation      func(r *risingwavev1alpha1.RisingWave)
	}{
		"storages-unchanged-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Global.Replicas.Meta = 1
			},
			pass: true,
		},
		"enable-openKruise-when-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
			},
			pass: false,
		},
		"enable-openKruise-when-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
			},
			openKruiseAvailable: true,
			pass:                true,
		},
		"disabled-openKruise-when-not-available": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(false)
			},
			pass: true,
			oldObjMutation: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = pointer.Bool(true)
				r.Spec.Global.UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly
			},
		},
		"illegal-changes-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{}
			},
			pass: false,
		},
		"meta-storage-changed-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: false,
		},
		"object-storage-changed-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
				}
			},
			pass: false,
		},
		"object-storage-changed-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"empty-image-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Image = ""
			},
			pass: false,
		},
		"empty-image-and-empty-component-images-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Image = ""
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
				r.Spec.Components.Connector.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
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
				r.Spec.Components.Connector.Groups = []risingwavev1alpha1.RisingWaveComponentGroup{
					{
						Name:     "a",
						Replicas: 1,
						RisingWaveComponentGroupTemplate: &risingwavev1alpha1.RisingWaveComponentGroupTemplate{
							Image: "ghcr.io/risingwavelabs/risingwave:latest",
						},
					},
				}
			},
			pass: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {

			// We want to create two copies, so we can compare the old state and new state
			// when transitioning from openkruise enabled to disabled with operator disabled.
			risingwave := testutils.FakeRisingWave()
			oldObj := risingwave.DeepCopy()
			if tc.oldObjMutation != nil {
				tc.oldObjMutation(oldObj)
			}
			tc.patch(risingwave)

			webhook := NewRisingWaveValidatingWebhook(tc.openKruiseAvailable)

			// test when operator is disabled and openKruise enabled -> disabled, risingwave should be set to default.
			if name == "disabled-openKruise-when-not-available" {
				if risingwave.Spec.Global.UpgradeStrategy.Type == oldObj.Spec.Global.UpgradeStrategy.Type {
					t.Fatal("Risingwave is not default")
				}
			}
			err := webhook.ValidateUpdate(context.Background(), oldObj, risingwave)
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

			err := NewRisingWaveValidatingWebhook(false).ValidateUpdate(context.Background(), obj, newObj)
			if tc.pass {
				assert.Nil(t, err, "unexpected error")
			} else {
				assert.NotNil(t, err, "should be nil")
			}
		})
	}
}
