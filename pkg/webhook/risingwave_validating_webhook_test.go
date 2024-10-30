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
	"strings"
	"testing"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	"github.com/samber/lo"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/testutils"
	"github.com/risingwavelabs/risingwave-operator/pkg/utils"
)

func Test_RisingWaveValidatingWebhook_ValidateDelete(t *testing.T) {
	_, err := NewRisingWaveValidatingWebhook(false).ValidateDelete(context.Background(), &risingwavev1alpha1.RisingWave{})
	assert.Nil(t, err)
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
				r.Spec.EnableOpenKruise = ptr.To(true)
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"invalid-enable-openKruise-not-available": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
			},
			pass: false,
		},
		"invalid-image-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Image = "1234_"
			},
			pass: false,
		},
		"service-meta-labels-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.AdditionalFrontendServiceMetadata = risingwavev1alpha1.PartialObjectMeta{
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				}
			},
			pass: true,
		},
		"service-meta-labels-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.AdditionalFrontendServiceMetadata = risingwavev1alpha1.PartialObjectMeta{
					Labels: map[string]string{
						"key1":            "value1",
						"risingwave/key2": "value2",
					},
				}
			},
			pass: false,
		},
		"invalid-upgrade-strategy-type-InPlaceIfPossible-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"invalid-upgrade-strategy-type-InPlaceOnly-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"invalid-partition-str-val": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
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
				r.Spec.Components.Compute.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Name:     "a",
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Image: "abc@/def:123",
								},
							},
						},
					},
				}
			},
			pass: false,
		},
		"invalid-image-in-component-group-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
					{
						Name:     "a",
						Replicas: 1,
						Template: risingwavev1alpha1.RisingWaveNodePodTemplate{
							Spec: risingwavev1alpha1.RisingWaveNodePodTemplateSpec{
								RisingWaveNodeContainer: risingwavev1alpha1.RisingWaveNodeContainer{
									Image: "abc@/def:123",
								},
							},
						},
					},
				}
			},
			pass: false,
		},
		"meta-group-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Replicas = 0
				r.Spec.Components.Meta.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
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
				r.Spec.Components.Frontend.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
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
				r.Spec.Components.Compactor.NodeGroups = []risingwavev1alpha1.RisingWaveNodeGroup{
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
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
				}
			},
			pass: true,
		},
		"rolling-upgrade-not-nil-when-create-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type:          risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{},
				}
			},
			pass: false,
		},
		"rolling-upgrade-nil-when-set-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
				}
			},
			pass: true,
		},
		"rolling-upgrade-not-nil-when-set-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type:          risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{},
				}
			},
			pass: true,
		},
		"upgrade-strategy-partition-valid-string": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					RollingUpdate: &risingwavev1alpha1.RisingWaveNodeGroupRollingUpdate{
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
				r.Spec.EnableOpenKruise = ptr.To(true)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly,
				}
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"upgrade-strategy-InPlaceIfPossible-openKruise-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type: risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
				}
			},
			pass:                true,
			openKruiseAvailable: true,
		},
		"inPlace-strategy-openKruise-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(false)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
					Type:                  risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible,
					InPlaceUpdateStrategy: &kruisepubs.InPlaceUpdateStrategy{},
				}
			},
			pass:                false,
			openKruiseAvailable: true,
		},
		"inPlace-strategy-Recreate-openKruise-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy = risingwavev1alpha1.RisingWaveNodeGroupUpgradeStrategy{
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
					Memory: ptr.To(false),
					Etcd:   &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{},
				}
			},
			pass: false,
		},
		"multiple-meta-storages-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.MetaStore = risingwavev1alpha1.RisingWaveMetaStoreBackend{
					Memory: ptr.To(true),
					Etcd: &risingwavev1alpha1.RisingWaveMetaStoreBackendEtcd{
						Endpoint: "etcd",
					},
				}
			},
			pass: false,
		},
		"minio-state-store-pass": {
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
		"s3-state-store-pass": {
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
		"s3-state-store-use-service-account-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket: "hummock",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							UseServiceAccount: ptr.To(true),
						},
					},
				}
			},
			pass: true,
		},
		"s3-state-store-use-service-account-pass-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket: "hummock",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							UseServiceAccount: ptr.To(true),
							SecretName:        "s3-creds",
						},
					},
				}
			},
			pass: true,
		},
		"s3-state-store-no-credentials-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket: "hummock",
					},
				}
			},
			pass: false,
		},
		"s3-compatible-state-store-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket:   "hummock",
						Endpoint: "123",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							SecretName: "s3-creds",
						},
					},
				}
			},
			pass: true,
		},
		"s3-compatible-state-store-use-service-account-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{
						Bucket:   "hummock",
						Endpoint: "123",
						RisingWaveS3Credentials: risingwavev1alpha1.RisingWaveS3Credentials{
							UseServiceAccount: ptr.To(true),
						},
					},
				}
			},
			pass: false,
		},
		"gcs-state-store-workload-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							UseWorkloadIdentity: ptr.To(true),
						},
						Bucket: "gcs-bucket",
						Root:   "gcs-root",
					},
				}
			},
			pass: true,
		},
		"gcs-state-store-workload-pass-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							UseWorkloadIdentity: ptr.To(true),
							SecretName:          "gcs-creds",
						},
						Bucket: "gcs-bucket",
						Root:   "gcs-root",
					},
				}
			},
			pass: true,
		},
		"gcs-state-store-secret-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						Bucket: "gcs-bucket",
						Root:   "gcs-root",
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							SecretName: "gcs-creds",
						},
					},
				}
			},
			pass: true,
		},
		"gcs-state-store-none-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						Bucket: "gcs-bucket",
						Root:   "gcs-root",
					},
				}
			},
			pass: false,
		},
		"gcs-state-store-edge-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					GCS: &risingwavev1alpha1.RisingWaveStateStoreBackendGCS{
						Bucket: "gcs-bucket",
						Root:   "gcs-root",
						RisingWaveGCSCredentials: risingwavev1alpha1.RisingWaveGCSCredentials{
							SecretName: "",
						},
					},
				}
			},
			pass: false,
		},
		"aliyun-oss-state-store-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					AliyunOSS: &risingwavev1alpha1.RisingWaveStateStoreBackendAliyunOSS{
						Bucket:           "hummock",
						Root:             "AliyunOSS-root",
						Region:           "cn-hangzhou",
						InternalEndpoint: false,
						RisingWaveAliyunOSSCredentials: risingwavev1alpha1.RisingWaveAliyunOSSCredentials{
							SecretName: "AliyunOSS-creds",
						},
					},
				}
			},
			pass: true,
		},
		"azure-blob-state-store-pass": {
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
		"hdfs-state-store-pass": {
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
		"webhdfs-state-store-pass": {
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
		"empty-state-store-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{}
			},
			pass: false,
		},
		"multiple-state-stores-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					Memory: ptr.To(true),
					MinIO:  &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
				}
			},
			pass: false,
		},
		"multiple-state-stores-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					Memory: ptr.To(true),
					S3:     &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"multiple-state-stores-fail-3": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					S3:    &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"multiple-state-stores-fail-4": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					HDFS:  &risingwavev1alpha1.RisingWaveStateStoreBackendHDFS{},
				}
			},
			pass: false,
		},
		"multiple-state-stores-fail-5": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO:     &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
					AzureBlob: &risingwavev1alpha1.RisingWaveStateStoreBackendAzureBlob{},
				}
			},
			pass: false,
		},
		"multiple-state-stores-fail-6": {
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
				r.Spec.Configuration.ConfigMap = &risingwavev1alpha1.RisingWaveNodeConfigurationConfigMapSource{}
			},
			pass: false,
		},
		"half-empty-configuration-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Configuration.ConfigMap = &risingwavev1alpha1.RisingWaveNodeConfigurationConfigMapSource{
					Name: "a",
				}
			},
			pass: false,
		},
		"half-empty-configuration-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Configuration.ConfigMap = &risingwavev1alpha1.RisingWaveNodeConfigurationConfigMapSource{
					Key: "a",
				}
			},
			pass: false,
		},
		"insufficient-resources-cpu-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu": resource.MustParse("250m"),
					},
					Requests: corev1.ResourceList{
						"cpu": resource.MustParse("1000m"),
					},
				}
			},
			pass: false,
		},
		"insufficient-resources-memory-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"memory": resource.MustParse("100Mi"),
					},
					Requests: corev1.ResourceList{
						"memory": resource.MustParse("1Gi"),
					},
				}
			},
			pass: false,
		},
		"pods-meta-labels-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.ObjectMeta = risingwavev1alpha1.PartialObjectMeta{
					Labels: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				}
			},
			pass: true,
		},
		"pods-meta-labels-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.ObjectMeta = risingwavev1alpha1.PartialObjectMeta{
					Labels: map[string]string{
						"key1":            "value1",
						"risingwave/key2": "value2",
					},
				}
			},
			pass: false,
		},
		"limit-not-exist-pass-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("100Mi"),
					},
				}
			},
			pass: true,
		},
		"limit-not-exist-pass-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu": resource.MustParse("1"),
					},
					Requests: corev1.ResourceList{
						"memory": resource.MustParse("100Mi"),
					},
				}
			},
			pass: true,
		},
		"limit-zero-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu": resource.MustParse("0"),
					},
					Requests: corev1.ResourceList{
						"cpu": resource.MustParse("1"),
					},
				}
			},
			pass: false,
		},
		"limit-zero-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"memory": resource.MustParse("0"),
					},
					Requests: corev1.ResourceList{
						"memory": resource.MustParse("100Mi"),
					},
				}
			},
			pass: false,
		},
		"limit-exist-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Template.Spec.Resources = corev1.ResourceRequirements{
					Limits: corev1.ResourceList{
						"cpu":    resource.MustParse("1"),
						"memory": resource.MustParse("1Gi"),
					},
					Requests: corev1.ResourceList{
						"cpu":    resource.MustParse("100m"),
						"memory": resource.MustParse("100Mi"),
					},
				}
			},
			pass: true,
		},
		"multi-memory-meta-fail": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Replicas = 2
			},
			pass: false,
		},
		"multi-etcd-meta-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Replicas = 2
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
		"invalid-data-dir-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "/"
			},
			pass: false,
		},
		"invalid-data-dir-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "/a"
			},
			pass: false,
		},
		"invalid-data-dir-3": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "a/"
			},
			pass: false,
		},
		"invalid-data-dir-4": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "a//b"
			},
			pass: false,
		},
		"invalid-data-dir-5": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = strings.Repeat("a", 801)
			},
			pass: false,
		},
		"valid-data-dir-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "a"
			},
			pass: true,
		},
		"valid-data-dir-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore.DataDirectory = "a/b"
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

			_, err := webhook.ValidateCreate(context.Background(), risingwave)
			if tc.pass != (err == nil) {
				t.Fatal(tc.pass, err)
			}
		})
	}
}

