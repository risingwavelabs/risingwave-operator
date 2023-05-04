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

package factory

import (
	"fmt"
	"math"
	"path"
	"strconv"
	"strings"
	"time"

	kruisepubs "github.com/openkruise/kruise-api/apps/pub"
	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/consts"
	"github.com/risingwavelabs/risingwave-operator/pkg/factory/envs"
)

const (
	risingWaveConfigVolume = "risingwave-config"
	risingWaveConfigMapKey = "risingwave.toml"

	risingwaveExecutablePath  = "/risingwave/bin/risingwave"
	risingwaveConfigMountPath = "/risingwave/config"
	risingwaveConfigFileName  = "risingwave.toml"
)

var (
	aliyunOSSEndpoint         = fmt.Sprintf("https://$(%s).oss-$(%s).aliyuncs.com", envs.S3CompatibleBucket, envs.S3CompatibleRegion)
	internalAliyunOSSEndpoint = fmt.Sprintf("https://$(%s).oss-$(%s)-internal.aliyuncs.com", envs.S3CompatibleBucket, envs.S3CompatibleRegion)
)

// RisingWaveObjectFactory is the object factory to help create owned objects like Deployments, StatefulSets, Services, etc.
type RisingWaveObjectFactory struct {
	scheme     *runtime.Scheme
	risingwave *risingwavev1alpha1.RisingWave

	inheritedLabels map[string]string
	operatorVersion string
}

func mustSetControllerReference[T client.Object](owner client.Object, controlled T, scheme *runtime.Scheme) T {
	err := ctrl.SetControllerReference(owner, controlled, scheme)
	if err != nil {
		panic(err)
	}
	return controlled
}

func (f *RisingWaveObjectFactory) namespace() string {
	return f.risingwave.Namespace
}

func (f *RisingWaveObjectFactory) isStateStoreS3() bool {
	return f.risingwave.Spec.StateStore.S3 != nil && len(f.risingwave.Spec.StateStore.S3.Endpoint) == 0
}

func (f *RisingWaveObjectFactory) isStateStoreS3Compatible() bool {
	return f.risingwave.Spec.StateStore.S3 != nil && len(f.risingwave.Spec.StateStore.S3.Endpoint) > 0
}

func (f *RisingWaveObjectFactory) isStateStoreGCS() bool {
	return f.risingwave.Spec.StateStore.GCS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreAliyunOSS() bool {
	return f.risingwave.Spec.StateStore.AliyunOSS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreAzureBlob() bool {
	return f.risingwave.Spec.StateStore.AzureBlob != nil
}

func (f *RisingWaveObjectFactory) isStateStoreHDFS() bool {
	return f.risingwave.Spec.StateStore.HDFS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreWebHDFS() bool {
	return f.risingwave.Spec.StateStore.WebHDFS != nil
}

func (f *RisingWaveObjectFactory) isStateStoreMinIO() bool {
	return f.risingwave.Spec.StateStore.MinIO != nil
}

func (f *RisingWaveObjectFactory) isMetaStoreEtcd() bool {
	return f.risingwave.Spec.MetaStore.Etcd != nil
}

func (f *RisingWaveObjectFactory) hummockConnectionStr() string {
	stateStore := f.risingwave.Spec.StateStore
	switch {
	case pointer.BoolDeref(stateStore.Memory, false):
		return "hummock+memory"
	case f.isStateStoreS3():
		bucket := stateStore.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case f.isStateStoreS3Compatible():
		bucket := stateStore.S3.Bucket
		return fmt.Sprintf("hummock+s3-compatible://%s", bucket)
	case stateStore.MinIO != nil:
		minio := stateStore.MinIO
		return fmt.Sprintf("hummock+minio://$(%s):$(%s)@%s/%s", envs.MinIOUsername, envs.MinIOPassword, minio.Endpoint, minio.Bucket)
	case f.isStateStoreGCS():
		return fmt.Sprintf("hummock+gcs://%s@%s", stateStore.GCS.Bucket, stateStore.GCS.Root)
	case stateStore.AliyunOSS != nil:
		aliyunOSS := stateStore.AliyunOSS
		// Redirect to s3-compatible.
		return fmt.Sprintf("hummock+s3-compatible://%s", aliyunOSS.Bucket)
	case stateStore.AzureBlob != nil:
		azureBlob := stateStore.AzureBlob
		// Redirect to s3-compatible.
		return fmt.Sprintf("hummock+azblob://%s@%s", azureBlob.Container, azureBlob.Root)
	case stateStore.HDFS != nil:
		hdfs := stateStore.HDFS
		return fmt.Sprintf("hummock+hdfs://%s@%s", hdfs.NameNode, hdfs.Root)
	case stateStore.WebHDFS != nil:
		webhdfs := stateStore.WebHDFS
		return fmt.Sprintf("hummock+webhdfs://%s@%s", webhdfs.NameNode, webhdfs.Root)
	default:
		panic("unrecognized storage type")
	}
}

func groupSuffix(group string) string {
	if group == "" {
		return ""
	}
	return "-" + group
}

func (f *RisingWaveObjectFactory) componentName(component, group string) string {
	switch component {
	case consts.ComponentMeta:
		return f.risingwave.Name + "-meta" + groupSuffix(group)
	case consts.ComponentCompute:
		return f.risingwave.Name + "-compute" + groupSuffix(group)
	case consts.ComponentFrontend:
		return f.risingwave.Name + "-frontend" + groupSuffix(group)
	case consts.ComponentCompactor:
		return f.risingwave.Name + "-compactor" + groupSuffix(group)
	case consts.ComponentConnector:
		return f.risingwave.Name + "-connector" + groupSuffix(group)
	case consts.ComponentConfig:
		return f.risingwave.Name + "-config"
	default:
		panic("never reach here")
	}
}

func (f *RisingWaveObjectFactory) objectMeta(name string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      name,
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	return objectMeta
}

func (f *RisingWaveObjectFactory) componentObjectMeta(component string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      f.componentName(component, ""),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveComponent:       component,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	if component == consts.ComponentFrontend {
		objectMeta.Labels = mergeMap(objectMeta.Labels, f.risingwave.Spec.AdditionalFrontendServiceMetadata.Labels)
		objectMeta.Annotations = mergeMap(objectMeta.Annotations, f.risingwave.Spec.AdditionalFrontendServiceMetadata.Annotations)
	}

	return objectMeta
}

func (f *RisingWaveObjectFactory) componentGroupObjectMeta(component, group string, sync bool) metav1.ObjectMeta {
	objectMeta := metav1.ObjectMeta{
		Name:      f.componentName(component, group),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:            f.risingwave.Name,
			consts.LabelRisingWaveComponent:       component,
			consts.LabelRisingWaveGeneration:      lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveGroup:           group,
			consts.LabelRisingWaveOperatorVersion: f.operatorVersion,
		},
	}

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.getInheritedLabels())

	objectMeta.Labels = mergeMap(objectMeta.Labels, f.risingwave.Spec.Global.Metadata.Labels)
	objectMeta.Annotations = mergeMap(objectMeta.Annotations, f.risingwave.Spec.Global.Metadata.Annotations)

	return objectMeta
}

func (f *RisingWaveObjectFactory) podLabelsOrSelectors(component string) map[string]string {
	return map[string]string{
		consts.LabelRisingWaveName:      f.risingwave.Name,
		consts.LabelRisingWaveComponent: component,
	}
}

func (f *RisingWaveObjectFactory) podLabelsOrSelectorsForGroup(component, group string) map[string]string {
	return map[string]string{
		consts.LabelRisingWaveName:      f.risingwave.Name,
		consts.LabelRisingWaveComponent: component,
		consts.LabelRisingWaveGroup:     group,
	}
}

// NewMetaService creates a new Service for the meta.
func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports

	metaService := &corev1.Service{
		ObjectMeta: f.componentObjectMeta(consts.ComponentMeta, true),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentMeta),
			Ports: []corev1.ServicePort{
				{
					Name:       consts.PortService,
					Protocol:   corev1.ProtocolTCP,
					Port:       metaPorts.ServicePort,
					TargetPort: intstr.FromString(consts.PortService),
				},
				{
					Name:       consts.PortMetrics,
					Protocol:   corev1.ProtocolTCP,
					Port:       metaPorts.MetricsPort,
					TargetPort: intstr.FromString(consts.PortMetrics),
				},
				{
					Name:       consts.PortDashboard,
					Protocol:   corev1.ProtocolTCP,
					Port:       metaPorts.DashboardPort,
					TargetPort: intstr.FromString(consts.PortDashboard),
				},
			},
		},
	}
	return mustSetControllerReference(f.risingwave, metaService, f.scheme)
}

