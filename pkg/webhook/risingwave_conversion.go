// Copyright 2023 RisingWave Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// ConvertFrontendService converts v1alpha1 service type and service meta.
func ConvertFrontendService(obj *v1alpha1.RisingWave) {
	if obj.Spec.Global.ServiceType != corev1.ServiceTypeClusterIP {
		obj.Spec.FrontendServiceType = obj.Spec.Global.ServiceType
	}

	if !equality.Semantic.DeepEqual(obj.Spec.Global.ServiceMeta, v1alpha1.RisingWavePodTemplatePartialObjectMeta{}) {
		obj.Spec.AdditionalFrontendServiceMetadata.Labels = make(map[string]string)
		obj.Spec.AdditionalFrontendServiceMetadata.Annotations = make(map[string]string)

		for key, value := range obj.Spec.Global.ServiceMeta.Labels {
			obj.Spec.AdditionalFrontendServiceMetadata.Labels[key] = value
		}
		for key, value := range obj.Spec.Global.ServiceMeta.Annotations {
			obj.Spec.AdditionalFrontendServiceMetadata.Annotations[key] = value
		}
	}
}

// ConvertGlobalImage converts v1alpha1 global image.
func ConvertGlobalImage(obj *v1alpha1.RisingWave) {
	if obj.Spec.Global.Image != "" {
		obj.Spec.Image = obj.Spec.Global.Image
	}
}

// ConvertStorages converts v1alpha1 storages.
func ConvertStorages(obj *v1alpha1.RisingWave) {
	if !equality.Semantic.DeepEqual(obj.Spec.Storages, v1alpha1.RisingWaveStoragesSpec{}) {
		obj.Spec.MetaStore = *obj.Spec.Storages.Meta.DeepCopy()
		obj.Spec.StateStore = *obj.Spec.Storages.Object.DeepCopy()
	}

	metaStorage := &obj.Spec.MetaStore
	if metaStorage.Etcd != nil && metaStorage.Etcd.Secret != "" {
		if metaStorage.Etcd.RisingWaveEtcdCredentials == nil {
			metaStorage.Etcd.RisingWaveEtcdCredentials = &v1alpha1.RisingWaveEtcdCredentials{}
		}
		metaStorage.Etcd.RisingWaveEtcdCredentials.SecretName = metaStorage.Etcd.Secret
	}

	stateStorage := &obj.Spec.StateStore
	switch {
	case stateStorage.MinIO != nil && stateStorage.MinIO.Secret != "":
		stateStorage.MinIO.RisingWaveMinIOCredentials.SecretName = stateStorage.MinIO.Secret
	case stateStorage.S3 != nil && stateStorage.S3.Secret != "":
		stateStorage.S3.RisingWaveS3Credentials.SecretName = stateStorage.S3.Secret
	case stateStorage.GCS != nil && stateStorage.GCS.Secret != "":
		stateStorage.GCS.RisingWaveGCSCredentials.SecretName = stateStorage.GCS.Secret
	case stateStorage.AliyunOSS != nil && stateStorage.AliyunOSS.Secret != "":
		stateStorage.AliyunOSS.RisingWaveS3Credentials.SecretName = stateStorage.AliyunOSS.Secret
	case stateStorage.AzureBlob != nil && stateStorage.AzureBlob.Secret != "":
		stateStorage.AzureBlob.RisingWaveAzureBlobCredentials.SecretName = stateStorage.AzureBlob.Secret
	default:
	}
}

// ConvertToV1alpha2Features converts v1alpha1 features to v1alpha2 features.
func ConvertToV1alpha2Features(obj *v1alpha1.RisingWave) {
	ConvertFrontendService(obj)
	ConvertGlobalImage(obj)
	ConvertStorages(obj)
}
