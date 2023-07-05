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
	"fmt"
	"reflect"

	"github.com/samber/lo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
)

// ConvertFrontendService converts v1alpha1 service type and service meta.
// nolint
func ConvertFrontendService(obj *v1alpha1.RisingWave) {
	if obj.Spec.Global.ServiceType != "" && obj.Spec.Global.ServiceType != corev1.ServiceTypeClusterIP {
		obj.Spec.FrontendServiceType = obj.Spec.Global.ServiceType
	}

	if !equality.Semantic.DeepEqual(obj.Spec.Global.ServiceMeta, v1alpha1.PartialObjectMeta{}) {
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
// nolint
func ConvertGlobalImage(obj *v1alpha1.RisingWave) {
	if obj.Spec.Global.Image != "" {
		obj.Spec.Image = obj.Spec.Global.Image
	}
}

// ConvertStorages converts v1alpha1 storages.
// nolint
func ConvertStorages(obj *v1alpha1.RisingWave) {
	if !equality.Semantic.DeepEqual(obj.Spec.Storages, v1alpha1.RisingWaveStoragesSpec{}) &&
		!equality.Semantic.DeepEqual(obj.Spec.Storages, v1alpha1.RisingWaveStoragesSpec{
			Object: v1alpha1.RisingWaveStateStoreBackend{
				DataDirectory: "hummock",
			},
		}) {
		obj.Spec.MetaStore = *obj.Spec.Storages.Meta.DeepCopy()
		obj.Spec.StateStore = *obj.Spec.Storages.Object.DeepCopy()
	}

	metaStore := &obj.Spec.MetaStore
	if metaStore.Etcd != nil && metaStore.Etcd.Secret != "" {
		if metaStore.Etcd.RisingWaveEtcdCredentials == nil {
			metaStore.Etcd.RisingWaveEtcdCredentials = &v1alpha1.RisingWaveEtcdCredentials{}
		}
		metaStore.Etcd.RisingWaveEtcdCredentials.SecretName = metaStore.Etcd.Secret
	}

	stateStore := &obj.Spec.StateStore
	switch {
	case stateStore.MinIO != nil && stateStore.MinIO.Secret != "":
		stateStore.MinIO.RisingWaveMinIOCredentials.SecretName = stateStore.MinIO.Secret
	case stateStore.S3 != nil && stateStore.S3.Secret != "":
		stateStore.S3.RisingWaveS3Credentials.SecretName = stateStore.S3.Secret
	case stateStore.GCS != nil && stateStore.GCS.Secret != "":
		stateStore.GCS.RisingWaveGCSCredentials.SecretName = stateStore.GCS.Secret
	case stateStore.AliyunOSS != nil && stateStore.AliyunOSS.Secret != "":
		stateStore.AliyunOSS.RisingWaveAzureBlobCredentials.SecretName = stateStore.AliyunOSS.Secret
	case stateStore.AzureBlob != nil && stateStore.AzureBlob.Secret != "":
		stateStore.AzureBlob.RisingWaveAzureBlobCredentials.SecretName = stateStore.AzureBlob.Secret
	default:
	}
}

func setDefaultValueForField(field reflect.StructField, target, base reflect.Value) {
	// Only consider primitive types and Map, Slice, and Ptr of these types.
	switch field.Type.Kind() {
	case reflect.Map, reflect.Slice:
		if target.IsZero() || target.Len() == 0 {
			target.Set(base)
		}
	case reflect.Ptr:
		if target.IsZero() {
			target.Set(base)
		}
	default:
		if target.IsZero() {
			target.Set(base)
		}
	}
}

func setDefaultValueForFirstLevelFields[T any](target, base *T) {
	if target == nil || base == nil {
		return
	}

	tType := reflect.TypeOf(target).Elem()

	if tType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("type %s isn't a struct", tType.Name()))
	}

	targetValue, baseValue := reflect.ValueOf(target).Elem(), reflect.ValueOf(base).Elem()

	// Iterate over the fields and set the default values.
	for i := 0; i < tType.NumField(); i++ {
		setDefaultValueForField(tType.Field(i), targetValue.Field(i), baseValue.Field(i))
	}
}

