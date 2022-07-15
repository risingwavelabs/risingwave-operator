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

package factory

import (
	"fmt"
	"strconv"
	"time"

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/consts"
)

func nonZeroOrDefault[T comparable](v T, defaultVal T) T {
	var zero T
	if v == zero {
		return defaultVal
	}
	return v
}

const (
	risingWaveConfigVolume   = "risingwave-config"
	risingWaveConfigMapKey   = "risingwave.toml"
	risingWaveConfigTemplate = `[ server ]
heartbeat_interval = 1000

[ streaming ]
checkpoint_interval_ms = 100

[ storage ]
sstable_size_mb = 256
block_size_kb = 16
bloom_false_positive = 0.1
share_buffers_sync_parallelism = 2
shared_buffer_capacity_mb = 1024
data_directory = "hummock_001"
write_conflict_detection_enabled = true
block_cache_capacity_mb = 256
meta_cache_capacity_mb = 64
disable_remote_compactor = false
enable_local_spill = true
local_object_store = "tempdisk"`

	envMinIOUsername = "MINIO_USERNAME"
	envMinIOPassword = "MINIO_PASSWORD"
	envEtcdUsername  = "ETCD_USERNAME"
	envEtcdPassword  = "ETCD_PASSWORD"

	risingwaveExecutablePath = "/risingwave/bin/risingwave"
)

type RisingWaveObjectFactory struct {
	scheme     *runtime.Scheme
	risingwave *risingwavev1alpha1.RisingWave
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

func (f *RisingWaveObjectFactory) isObjectStorageS3() bool {
	return f.risingwave.Spec.Storages.Object.S3 != nil
}

func (f *RisingWaveObjectFactory) isObjectStorageMinIO() bool {
	return f.risingwave.Spec.Storages.Object.MinIO != nil
}

func (f *RisingWaveObjectFactory) isMetaStorageEtcd() bool {
	return f.risingwave.Spec.Storages.Meta.Etcd != nil
}

func (f *RisingWaveObjectFactory) hummockConnectionStr() string {
	objectStorage := f.risingwave.Spec.Storages.Object
	switch {
	case objectStorage.Memory != nil && *objectStorage.Memory:
		return "hummock+memory"
	case objectStorage.S3 != nil:
		bucket := objectStorage.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case objectStorage.MinIO != nil:
		minio := objectStorage.MinIO
		return fmt.Sprintf("hummock+minio://$(%s):$(%s)@%s/%s", envMinIOUsername, envMinIOPassword, minio.Endpoint, minio.Bucket)
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
	case consts.ComponentConfig:
		return f.risingwave.Name + "-config"
	default:
		panic("never reach here")
	}
}

func (f *RisingWaveObjectFactory) objectMeta(component string, sync bool) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      f.componentName(component, ""),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:       f.risingwave.Name,
			consts.LabelRisingWaveComponent:  component,
			consts.LabelRisingWaveGeneration: lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
		},
	}
}

func (f *RisingWaveObjectFactory) groupObjectMeta(component, group string, sync bool) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      f.componentName(component, group),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:       f.risingwave.Name,
			consts.LabelRisingWaveComponent:  component,
			consts.LabelRisingWaveGeneration: lo.If(!sync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
			consts.LabelRisingWaveGroup:      group,
		},
	}
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

func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports

	metaService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentMeta, true),
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

func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	frontendPorts := &f.risingwave.Spec.Components.Frontend.Ports

	frontendService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentFrontend, true),
		Spec: corev1.ServiceSpec{
			Type:     f.risingwave.Spec.Global.ServiceType,
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

func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	computePorts := &f.risingwave.Spec.Components.Compute.Ports

	computeService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentCompute, true),
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

func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	compactorPorts := &f.risingwave.Spec.Components.Compactor.Ports

	compactorService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentCompactor, true),
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

func (f *RisingWaveObjectFactory) envsForEtcd() []corev1.EnvVar {
	etcd := f.risingwave.Spec.Storages.Meta.Etcd

	// Empty secret indicates no authentication.
	if etcd.Secret == "" {
		return []corev1.EnvVar{}
	}

	secretRef := corev1.LocalObjectReference{
		Name: etcd.Secret,
	}

	return []corev1.EnvVar{
		{
			Name: envEtcdUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyEtcdUsername,
				},
			},
		},
		{
			Name: envEtcdPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyEtcdPassword,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) argsForMeta() []string {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	metaStorage := &f.risingwave.Spec.Storages.Meta

	args := []string{
		"meta-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--listen-addr", fmt.Sprintf("0.0.0.0:%d", metaPorts.ServicePort),
		"--host", "$(POD_IP)",
		"--dashboard-host", fmt.Sprintf("0.0.0.0:%d", metaPorts.DashboardPort),
		"--prometheus-host", fmt.Sprintf("0.0.0.0:%d", metaPorts.MetricsPort),
	}

	switch {
	case metaStorage.Memory != nil && *metaStorage.Memory:
		args = append(args, "--backend", "mem")
	case metaStorage.Etcd != nil:
		args = append(args, "--backend", "etcd", "--etcd-endpoints", metaStorage.Etcd.Endpoint)
	default:
		panic("unsupported meta storage type")
	}

	return args
}