// NewFrontendService creates a new Service for the frontend.
func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	frontendPorts := &f.risingwave.Spec.Components.Frontend.Ports

	frontendService := &corev1.Service{
		ObjectMeta: f.componentObjectMeta(consts.ComponentFrontend, true),
		Spec: corev1.ServiceSpec{
			Type:     f.risingwave.Spec.FrontendServiceType,
			Selector: f.podLabelsOrSelectors(consts.ComponentFrontend),
			Ports: []corev1.ServicePort{
				{
					Name:       consts.PortService,
					Protocol:   corev1.ProtocolTCP,
					Port:       frontendPorts.ServicePort,
					TargetPort: intstr.FromString(consts.PortService),
				},
				{
					Name:       consts.PortMetrics,
					Protocol:   corev1.ProtocolTCP,
					Port:       frontendPorts.MetricsPort,
					TargetPort: intstr.FromString(consts.PortMetrics),
				},
			},
		},
	}
	return mustSetControllerReference(f.risingwave, frontendService, f.scheme)
}

// NewComputeService creates a new Service for the compute nodes.
func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	computePorts := &f.risingwave.Spec.Components.Compute.Ports

	computeService := &corev1.Service{
		ObjectMeta: f.componentObjectMeta(consts.ComponentCompute, true),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentCompute),
			Ports: []corev1.ServicePort{
				{
					Name:       consts.PortService,
					Protocol:   corev1.ProtocolTCP,
					Port:       computePorts.ServicePort,
					TargetPort: intstr.FromString(consts.PortService),
				},
				{
					Name:       consts.PortMetrics,
					Protocol:   corev1.ProtocolTCP,
					Port:       computePorts.MetricsPort,
					TargetPort: intstr.FromString(consts.PortMetrics),
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, computeService, f.scheme)
}

// NewCompactorService creates a new Service for the compactor.
func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	compactorPorts := &f.risingwave.Spec.Components.Compactor.Ports

	compactorService := &corev1.Service{
		ObjectMeta: f.componentObjectMeta(consts.ComponentCompactor, true),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentCompactor),
			Ports: []corev1.ServicePort{
				{
					Name:       consts.PortService,
					Protocol:   corev1.ProtocolTCP,
					Port:       compactorPorts.ServicePort,
					TargetPort: intstr.FromString(consts.PortService),
				},
				{
					Name:       consts.PortMetrics,
					Protocol:   corev1.ProtocolTCP,
					Port:       compactorPorts.MetricsPort,
					TargetPort: intstr.FromString(consts.PortMetrics),
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, compactorService, f.scheme)
}

// NewConnectorService creates a new Service for the connector.
func (f *RisingWaveObjectFactory) NewConnectorService() *corev1.Service {
	connectorPorts := f.getConnectorPorts()

	connectorService := &corev1.Service{
		ObjectMeta: f.componentObjectMeta(consts.ComponentConnector, true),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentConnector),
			Ports: []corev1.ServicePort{
				{
					Name:       consts.PortService,
					Protocol:   corev1.ProtocolTCP,
					Port:       connectorPorts.ServicePort,
					TargetPort: intstr.FromString(consts.PortService),
				},
				{
					Name:       consts.PortMetrics,
					Protocol:   corev1.ProtocolTCP,
					Port:       connectorPorts.MetricsPort,
					TargetPort: intstr.FromString(consts.PortMetrics),
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, connectorService, f.scheme)
}

func (f *RisingWaveObjectFactory) envsForEtcd() []corev1.EnvVar {
	credentials := f.risingwave.Spec.MetaStore.Etcd.RisingWaveEtcdCredentials

	// Empty secret indicates no authentication.
	if credentials == nil || credentials.SecretName == "" {
		return []corev1.EnvVar{}
	}

	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	return []corev1.EnvVar{
		// Keep the legacy environment variables for compatibility. Will remove them later.
		{
			Name: envs.EtcdUsernameLegacy,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.EtcdPasswordLegacy,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
		// Environment variables for etcd auth.
		{
			Name: envs.RWEtcdUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.RWEtcdPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForMetaArgs() []corev1.EnvVar {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	metaStore := &f.risingwave.Spec.MetaStore
	stateStore := f.risingwave.Spec.StateStore

	connectorPorts := f.getConnectorPorts()
	envVars := []corev1.EnvVar{
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", metaPorts.ServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_NAME).%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWStateStore,
			Value: f.hummockConnectionStr(),
		},
		{
			Name:  envs.RWDataDirectory,
			Value: stateStore.DataDirectory,
		},
		{
			Name:  envs.RWDashboardHost,
			Value: fmt.Sprintf("0.0.0.0:%d", metaPorts.DashboardPort),
		},
		{
			Name:  envs.RWPrometheusHost,
			Value: fmt.Sprintf("0.0.0.0:%d", metaPorts.MetricsPort),
		},
		{
			Name:  envs.RWConnectorRPCEndPoint,
			Value: fmt.Sprintf("%s:%d", f.componentName(consts.ComponentConnector, ""), connectorPorts.ServicePort),
		},
	}

	switch {
	case pointer.BoolDeref(metaStore.Memory, false):
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWBackend,
			Value: "mem",
		})
	case metaStore.Etcd != nil:
		envVars = append(envVars, []corev1.EnvVar{
			{
				Name:  envs.RWBackend,
				Value: "etcd",
			},
			{
				Name:  envs.RWEtcdEndpoints,
				Value: metaStore.Etcd.Endpoint,
			},
		}...)
		credentials := f.risingwave.Spec.MetaStore.Etcd.RisingWaveEtcdCredentials
		if credentials != nil && credentials.SecretName != "" {
			envVars = append(envVars, corev1.EnvVar{
				Name:  envs.RWEtcdAuth,
				Value: "true",
			})
		}
	default:
		panic("unsupported meta storage type")
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForFrontendArgs() []corev1.EnvVar {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	frontendPorts := &f.risingwave.Spec.Components.Frontend.Ports

	return []corev1.EnvVar{
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", frontendPorts.ServicePort),
		},
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_IP):%d", frontendPorts.ServicePort),
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetricsLevel,
			Value: "1",
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", frontendPorts.MetricsPort),
		},
	}
}

func (f *RisingWaveObjectFactory) envsForComputeArgs(cpuLimit int64, memLimit int64) []corev1.EnvVar {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	computePorts := &f.risingwave.Spec.Components.Compute.Ports
	connectorPorts := f.getConnectorPorts()

	envVars := []corev1.EnvVar{
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", computePorts.ServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_NAME).%s:%d", f.componentName(consts.ComponentCompute, ""), computePorts.ServicePort),
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetricsLevel,
			Value: "1",
		},
		{
			Name:  envs.RWConnectorRPCEndPoint,
			Value: fmt.Sprintf("%s:%d", f.componentName(consts.ComponentConnector, ""), connectorPorts.ServicePort),
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", computePorts.MetricsPort),
		},
	}

	if cpuLimit != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWParallelism,
			Value: strconv.FormatInt(cpuLimit, 10),
		})
	}

	if memLimit != 0 {
		envVars = append(envVars, corev1.EnvVar{
			Name:  envs.RWTotalMemoryBytes,
			Value: strconv.FormatInt(memLimit, 10),
		})
	}

	return envVars
}