func mergeComponentGroupTemplates(base, overlay *v1alpha1.RisingWaveComponentGroupTemplate) *v1alpha1.RisingWaveComponentGroupTemplate {
	if overlay == nil {
		return base
	}

	r := overlay.DeepCopy()
	setDefaultValueForFirstLevelFields(r, base.DeepCopy())
	return r
}

func convertComponentGroupToNodeGroup(globalTemplate *v1alpha1.RisingWaveComponentGroupTemplate, componentGroup v1alpha1.RisingWaveComponentGroup, restartAt *metav1.Time) v1alpha1.RisingWaveNodeGroup {
	template := mergeComponentGroupTemplates(globalTemplate, componentGroup.RisingWaveComponentGroupTemplate)

	nodeGroup := v1alpha1.RisingWaveNodeGroup{
		Name:            componentGroup.Name,
		Replicas:        componentGroup.Replicas,
		RestartAt:       restartAt,
		Configuration:   nil,
		UpgradeStrategy: template.UpgradeStrategy,
		Template: v1alpha1.RisingWaveNodePodTemplate{
			ObjectMeta: template.Metadata,
			Spec: v1alpha1.RisingWaveNodePodTemplateSpec{
				RisingWaveNodeContainer: v1alpha1.RisingWaveNodeContainer{
					Image:           template.Image,
					EnvFrom:         template.EnvFrom,
					Env:             template.Env,
					Resources:       template.Resources,
					ImagePullPolicy: template.ImagePullPolicy,
				},
				TerminationGracePeriodSeconds: template.TerminationGracePeriodSeconds,
				NodeSelector:                  template.NodeSelector,
				SecurityContext:               template.SecurityContext,
				ImagePullSecrets: lo.Map(template.ImagePullSecrets, func(s string, _ int) corev1.LocalObjectReference {
					return corev1.LocalObjectReference{
						Name: s,
					}
				}),
				Affinity:          template.Affinity,
				Tolerations:       template.Tolerations,
				PriorityClassName: template.PriorityClassName,
				DNSConfig:         template.DNSConfig,
			},
		},
	}

	return nodeGroup
}