func Test_RisingWaveValidatingWebhook_ValidateUpdate(t *testing.T) {
	testcases := map[string]struct {
		init                func(r *risingwavev1alpha1.RisingWave)
		patch               func(r *risingwavev1alpha1.RisingWave)
		pass                bool
		openKruiseAvailable bool
		oldObjMutation      func(r *risingwavev1alpha1.RisingWave)
	}{
		"storages-unchanged-pass": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Meta.NodeGroups[0].Replicas = 1
			},
			pass: true,
		},
		"enable-openKruise-when-disabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
			},
			pass: false,
		},
		"enable-openKruise-when-enabled": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
			},
			openKruiseAvailable: true,
			pass:                true,
		},
		"disabled-openKruise-when-not-available": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(false)
			},
			pass: true,
			oldObjMutation: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.EnableOpenKruise = ptr.To(true)
				r.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy.Type = risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly
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
		"state-store-changed-fail-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					MinIO: &risingwavev1alpha1.RisingWaveStateStoreBackendMinIO{},
				}
			},
			pass: false,
		},
		"state-store-changed-fail-2": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.StateStore = risingwavev1alpha1.RisingWaveStateStoreBackend{
					S3: &risingwavev1alpha1.RisingWaveStateStoreBackendS3{},
				}
			},
			pass: false,
		},
		"secret-store-nil-unchanged-success": { // nil secret store
			patch: func(r *risingwavev1alpha1.RisingWave) {},
			pass:  true,
		},
		"secret-store-changed-from-nil-success-0": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To(lo.Must(utils.RandomHex(32)))
			},
			pass: true,
		},
		"secret-store-changed-from-nil-success-1": {
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.SecretRef = &risingwavev1alpha1.RisingWaveSecretStorePrivateKeySecretReference{
					Name: "test",
					Key:  "test",
				}
			},
			pass: true,
		},
		"secret-store-changed-from-non-nil-to-nil-fail-0": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To(lo.Must(utils.RandomHex(32)))
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore = risingwavev1alpha1.RisingWaveSecretStore{}
			},
			pass: false,
		},
		"secret-store-changed-from-non-nil-to-nil-fail-1": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.SecretRef = &risingwavev1alpha1.RisingWaveSecretStorePrivateKeySecretReference{
					Name: "test",
					Key:  "test",
				}
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore = risingwavev1alpha1.RisingWaveSecretStore{}
			},
			pass: false,
		},
		"secret-store-unchanged-from-value-to-value-success": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("123")
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("123")
			},
			pass: true,
		},
		"secret-store-changed-from-value-to-value-fail": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("123")
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("456")
			},
			pass: false,
		},
		"secret-store-changed-from-value-to-secret-success": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("123")
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.SecretRef = &risingwavev1alpha1.RisingWaveSecretStorePrivateKeySecretReference{
					Name: "test",
					Key:  "test",
				}
			},
			pass: true,
		},
		"secret-store-changed-from-secret-to-value-success": {
			init: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.SecretRef = &risingwavev1alpha1.RisingWaveSecretStorePrivateKeySecretReference{
					Name: "test",
					Key:  "test",
				}
			},
			patch: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.SecretStore.PrivateKey.Value = ptr.To("123")
			},
			pass: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {

			// We want to create two copies, so we can compare the old state and new state
			// when transitioning from openkruise enabled to disabled with operator disabled.
			risingwave := testutils.FakeRisingWave()
			if tc.init != nil {
				tc.init(risingwave)
			}

			oldObj := risingwave.DeepCopy()
			if tc.oldObjMutation != nil {
				tc.oldObjMutation(oldObj)
			}
			tc.patch(risingwave)

			webhook := NewRisingWaveValidatingWebhook(tc.openKruiseAvailable)

			// test when operator is disabled and openKruise enabled -> disabled, risingwave should be set to default.
			if name == "disabled-openKruise-when-not-available" {
				if risingwave.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy.Type == oldObj.Spec.Components.Meta.NodeGroups[0].UpgradeStrategy.Type {
					t.Fatal("Risingwave is not default")
				}
			}
			_, err := webhook.ValidateUpdate(context.Background(), oldObj, risingwave)
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
			mutate: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Frontend.NodeGroups[0].Replicas++
			},
			pass: true,
		},
		"scale-views-on-frontend": {
			mutate: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Frontend.NodeGroups[0].Replicas++
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
			mutate: func(r *risingwavev1alpha1.RisingWave) {
				r.Spec.Components.Frontend.NodeGroups[0].Replicas++
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
			origin: testutils.FakeRisingWave(),
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Components.Frontend.NodeGroups = nil
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentFrontend,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     testutils.GetNodeGroupName(0),
							Replicas: 0,
						},
					},
				})
			},
			pass: false,
		},
		"multiple-locked-groups": {
			origin: testutils.FakeRisingWave(),
			mutate: func(wave *risingwavev1alpha1.RisingWave) {
				wave.Spec.Components.Frontend.NodeGroups[0].Replicas = 2
			},
			statusPatch: func(status *risingwavev1alpha1.RisingWaveStatus) {
				status.ScaleViews = append(status.ScaleViews, risingwavev1alpha1.RisingWaveScaleViewLock{
					Name:       "x",
					UID:        "1",
					Component:  consts.ComponentFrontend,
					Generation: 1,
					GroupLocks: []risingwavev1alpha1.RisingWaveScaleViewLockGroupLock{
						{
							Name:     testutils.GetNodeGroupName(0),
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
							Name:     testutils.GetNodeGroupName(0),
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

			_, err := NewRisingWaveValidatingWebhook(false).ValidateUpdate(context.Background(), obj, newObj)
			if tc.pass {
				assert.Nil(t, err, "unexpected error")
			} else {
				assert.NotNil(t, err, "should be nil")
			}
		})
	}
}