func (f *RisingWaveObjectFactory) envsForCompactorArgs() []corev1.EnvVar {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	compactorPorts := &f.risingwave.Spec.Components.Compactor.Ports

	return []corev1.EnvVar{
		{
			Name:  envs.RWConfigPath,
			Value: path.Join(risingwaveConfigMountPath, risingwaveConfigFileName),
		},
		{
			Name:  envs.RWListenAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", compactorPorts.ServicePort),
		},
		{
			Name:  envs.RWAdvertiseAddr,
			Value: fmt.Sprintf("$(POD_IP):%d", compactorPorts.ServicePort),
		},
		{
			Name:  envs.RWPrometheusListenerAddr,
			Value: fmt.Sprintf("0.0.0.0:%d", compactorPorts.MetricsPort),
		},
		{
			Name:  envs.RWMetaAddr,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetaAddrLegacy,
			Value: fmt.Sprintf("load-balance+http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
		},
		{
			Name:  envs.RWMetricsLevel,
			Value: "1",
		},
	}
}

func (f *RisingWaveObjectFactory) argsForConnector() []string {
	connectorPorts := f.getConnectorPorts()

	return []string{
		"-p", fmt.Sprintf("%d", connectorPorts.ServicePort),
	}
}

func mergeListWhenKeyEquals[T any](list []T, val T, equals func(a, b *T) bool) []T {
	for i, v := range list {
		if equals(&val, &v) {
			list[i] = val
			return list
		}
	}
	return append(list, val)
}

func mergeListByKey[T any](list []T, val T, keyPred func(*T) bool) []T {
	for i, v := range list {
		if keyPred(&v) {
			list[i] = val
			return list
		}
	}
	return append(list, val)
}

func (f *RisingWaveObjectFactory) envsForMinIO() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	credentials := &stateStore.MinIO.RisingWaveMinIOCredentials

	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	return []corev1.EnvVar{
		{
			Name:  envs.MinIOEndpoint,
			Value: stateStore.MinIO.Endpoint,
		},
		{
			Name:  envs.MinIOBucket,
			Value: stateStore.MinIO.Bucket,
		},
		{
			Name: envs.MinIOUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.UsernameKeyRef,
				},
			},
		},
		{
			Name: envs.MinIOPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.PasswordKeyRef,
				},
			},
		},
	}
}

func envsForAWSS3(region, bucket string, credentials risingwavev1alpha1.RisingWaveS3Credentials) []corev1.EnvVar {
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	var regionEnvVar corev1.EnvVar
	if len(region) > 0 {
		regionEnvVar = corev1.EnvVar{
			Name:  envs.AWSRegion,
			Value: region,
		}
	} else {
		regionEnvVar = corev1.EnvVar{
			Name: envs.AWSRegion,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3Region,
				},
			},
		}
	}

	return []corev1.EnvVar{
		{
			Name:  envs.AWSS3Bucket,
			Value: bucket,
		},
		regionEnvVar,
		{
			Name: envs.AWSAccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeyRef,
				},
			},
		},
		{
			Name: envs.AWSSecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.SecretAccessKeyRef,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForS3() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	s3Spec := stateStore.S3

	if len(s3Spec.Endpoint) > 0 {
		// S3 compatible mode.
		endpoint := strings.TrimSpace(s3Spec.Endpoint)

		// Interpret the variables.
		endpoint = strings.ReplaceAll(endpoint, "${REGION}", fmt.Sprintf("$(%s)", envs.S3CompatibleRegion))
		endpoint = strings.ReplaceAll(endpoint, "${BUCKET}", fmt.Sprintf("$(%s)", envs.S3CompatibleBucket))

		if !strings.HasPrefix(endpoint, "https://") {
			endpoint = "https://" + endpoint
		}
		return envsForS3Compatible(s3Spec.Region, endpoint, s3Spec.Bucket, s3Spec.RisingWaveS3Credentials)
	}
	// AWS S3 mode.
	return envsForAWSS3(s3Spec.Region, s3Spec.Bucket, s3Spec.RisingWaveS3Credentials)
}

func envsForS3Compatible(region, endpoint, bucket string, credentials risingwavev1alpha1.RisingWaveS3Credentials) []corev1.EnvVar {
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}

	var regionEnvVar corev1.EnvVar
	if len(region) > 0 {
		regionEnvVar = corev1.EnvVar{
			Name:  envs.S3CompatibleRegion,
			Value: region,
		}
	} else {
		regionEnvVar = corev1.EnvVar{
			Name: envs.S3CompatibleRegion,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3Region,
				},
			},
		}
	}

	return []corev1.EnvVar{
		{
			// Disable auto region loading. Refer to the original source for more information.
			// https://github.com/awslabs/aws-sdk-rust/blob/main/sdk/aws-config/src/imds/region.rs
			// cspell:disable-next-line
			Name:  envs.AWSEC2MetadataDisabled,
			Value: "true",
		},
		{
			Name:  envs.S3CompatibleBucket,
			Value: bucket,
		},
		regionEnvVar,
		{
			Name: envs.S3CompatibleAccessKeyID,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccessKeyRef,
				},
			},
		},
		{
			Name: envs.S3CompatibleSecretAccessKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.SecretAccessKeyRef,
				},
			},
		},
		{
			Name:  envs.S3CompatibleEndpoint,
			Value: endpoint,
		},
	}
}