// ConvertNodeGroups converts the old node groups to the new node groups.
// nolint
func ConvertNodeGroups(obj *v1alpha1.RisingWave) {
	components := &obj.Spec.Components

	if components.Meta.Groups != nil || obj.Spec.Global.Replicas.Meta > 0 {
		components.Meta.NodeGroups = lo.Map(components.Meta.Groups, func(ng v1alpha1.RisingWaveComponentGroup, _ int) v1alpha1.RisingWaveNodeGroup {
			return convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, ng, components.Meta.RestartAt)
		})
		if obj.Spec.Global.Replicas.Meta > 0 {
			components.Meta.NodeGroups = append(components.Meta.NodeGroups, convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             "",
				Replicas:                         obj.Spec.Global.Replicas.Meta,
				RisingWaveComponentGroupTemplate: nil,
			}, nil))
		}
	}

	if components.Frontend.Groups != nil || obj.Spec.Global.Replicas.Frontend > 0 {
		components.Frontend.NodeGroups = lo.Map(components.Frontend.Groups, func(ng v1alpha1.RisingWaveComponentGroup, _ int) v1alpha1.RisingWaveNodeGroup {
			return convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, ng, components.Frontend.RestartAt)
		})
		if obj.Spec.Global.Replicas.Frontend > 0 {
			components.Frontend.NodeGroups = append(components.Frontend.NodeGroups, convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             "",
				Replicas:                         obj.Spec.Global.Replicas.Frontend,
				RisingWaveComponentGroupTemplate: nil,
			}, nil))
		}
	}

	if components.Compute.Groups != nil || obj.Spec.Global.Replicas.Compute > 0 {
		components.Compute.NodeGroups = lo.Map(components.Compute.Groups, func(ng v1alpha1.RisingWaveComputeGroup, _ int) v1alpha1.RisingWaveNodeGroup {
			var groupTemplate *v1alpha1.RisingWaveComponentGroupTemplate
			if ng.RisingWaveComputeGroupTemplate != nil {
				groupTemplate = &ng.RisingWaveComponentGroupTemplate
			}
			nodeGroup := convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             ng.Name,
				Replicas:                         ng.Replicas,
				RisingWaveComponentGroupTemplate: groupTemplate,
			}, components.Compute.RestartAt)
			nodeGroup.VolumeClaimTemplates = obj.Spec.Storages.PVCTemplates

			if ng.RisingWaveComputeGroupTemplate != nil {
				nodeGroup.Template.Spec.VolumeMounts = ng.VolumeMounts
			}

			return nodeGroup
		})
		if obj.Spec.Global.Replicas.Compute > 0 {
			components.Compute.NodeGroups = append(components.Compute.NodeGroups, convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             "",
				Replicas:                         obj.Spec.Global.Replicas.Compute,
				RisingWaveComponentGroupTemplate: nil,
			}, nil))
		}
	}

	if components.Compactor.Groups != nil || obj.Spec.Global.Replicas.Compactor > 0 {
		components.Compactor.NodeGroups = lo.Map(components.Compactor.Groups, func(ng v1alpha1.RisingWaveComponentGroup, _ int) v1alpha1.RisingWaveNodeGroup {
			return convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, ng, components.Compactor.RestartAt)
		})
		if obj.Spec.Global.Replicas.Compactor > 0 {
			components.Compactor.NodeGroups = append(components.Compactor.NodeGroups, convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             "",
				Replicas:                         obj.Spec.Global.Replicas.Compactor,
				RisingWaveComponentGroupTemplate: nil,
			}, nil))
		}
	}

	if components.Connector.Groups != nil || obj.Spec.Global.Replicas.Connector > 0 {
		components.Connector.NodeGroups = lo.Map(components.Connector.Groups, func(ng v1alpha1.RisingWaveComponentGroup, _ int) v1alpha1.RisingWaveNodeGroup {
			return convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, ng, components.Connector.RestartAt)
		})
		if obj.Spec.Global.Replicas.Connector > 0 {
			components.Connector.NodeGroups = append(components.Connector.NodeGroups, convertComponentGroupToNodeGroup(&obj.Spec.Global.RisingWaveComponentGroupTemplate, v1alpha1.RisingWaveComponentGroup{
				Name:                             "",
				Replicas:                         obj.Spec.Global.Replicas.Connector,
				RisingWaveComponentGroupTemplate: nil,
			}, nil))
		}
	}
}

// ConvertGlobalConfig converts the global config to the new global config.
func ConvertGlobalConfig(obj *v1alpha1.RisingWave) {
	if obj.Spec.Configuration.ConfigMap != nil {
		obj.Spec.Configuration.RisingWaveNodeConfiguration.ConfigMap = &v1alpha1.RisingWaveNodeConfigurationConfigMapSource{
			Name:     obj.Spec.Configuration.ConfigMap.Name,
			Key:      obj.Spec.Configuration.ConfigMap.Key,
			Optional: obj.Spec.Configuration.ConfigMap.Optional,
		}
	}
}

// ConvertToV1alpha2Features converts v1alpha1 features to v1alpha2 features.
func ConvertToV1alpha2Features(obj *v1alpha1.RisingWave) {
	ConvertFrontendService(obj)
	ConvertGlobalImage(obj)
	ConvertStorages(obj)
	ConvertNodeGroups(obj)
	ConvertGlobalConfig(obj)
}