func (f *RisingWaveObjectFactory) argsForFrontend() []string {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	frontendPorts := &f.risingwave.Spec.Components.Frontend.Ports

	return []string{
		"frontend-node",
		"--host", fmt.Sprintf("$(POD_IP):%d", frontendPorts.ServicePort),
		"--meta-addr", fmt.Sprintf("http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
	}
}

func (f *RisingWaveObjectFactory) argsForCompute() []string {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	computePorts := &f.risingwave.Spec.Components.Compute.Ports

	return []string{ // TODO: mv args -> configuration file
		"compute-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--host", fmt.Sprintf("$(POD_IP):%d", computePorts.ServicePort),
		fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", computePorts.MetricsPort),
		"--metrics-level=1",
		"--state-store",
		f.hummockConnectionStr(),
		"--meta-address",
		fmt.Sprintf("http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
	}
}

func (f *RisingWaveObjectFactory) argsForCompactor() []string {
	metaPorts := &f.risingwave.Spec.Components.Meta.Ports
	compactorPorts := &f.risingwave.Spec.Components.Compactor.Ports

	return []string{
		"compactor-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--host", fmt.Sprintf("$(POD_IP):%d", compactorPorts.ServicePort),
		"--prometheus-listener-addr", fmt.Sprintf("0.0.0.0:%d", compactorPorts.MetricsPort),
		"--metrics-level=1",
		"--state-store", f.hummockConnectionStr(),
		"--meta-address", fmt.Sprintf("http://%s:%d", f.componentName(consts.ComponentMeta, ""), metaPorts.ServicePort),
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
	objectStorage := &f.risingwave.Spec.Storages.Object

	secretRef := corev1.LocalObjectReference{
		Name: objectStorage.MinIO.Secret,
	}

	return []corev1.EnvVar{
		{
			Name: envMinIOUsername,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyMinIOUsername,
				},
			},
		},
		{
			Name: envMinIOPassword,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyMinIOPassword,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) envsForS3() []corev1.EnvVar {
	objectStorage := &f.risingwave.Spec.Storages.Object

	secretRef := corev1.LocalObjectReference{
		Name: objectStorage.S3.Secret,
	}

	return []corev1.EnvVar{
		{
			Name: "AWS_REGION",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3Region,
				},
			},
		},
		{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3AccessKeyID,
				},
			},
		},
		{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.SecretKeyAWSS3SecretAccessKey,
				},
			},
		},
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
						Path: "risingwave.toml",
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
		MountPath: "/risingwave/config",
		ReadOnly:  true,
	}
}