func (f *RisingWaveObjectFactory) envsForGCS() []corev1.EnvVar {
	gcs := f.risingwave.Spec.StateStore.GCS
	useWorkloadIdentity := gcs.UseWorkloadIdentity
	if useWorkloadIdentity {
		return []corev1.EnvVar{}
	}

	credentials := gcs.RisingWaveGCSCredentials
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}
	return []corev1.EnvVar{
		{
			Name: envs.GoogleApplicationCredentials,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.ServiceAccountCredentialsKeyRef,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForAliyunOSS() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore

	var endpoint string
	if stateStore.AliyunOSS.InternalEndpoint {
		endpoint = internalAliyunOSSEndpoint
	} else {
		endpoint = aliyunOSSEndpoint
	}

	return envsForS3Compatible(stateStore.AliyunOSS.Region, endpoint, stateStore.AliyunOSS.Bucket, stateStore.AliyunOSS.RisingWaveS3Credentials)
}

func (f *RisingWaveObjectFactory) envsForAzureBlob() []corev1.EnvVar {
	stateStore := &f.risingwave.Spec.StateStore
	credentials := stateStore.AzureBlob.RisingWaveAzureBlobCredentials
	secretRef := corev1.LocalObjectReference{
		Name: credentials.SecretName,
	}
	return []corev1.EnvVar{

		{
			Name: envs.AzureBlobAccountName,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccountNameRef,
				},
			},
		},
		{
			Name: envs.AzureBlobAccountKey,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  credentials.AccountKeyRef,
				},
			},
		},
		{
			Name:  envs.AzureBlobEndpoint,
			Value: stateStore.AzureBlob.Endpoint,
		},
	}

}

func (f *RisingWaveObjectFactory) envsForHDFS() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (f *RisingWaveObjectFactory) envsForWebHDFS() []corev1.EnvVar {
	return []corev1.EnvVar{}
}

func (f *RisingWaveObjectFactory) envsForStateStore() []corev1.EnvVar {
	switch {
	case f.isStateStoreMinIO():
		return f.envsForMinIO()
	case f.isStateStoreS3() || f.isStateStoreS3Compatible():
		return f.envsForS3()
	case f.isStateStoreGCS():
		return f.envsForGCS()
	case f.isStateStoreAliyunOSS():
		return f.envsForAliyunOSS()
	case f.isStateStoreAzureBlob():
		return f.envsForAzureBlob()
	case f.isStateStoreHDFS():
		return f.envsForHDFS()
	case f.isStateStoreWebHDFS():
		return f.envsForWebHDFS()
	default:
		return nil
	}
}

func (f *RisingWaveObjectFactory) risingWaveConfigVolume() corev1.Volume {
	return corev1.Volume{
		Name: risingWaveConfigVolume,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				Items: []corev1.KeyToPath{
					{
						Key:  "risingwave.toml",
						Path: risingwaveConfigFileName,
					},
				},
				LocalObjectReference: corev1.LocalObjectReference{
					Name: f.componentName(consts.ComponentConfig, ""),
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) volumeMountForConfig() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      risingWaveConfigVolume,
		MountPath: risingwaveConfigMountPath,
		ReadOnly:  true,
	}
}

// NewConfigConfigMap creates a new ConfigMap with the specified string value for risingwave.toml.
func (f *RisingWaveObjectFactory) NewConfigConfigMap(val string) *corev1.ConfigMap {
	risingwaveConfigConfigMap := &corev1.ConfigMap{
		ObjectMeta: f.componentObjectMeta(consts.ComponentConfig, false), // not synced
		Data: map[string]string{
			risingWaveConfigMapKey: nonZeroOrDefault(val, ""),
		},
	}
	return mustSetControllerReference(f.risingwave, risingwaveConfigConfigMap, f.scheme)
}

func findTheFirstMatchPtr[T any](list []T, predicate func(*T) bool) *T {
	for _, v := range list {
		if predicate(&v) {
			return &v
		}
	}
	return nil
}

func mergeMap[K comparable, V any](a, b map[K]V) map[K]V {
	if a == nil && b == nil {
		return nil
	}

	r := make(map[K]V)
	for k, v := range a {
		r[k] = v
	}
	for k, v := range b {
		r[k] = v
	}

	return r
}

func mergeComponentGroupTemplates(base, overlay *risingwavev1alpha1.RisingWaveComponentGroupTemplate) *risingwavev1alpha1.RisingWaveComponentGroupTemplate {
	if overlay == nil {
		return base
	}

	r := overlay.DeepCopy()
	setDefaultValueForFirstLevelFields(r, base.DeepCopy())
	return r
}

func buildPodTemplateSpecFrom(t *risingwavev1alpha1.RisingWavePodTemplateSpec) corev1.PodTemplateSpec {
	t = t.DeepCopy()
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:      t.Labels,
			Annotations: t.Annotations,
		},
		Spec: t.Spec,
	}
}

func captureInheritedLabels(risingwave *risingwavev1alpha1.RisingWave) map[string]string {
	inheritLabelPrefix, exist := risingwave.Annotations[consts.AnnotationInheritLabelPrefix]
	if !exist {
		return nil
	}

	// Parse the label prefixes (separated by comma) from the annotation value.
	prefixes := strings.Split(inheritLabelPrefix, ",")
	for i, prefix := range prefixes {
		prefixes[i] = strings.TrimSpace(prefix)
	}
	prefixes = lo.Filter(prefixes, func(s string, _ int) bool {
		return len(s) > 0 && s != "risingwave"
	})

	if len(prefixes) == 0 {
		return nil
	}

	// Match labels with naive algorithm here.
	matchLabelKey := func(s string) bool {
		for _, p := range prefixes {
			if strings.HasPrefix(s, p+"/") {
				return true
			}
		}
		return false
	}

	inheritedLabels := make(map[string]string)
	for k, v := range risingwave.Labels {
		if matchLabelKey(k) {
			inheritedLabels[k] = v
		}
	}

	if len(inheritedLabels) == 0 {
		return nil
	}

	return inheritedLabels
}

func (f *RisingWaveObjectFactory) getInheritedLabels() map[string]string {
	return f.inheritedLabels
}

func (f *RisingWaveObjectFactory) buildPodTemplate(component, group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate,
	groupTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate, restartAt *metav1.Time) corev1.PodTemplateSpec {
	var podTemplate corev1.PodTemplateSpec

	if groupTemplate != nil {
		if groupTemplate.PodTemplate != nil && *groupTemplate.PodTemplate != "" {
			t := podTemplates[*groupTemplate.PodTemplate]
			podTemplate = buildPodTemplateSpecFrom(&t.Template)
		}

		// Set the image pull secrets.
		podTemplate.Spec.ImagePullSecrets = append(podTemplate.Spec.ImagePullSecrets, lo.Map(groupTemplate.ImagePullSecrets, func(s string, _ int) corev1.LocalObjectReference {
			return corev1.LocalObjectReference{
				Name: s,
			}
		})...)

		// Set the node selector.
		podTemplate.Spec.NodeSelector = groupTemplate.NodeSelector

		// Set the tolerations.
		podTemplate.Spec.Tolerations = append(podTemplate.Spec.Tolerations, groupTemplate.Tolerations...)

		// Set the PriorityClassName.
		podTemplate.Spec.PriorityClassName = groupTemplate.PriorityClassName

		// Set the security context.
		podTemplate.Spec.SecurityContext = groupTemplate.SecurityContext.DeepCopy()

		// Set the dns config.
		podTemplate.Spec.DNSConfig = groupTemplate.DNSConfig.DeepCopy()

		// Set the termination grace period seconds.
		if groupTemplate.TerminationGracePeriodSeconds != nil {
			podTemplate.Spec.TerminationGracePeriodSeconds = pointer.Int64(*groupTemplate.TerminationGracePeriodSeconds)
		}

		// Set the affinity.
		podTemplate.Spec.Affinity = groupTemplate.Affinity.DeepCopy()
	}

	// Set config volume.
	podTemplate.Spec.Volumes = mergeListWhenKeyEquals(podTemplate.Spec.Volumes, f.risingWaveConfigVolume(), func(a, b *corev1.Volume) bool {
		return a.Name == b.Name
	})

	// Set labels and annotations.
	podTemplate.Labels = mergeMap(podTemplate.Labels, f.podLabelsOrSelectorsForGroup(component, group))

	// Inherit labels from RisingWave, according to the hint.
	podTemplate.Labels = mergeMap(podTemplate.Labels, f.getInheritedLabels())

	if restartAt != nil {
		if podTemplate.Annotations == nil {
			podTemplate.Annotations = make(map[string]string)
		}
		podTemplate.Annotations[consts.AnnotationRestartAt] = restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z")
	}

	// Set up the first container.
	if len(podTemplate.Spec.Containers) == 0 {
		podTemplate.Spec.Containers = append(podTemplate.Spec.Containers, corev1.Container{})
	}

	// By default, disable the service links.
	if podTemplate.Spec.EnableServiceLinks == nil {
		podTemplate.Spec.EnableServiceLinks = pointer.Bool(false)
	}

	return podTemplate
}

func buildComponentGroup(globalReplicas int32, globalTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate,
	group string, groups []risingwavev1alpha1.RisingWaveComponentGroup) *risingwavev1alpha1.RisingWaveComponentGroup {
	var componentGroup *risingwavev1alpha1.RisingWaveComponentGroup
	if group == "" {
		componentGroup = &risingwavev1alpha1.RisingWaveComponentGroup{
			Replicas:                         globalReplicas,
			RisingWaveComponentGroupTemplate: globalTemplate.DeepCopy(),
		}
	} else {
		componentGroup = findTheFirstMatchPtr(groups, func(g *risingwavev1alpha1.RisingWaveComponentGroup) bool {
			return g.Name == group
		}).DeepCopy()

		if componentGroup == nil {
			return nil
		}
		componentGroup.RisingWaveComponentGroupTemplate = mergeComponentGroupTemplates(globalTemplate, componentGroup.RisingWaveComponentGroupTemplate)
	}
	return componentGroup
}

func buildComputeGroup(globalReplicas int32, globalTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate,
	group string, groups []risingwavev1alpha1.RisingWaveComputeGroup) *risingwavev1alpha1.RisingWaveComputeGroup {
	var componentGroup *risingwavev1alpha1.RisingWaveComputeGroup
	if group == "" {
		replicas := globalReplicas
		componentGroup = &risingwavev1alpha1.RisingWaveComputeGroup{
			Replicas: replicas,
			RisingWaveComputeGroupTemplate: &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
				RisingWaveComponentGroupTemplate: *globalTemplate.DeepCopy(),
			},
		}
	} else {
		componentGroup = findTheFirstMatchPtr(groups, func(g *risingwavev1alpha1.RisingWaveComputeGroup) bool {
			return g.Name == group
		}).DeepCopy()
		if componentGroup == nil {
			return nil
		}

		if componentGroup.RisingWaveComputeGroupTemplate != nil {
			componentGroup.RisingWaveComputeGroupTemplate.RisingWaveComponentGroupTemplate = *mergeComponentGroupTemplates(globalTemplate,
				&componentGroup.RisingWaveComputeGroupTemplate.RisingWaveComponentGroupTemplate)
		} else {
			componentGroup.RisingWaveComputeGroupTemplate = &risingwavev1alpha1.RisingWaveComputeGroupTemplate{
				RisingWaveComponentGroupTemplate: *globalTemplate.DeepCopy(),
			}
		}
	}
	return componentGroup
}

func (f *RisingWaveObjectFactory) portsForMetaContainer() []corev1.ContainerPort {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports

	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: metaPorts.ServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: metaPorts.MetricsPort,
		},
		{
			Name:          consts.PortDashboard,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: metaPorts.DashboardPort,
		},
	}
}

func basicSetupContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate, image string) {
	container.Image = image
	container.ImagePullPolicy = template.ImagePullPolicy
	container.Command = []string{risingwaveExecutablePath}

	// Copy the template's envFrom.
	container.EnvFrom = make([]corev1.EnvFromSource, 0, len(template.EnvFrom))
	for _, envFrom := range template.EnvFrom {
		container.EnvFrom = append(container.EnvFrom, *envFrom.DeepCopy())
	}

	// Copy the template's env.
	container.Env = make([]corev1.EnvVar, 0, len(template.Env))
	for _, env := range template.Env {
		container.Env = append(container.Env, *env.DeepCopy())
	}

	// Setting the system environment variables.
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: envs.PodIP,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.PodIP })
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: envs.PodName,
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		},
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.PodName })
	// Set RUST_BACKTRACE=1 by default.
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name:  envs.RustBacktrace,
		Value: "full",
	}, func(env *corev1.EnvVar) bool { return env.Name == envs.RustBacktrace })
	if cpuLimit, ok := template.Resources.Limits[corev1.ResourceCPU]; ok {
		container.Env = mergeListByKey(container.Env, corev1.EnvVar{
			Name:  envs.RWWorkerThreads,
			Value: strconv.FormatInt(cpuLimit.Value(), 10),
		}, func(env *corev1.EnvVar) bool { return env.Name == envs.RWWorkerThreads })
	}
	container.Resources = template.Resources
	container.StartupProbe = nil
	container.LivenessProbe = &corev1.Probe{
		InitialDelaySeconds: 2,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(consts.PortService),
			},
		},
	}
	container.ReadinessProbe = &corev1.Probe{
		InitialDelaySeconds: 2,
		PeriodSeconds:       10,
		ProbeHandler: corev1.ProbeHandler{
			TCPSocket: &corev1.TCPSocketAction{
				Port: intstr.FromString(consts.PortService),
			},
		},
	}
	container.Stdin = false
	container.StdinOnce = false
	container.TTY = false
}