func (f *RisingWaveObjectFactory) NewConfigConfigMap(val string) *corev1.ConfigMap {
	risingwaveConfigConfigMap := &corev1.ConfigMap{
		ObjectMeta: f.objectMeta(consts.ComponentConfig, false), // not synced
		Data: map[string]string{
			risingWaveConfigMapKey: nonZeroOrDefault(val, risingWaveConfigTemplate),
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

func mergeValue[T comparable](a, b T) T {
	var zero T
	if b == zero {
		return a
	}
	return b
}

func mergeValueList[T any](a, b []T) []T {
	return append(a, b...)
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

func isResourcesEmpty(resources corev1.ResourceRequirements) bool {
	return len(resources.Limits) == 0 && len(resources.Requests) == 0
}

func mergeComponentGroupTemplates(base, overlay *risingwavev1alpha1.RisingWaveComponentGroupTemplate) *risingwavev1alpha1.RisingWaveComponentGroupTemplate {
	if overlay == nil {
		return base.DeepCopy()
	}
	if base == nil {
		return overlay.DeepCopy()
	}

	r := base.DeepCopy()
	r.Image = mergeValue(r.Image, overlay.Image)
	r.ImagePullPolicy = mergeValue(r.ImagePullPolicy, overlay.ImagePullPolicy)
	r.ImagePullSecrets = mergeValueList(r.ImagePullSecrets, overlay.ImagePullSecrets)
	r.UpgradeStrategy = mergeValue(r.UpgradeStrategy, overlay.UpgradeStrategy)
	if !isResourcesEmpty(overlay.Resources) {
		r.Resources = overlay.Resources
	}
	r.NodeSelector = mergeMap(r.NodeSelector, overlay.NodeSelector)
	r.PodTemplate = mergeValue(r.PodTemplate, overlay.PodTemplate)

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

func (f *RisingWaveObjectFactory) buildPodTemplate(component, group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate,
	groupTemplate *risingwavev1alpha1.RisingWaveComponentGroupTemplate, restartAt *metav1.Time) corev1.PodTemplateSpec {
	var podTemplate corev1.PodTemplateSpec

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

	// Set config volume.
	podTemplate.Spec.Volumes = mergeListWhenKeyEquals(podTemplate.Spec.Volumes, f.risingWaveConfigVolume(), func(a, b *corev1.Volume) bool {
		return a.Name == b.Name
	})

	// Set labels and annotations.
	podTemplate.Labels = mergeMap(podTemplate.Labels, f.podLabelsOrSelectorsForGroup(component, group))
	if restartAt != nil {
		if podTemplate.Annotations == nil {
			podTemplate.Annotations = make(map[string]string)
		}
		podTemplate.Annotations[consts.AnnotationRestartAt] = restartAt.In(time.UTC).Format("2006-01-02T15:04:05Z")
	}

	// Setup the first container.
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
		})
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
		})
		if componentGroup == nil {
			return nil
		}
		componentGroup.RisingWaveComputeGroupTemplate.RisingWaveComponentGroupTemplate = *mergeComponentGroupTemplates(globalTemplate,
			&componentGroup.RisingWaveComputeGroupTemplate.RisingWaveComponentGroupTemplate)
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

func basicSetupContainer(container *corev1.Container, template *risingwavev1alpha1.RisingWaveComponentGroupTemplate) {
	container.Image = template.Image
	container.ImagePullPolicy = template.ImagePullPolicy
	container.Command = []string{risingwaveExecutablePath}
	container.Env = mergeListByKey(container.Env, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	}, func(env *corev1.EnvVar) bool {
		return env.Name == "POD_IP"
	})
	container.Resources = template.Resources
	container.StartupProbe = nil
	container.LivenessProbe = nil
	container.ReadinessProbe = &corev1.Probe{
		InitialDelaySeconds: 10,
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
	basicSetupContainer(container, template)

	container.Name = "meta"
	container.Args = f.argsForMeta()
	container.Ports = f.portsForMetaContainer()

	if f.isMetaStorageEtcd() {
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

func (f *RisingWaveObjectFactory) NewMetaDeployment(group string, podTemplates map[string]risingwavev1alpha1.RisingWavePodTemplate) *appsv1.Deployment {
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

	// Setup the first container.
	f.setupMetaContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentMeta, group)
	metaDeployment := &appsv1.Deployment{
		ObjectMeta: f.groupObjectMeta(consts.ComponentMeta, group, true),
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(componentGroup.Replicas),
			Strategy: buildUpgradeStrategyForDeployment(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template: podTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, metaDeployment, f.scheme)
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
	basicSetupContainer(container, template)

	container.Name = "frontend"
	container.Args = f.argsForFrontend()
	container.Ports = f.portsForFrontendContainer()

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

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

	// Setup the first container.
	f.setupFrontendContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentFrontend, group)
	frontendDeployment := &appsv1.Deployment{
		ObjectMeta: f.groupObjectMeta(consts.ComponentFrontend, group, true),
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
	basicSetupContainer(container, template)

	container.Name = "compactor"
	container.Args = f.argsForCompactor()
	container.Ports = f.portsForCompactorContainer()

	if f.isObjectStorageS3() {
		for _, env := range f.envsForS3() {
			container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
				return a.Name == b.Name
			})
		}
	} else if f.isObjectStorageMinIO() {
		for _, env := range f.envsForMinIO() {
			container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
				return a.Name == b.Name
			})
		}
	}

	container.VolumeMounts = mergeListWhenKeyEquals(container.VolumeMounts, f.volumeMountForConfig(), func(a, b *corev1.VolumeMount) bool {
		return a.MountPath == b.MountPath
	})
}

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

	// Setup the first container.
	f.setupCompactorContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComponentGroupTemplate)

	// Build the deployment.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompactor, group)
	compactorDeployment := &appsv1.Deployment{
		ObjectMeta: f.groupObjectMeta(consts.ComponentCompactor, group, true),
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
	basicSetupContainer(container, &template.RisingWaveComponentGroupTemplate)

	container.Name = "compute"
	container.Args = f.argsForCompute()
	container.Ports = f.portsForComputeContainer()

	if f.isObjectStorageS3() {
		for _, env := range f.envsForS3() {
			container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
				return a.Name == b.Name
			})
		}
	} else if f.isObjectStorageMinIO() {
		for _, env := range f.envsForMinIO() {
			container.Env = mergeListWhenKeyEquals(container.Env, env, func(a, b *corev1.EnvVar) bool {
				return a.Name == b.Name
			})
		}
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

	// Setup the first container.
	f.setupComputeContainer(&podTemplate.Spec.Containers[0], componentGroup.RisingWaveComputeGroupTemplate)

	// Build the statefulset.
	labelsOrSelectors := f.podLabelsOrSelectorsForGroup(consts.ComponentCompute, group)
	computeStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: f.groupObjectMeta(consts.ComponentCompute, group, true),
		Spec: appsv1.StatefulSetSpec{
			Replicas:       pointer.Int32(componentGroup.Replicas),
			UpdateStrategy: buildUpgradeStrategyForStatefulSet(componentGroup.UpgradeStrategy),
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsOrSelectors,
			},
			Template:             podTemplate,
			VolumeClaimTemplates: pvcTemplates,
			PodManagementPolicy:  appsv1.ParallelPodManagement,
			PersistentVolumeClaimRetentionPolicy: &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{
				WhenDeleted: appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
				WhenScaled:  appsv1.DeletePersistentVolumeClaimRetentionPolicyType,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, computeStatefulSet, f.scheme)
}

func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave: risingwave,
		scheme:     scheme,
	}
}