func (f *RisingWaveObjectFactory) setupMetaContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate) {
	basicSetupContainer(container, template, f.risingwave.Spec.Image)

	container.Name = "meta"
	container.Args = []string{"meta-node"}
	container.Ports = f.portsForMetaContainer()
	container.Env = append(container.Env, f.envsForMetaArgs()...)

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}
	if f.isMetaStoreEtcd() {
		for _, env := range f.envsForEtcd() {
			container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
				return a.Name == b.Name
			})
		}
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func rollingUpdateOrDefault(rollingUpdate *risingwavev1alpha1.RisingWaveRollingUpdate) risingwavev1alpha1.RisingWaveRollingUpdate {
	if rollingUpdate != nil {
		return *rollingUpdate
	}
	return risingwavev1alpha1.RisingWaveRollingUpdate{}
}

func buildUpgradeStrategyForDeployment(strategy risingwavev1alpha1.RisingWaveUpgradeStrategy) appsv1.DeploymentStrategy {
	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate:
		return appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
			},
		}
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		return appsv1.DeploymentStrategy{
			Type: appsv1.RecreateDeploymentStrategyType,
		}
	default:
		return appsv1.DeploymentStrategy{}
	}
}

func inPlaceUpdateStrategyOrDefault(strategy *kruisepubs.InPlaceUpdateStrategy) *kruisepubs.InPlaceUpdateStrategy {
	if strategy != nil {
		return strategy
	}
	return &kruisepubs.InPlaceUpdateStrategy{}
}

func buildUpgradeStrategyForCloneSet(strategy risingwavev1alpha1.RisingWaveUpgradeStrategy) kruiseappsv1alpha1.CloneSetUpdateStrategy {
	cloneSetUpdateStrategy := kruiseappsv1alpha1.CloneSetUpdateStrategy{}

	rollingUpdateStrategy := rollingUpdateOrDefault(strategy.RollingUpdate)
	cloneSetUpdateStrategy.MaxUnavailable = rollingUpdateStrategy.MaxUnavailable
	cloneSetUpdateStrategy.MaxSurge = rollingUpdateStrategy.MaxSurge

	cloneSetUpdateStrategy.Partition = rollingUpdateStrategy.Partition
	if strategy.InPlaceUpdateStrategy != nil {
		cloneSetUpdateStrategy.InPlaceUpdateStrategy = inPlaceUpdateStrategyOrDefault(strategy.InPlaceUpdateStrategy)
	}

	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.RecreateCloneSetUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.InPlaceIfPossibleCloneSetUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly:
		cloneSetUpdateStrategy.Type = kruiseappsv1alpha1.InPlaceOnlyCloneSetUpdateStrategyType
	default:
		return kruiseappsv1alpha1.CloneSetUpdateStrategy{}
	}

	return cloneSetUpdateStrategy
}

// NewMetaStatefulSet creates a new StatefulSet for the meta component and specified group.
func (f *RisingWaveObjectFactory) NewMetaStatefulSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.StatefulSet {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Meta,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Meta.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Meta.RestartAt

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentMeta, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupMetaContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the StatefulSet.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentMeta, group)
	metaSts := &appsv1.StatefulSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentMeta, group, true),
		Spec: appsv1.StatefulSetSpec{
			ServiceName:    f.componentName(consts.ComponentMeta, ""),
			Replicas:       pointer.Int32(componentGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForStatefulSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template:            podTemplate,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, metaSts, f.scheme)
}

// NewMetaAdvancedStatefulSet creates a new OpenKruise StatefulSet for the meta component and specified group.
func (f *RisingWaveObjectFactory) NewMetaAdvancedStatefulSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1beta1.StatefulSet {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Meta,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Meta.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Meta.RestartAt

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentMeta, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupMetaContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Set readiness gate for in place update strategy.
	podTemplate.Spec.ReadinessGates = append(podTemplate.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the CloneSet
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentMeta, group)
	metaSts := &kruiseappsv1beta1.StatefulSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentMeta, group, true),
		Spec: kruiseappsv1beta1.StatefulSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			ServiceName:    f.componentName(consts.ComponentMeta, ""),
			UpdateStrategy: buildUpgradeStrategyForAdvancedStatefulSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template:            podTemplate,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: &kruiseappsv1beta1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: kruiseappsv1beta1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  kruiseappsv1beta1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, metaSts, f.scheme)
}

func (f *RisingWaveObjectFactory) portsForFrontendContainer() []corev1.ContainerPort {
	frontendPorts := &f.risingwave.Spec.Components.Frontend.Ports

	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendPorts.ServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: frontendPorts.MetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupFrontendContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate) {
	basicSetupContainer(container, template, f.risingwave.Spec.Image)

	container.Name = "frontend"
	container.Args = []string{"frontend-node"}
	container.Env = append(container.Env, f.envsForFrontendArgs()...)
	container.Ports = f.portsForFrontendContainer()

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

// NewFrontendDeployment creates a new Deployment for the frontend component and specified group.
func (f *RisingWaveObjectFactory) NewFrontendDeployment(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
	// TODO setup the TLS configs

	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Frontend,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Frontend.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Frontend.RestartAt

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentFrontend, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupFrontendContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentFrontend, group)
	frontendDeployment := &appsv1.Deployment{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentFrontend, group, true),
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(componentGroup.Replicas),
			Strategy: buildUpgradeStrategyForDeployment(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, frontendDeployment, f.scheme)
}

// NewFrontendCloneSet creates a new CloneSet for the frontend component and specified group.
func (f *RisingWaveObjectFactory) NewFrontendCloneSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1alpha1.CloneSet {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Frontend,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Frontend.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Frontend.RestartAt

	podTemplate := f.buildPodTemplate(consts.ComponentFrontend, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	f.setupFrontendContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	keepPodSpecConsistent(&podTemplate.Spec)

	// Set readiness gate for in place update strategy.
	podTemplate.Spec.ReadinessGates = append(podTemplate.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentFrontend, group)
	frontendCloneSet := &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentFrontend, group, true),
		Spec: kruiseappsv1alpha1.CloneSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForCloneSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, frontendCloneSet, f.scheme)
}

func (f *RisingWaveObjectFactory) portsForCompactorContainer() []corev1.ContainerPort {
	compactorPorts := &f.risingwave.Spec.Components.Compactor.Ports

	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: compactorPorts.ServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: compactorPorts.MetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupCompactorContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate) {
	basicSetupContainer(container, template, f.risingwave.Spec.Image)

	container.Name = "compactor"
	container.Args = []string{"compactor-node"}
	container.Env = append(container.Env, f.envsForCompactorArgs()...)
	container.Ports = f.portsForCompactorContainer()

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

// NewCompactorDeployment creates a new Deployment for the compactor component and specified group.
func (f *RisingWaveObjectFactory) NewCompactorDeployment(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Compactor,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Compactor.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Compactor.RestartAt

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentCompactor, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupCompactorContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompactor, group)
	compactorDeployment := &appsv1.Deployment{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentCompactor, group, true),
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(componentGroup.Replicas),
			Strategy: buildUpgradeStrategyForDeployment(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, compactorDeployment, f.scheme)
}

// NewCompactorCloneSet creates a new CloneSet for the compactor component and specified group.
func (f *RisingWaveObjectFactory) NewCompactorCloneSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1alpha1.CloneSet {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Compactor,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Compactor.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Compactor.RestartAt

	podTemplate := f.buildPodTemplate(consts.ComponentCompactor, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	f.setupCompactorContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Set readiness gate for in place update strategy.
	podTemplate.Spec.ReadinessGates = append(podTemplate.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	keepPodSpecConsistent(&podTemplate.Spec)
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompactor, group)
	compactorCloneSet := &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentCompactor, group, true),
		Spec: kruiseappsv1alpha1.CloneSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForCloneSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, compactorCloneSet, f.scheme)

}

func (f *RisingWaveObjectFactory) portsForConnectorContainer() []corev1.ContainerPort {
	connectorPorts := f.getConnectorPorts()

	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: connectorPorts.ServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: connectorPorts.MetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupConnectorContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate) {
	basicSetupContainer(container, template, f.risingwave.Spec.Image)

	container.Name = "connector"
	container.Args = f.argsForConnector()
	container.Ports = f.portsForConnectorContainer()
	container.Command = []string{"/risingwave/bin/connector-node/start-service.sh"}
	memLimits := template.Resources.Limits.Memory().Value()
	if memLimits != 0 {
		container.Env = append(container.Env, corev1.EnvVar{
			Name:  envs.JavaOpts,
			Value: fmt.Sprintf("-Xmx%d", memLimits),
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

// NewConnectorDeployment creates a new Deployment for the connector component and specified group.
func (f *RisingWaveObjectFactory) NewConnectorDeployment(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Connector,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Connector.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Connector.RestartAt

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentConnector, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupConnectorContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentConnector, group)
	connectorDeployment := &appsv1.Deployment{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentConnector, group, true),
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(componentGroup.Replicas),
			Strategy: buildUpgradeStrategyForDeployment(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, connectorDeployment, f.scheme)
}

// NewConnectorCloneSet creates a new CloneSet for the connector component and specified group.
func (f *RisingWaveObjectFactory) NewConnectorCloneSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1alpha1.CloneSet {
	componentGroup := buildComponentGroup(
		f.risingwave.Spec.Global.Replicas.Connector,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Connector.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Connector.RestartAt

	podTemplate := f.buildPodTemplate(consts.ComponentConnector, group, podTemplates, componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	f.setupConnectorContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Set readiness gate for in place update strategy.
	podTemplate.Spec.ReadinessGates = append(podTemplate.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	keepPodSpecConsistent(&podTemplate.Spec)
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentConnector, group)
	connectorCloneSet := &kruiseappsv1alpha1.CloneSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentConnector, group, true),
		Spec: kruiseappsv1alpha1.CloneSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForCloneSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, connectorCloneSet, f.scheme)
}

func buildUpgradeStrategyForStatefulSet(strategy risingwavev1alpha1.RisingWaveUpgradeStrategy) appsv1.StatefulSetUpdateStrategy {
	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRollingUpdate:
		return appsv1.StatefulSetUpdateStrategy{
			Type: appsv1.RollingUpdateStatefulSetStrategyType,
			RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{
				MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
			},
		}
	default:
		return appsv1.StatefulSetUpdateStrategy{}
	}
}

func buildUpgradeStrategyForAdvancedStatefulSet(strategy risingwavev1alpha1.RisingWaveUpgradeStrategy) kruiseappsv1beta1.StatefulSetUpdateStrategy {
	advancedStatefulSetUpgradeStrategy := kruiseappsv1beta1.StatefulSetUpdateStrategy{}
	advancedStatefulSetUpgradeStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType
	advancedStatefulSetUpgradeStrategy.RollingUpdate = &kruiseappsv1beta1.RollingUpdateStatefulSetStrategy{
		MaxUnavailable: rollingUpdateOrDefault(strategy.RollingUpdate).MaxUnavailable,
	}
	if strategy.InPlaceUpdateStrategy != nil {
		advancedStatefulSetUpgradeStrategy.RollingUpdate.InPlaceUpdateStrategy = strategy.InPlaceUpdateStrategy.DeepCopy()
	}

	if rollingUpdateOrDefault(strategy.RollingUpdate).Partition != nil {
		// Change a percentage to an integer, partition only accepts int pointers for advanced stateful sets
		if rollingUpdateOrDefault(strategy.RollingUpdate).Partition.Type != intstr.Int {
			intValue, err := strconv.Atoi(strings.Replace((strategy.RollingUpdate).Partition.StrVal, "%", "", -1))
			if err != nil {
				panic(err)
			}
			advancedStatefulSetUpgradeStrategy.RollingUpdate.Partition = pointer.Int32(int32(intValue))
		} else {
			advancedStatefulSetUpgradeStrategy.RollingUpdate.Partition = pointer.Int32(rollingUpdateOrDefault(strategy.RollingUpdate).Partition.IntVal)
		}
	}

	switch strategy.Type {
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeRecreate:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.RecreatePodUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceIfPossible:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.InPlaceIfPossiblePodUpdateStrategyType
	case risingwavev1alpha1.RisingWaveUpgradeStrategyTypeInPlaceOnly:
		advancedStatefulSetUpgradeStrategy.RollingUpdate.PodUpdatePolicy = kruiseappsv1beta1.InPlaceOnlyPodUpdateStrategyType
	default:
		return kruiseappsv1beta1.StatefulSetUpdateStrategy{}
	}

	return advancedStatefulSetUpgradeStrategy
}

func (f *RisingWaveObjectFactory) portsForComputeContainer() []corev1.ContainerPort {
	computePorts := &f.risingwave.Spec.Components.Compute.Ports

	return []corev1.ContainerPort{
		{
			Name:          consts.PortService,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: computePorts.ServicePort,
		},
		{
			Name:          consts.PortMetrics,
			Protocol:      corev1.ProtocolTCP,
			ContainerPort: computePorts.MetricsPort,
		},
	}
}

func (f *RisingWaveObjectFactory) setupComputeContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComputeGroupTemplate) {
	basicSetupContainer(container, &template.RisingWaveComponentGroupTemplate, f.risingwave.Spec.Image)

	container.Name = "compute"
	container.Args = []string{"compute-node"}

	cpuLimit := int64(math.Ceil(container.Resources.Limits.Cpu().AsApproximateFloat64()))
	memLimit, _ := container.Resources.Limits.Memory().AsInt64()
	container.Env = append(container.Env, f.envsForComputeArgs(cpuLimit, memLimit)...)
	container.Ports = f.portsForComputeContainer()

	for _, env := range f.envsForStateStore() {
		container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
			return a.Name == b.Name
		})
	}

	for _, volumeMount := range template.VolumeMounts {
		container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, volumeMount, func(a, b *corev1.VolumeMount) bool {
			return a.MountPath == b.MountPath
		})
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

func buildPersistentVolumeClaims(claims []risingwavev1alpha1.PersistentVolumeClaim) []corev1.PersistentVolumeClaim {
	if claims == nil {
		return nil
	}

	result := make([]corev1.PersistentVolumeClaim, 0, len(claims))
	for _, claim := range claims {
		claim := *claim.DeepCopy()
		result = append(result, corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:        claim.Name,
				Labels:      claim.Labels,
				Annotations: claim.Annotations,
				Finalizers:  claim.Finalizers,
			},
			Spec: claim.Spec,
		})
	}

	return result
}

// NewComputeStatefulSet creates a new StatefulSet for the compute component and specified group.
func (f *RisingWaveObjectFactory) NewComputeStatefulSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.StatefulSet {
	componentGroup := buildComputeGroup(
		f.risingwave.Spec.Global.Replicas.Compute,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Compute.Groups,
	)
	if componentGroup == nil {
		return nil
	}

	restartAt := f.risingwave.Spec.Components.Compute.RestartAt
	pvcTemplates := f.risingwave.Spec.Storages.PVCTemplates

	// Build the pod template.
	podTemplate := f.buildPodTemplate(consts.ComponentCompute, group, podTemplates, &componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Set up the first container.
	f.setupComputeContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComputeGroupTemplate)

	// Make sure it's stable among builds.
	keepPodSpecConsistent(&podTemplate.Spec)

	// Build the statefulset.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompute, group)
	computeStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentCompute, group, true),
		Spec: appsv1.StatefulSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			ServiceName:    f.componentName(consts.ComponentCompute, ""),
			UpdateStrategy: buildUpgradeStrategyForStatefulSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template:             podTemplate,
			VolumeClaimTemplates: buildPersistentVolumeClaims(pvcTemplates),
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, computeStatefulSet, f.scheme)
}

// NewComputeAdvancedStatefulSet creates a new OpenKruise StatefulSet for the compute component and specified group.
func (f *RisingWaveObjectFactory) NewComputeAdvancedStatefulSet(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *kruiseappsv1beta1.StatefulSet {
	componentGroup := buildComputeGroup(
		f.risingwave.Spec.Global.Replicas.Compute,
		&f.risingwave.Spec.Global.RisingWaveComponentGroupTemplate,
		group,
		f.risingwave.Spec.Components.Compute.Groups,
	)
	if componentGroup == nil {
		return nil
	}
	restartAt := f.risingwave.Spec.Components.Compute.RestartAt
	pvcTemplates := f.risingwave.Spec.Storages.PVCTemplates

	podTemplate := f.buildPodTemplate(consts.ComponentCompute, group, podTemplates, &componentGroup.RisingWaveComponentGroupTemplate, restartAt)

	// Readiness gate InPlaceUpdateReady required for advanced statefulset
	podTemplate.Spec.ReadinessGates = append(podTemplate.Spec.ReadinessGates, corev1.PodReadinessGate{
		ConditionType: kruisepubs.InPlaceUpdateReady,
	})

	f.setupComputeContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComputeGroupTemplate)

	keepPodSpecConsistent(&podTemplate.Spec)

	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompute, group)

	computeAdvancedStatefulSet := &kruiseappsv1beta1.StatefulSet{
		ObjectMeta: f.componentGroupObjectMeta(consts.ComponentCompute, group, true),
		Spec: kruiseappsv1beta1.StatefulSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			ServiceName:    f.componentName(consts.ComponentCompute, ""),
			UpdateStrategy: buildUpgradeStrategyForAdvancedStatefulSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template:             podTemplate,
			VolumeClaimTemplates: buildPersistentVolumeClaims(pvcTemplates),
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: &kruiseappsv1beta1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: kruiseappsv1beta1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  kruiseappsv1beta1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}
	return mustSetControllerReference(f.risingwave, computeAdvancedStatefulSet, f.scheme)

}

// NewServiceMonitor creates a new ServiceMonitor.
func (f *RisingWaveObjectFactory) NewServiceMonitor() *prometheusv1.ServiceMonitor {
	const (
		interval      = 5 * time.Second
		scrapeTimeout = 5 * time.Second
	)

	serviceMonitor := &prometheusv1.ServiceMonitor{
		ObjectMeta: f.objectMeta("risingwave-"+f.risingwave.Name, true),
		Spec: prometheusv1.ServiceMonitorSpec{
			JobLabel: "risingwave/" + f.risingwave.Name,
			TargetLabels: []string{
				consts.LabelRisingWaveName,
				consts.LabelRisingWaveComponent,
				consts.LabelRisingWaveGroup,
			},
			Endpoints: []prometheusv1.Endpoint{
				{
					Port:          consts.PortMetrics,
					Interval:      prometheusv1.Duration(fmt.Sprintf("%.0fs", interval.Seconds())),
					ScrapeTimeout: prometheusv1.Duration(fmt.Sprintf("%.0fs", scrapeTimeout.Seconds())),
					// we need to drop some metrics which maybe will produce too many series.
					MetricRelabelConfigs: []*prometheusv1.RelabelConfig{
						{
							SourceLabels: []prometheusv1.LabelName{"__name__"},
							Action:       "drop",
							Regex:        "batch_.+",
						},
						{
							SourceLabels: []prometheusv1.LabelName{"__name__"},
							Action:       "drop",
							Regex:        "stream_exchange_.+",
						},
					},
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.LabelRisingWaveName: f.risingwave.Name,
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, serviceMonitor, f.scheme)
}

func (f *RisingWaveObjectFactory) getConnectorPorts() *risingwavev1alpha1.RisingWaveComponentCommonPorts {
	connectorPorts := f.risingwave.Spec.Components.Connector.Ports.DeepCopy()
	if connectorPorts.ServicePort == 0 {
		connectorPorts.ServicePort = consts.DefaultConnectorServicePort
	}
	if connectorPorts.MetricsPort == 0 {
		connectorPorts.MetricsPort = consts.DefaultConnectorMetricsPort
	}
	return connectorPorts
}

// NewRisingWaveObjectFactory creates a new RisingWaveObjectFactory.
func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme, operatorVersion string) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave:      risingwave,
		scheme:          scheme,
		inheritedLabels: captureInheritedLabels(risingwave),
		operatorVersion: operatorVersion,
	}
}
