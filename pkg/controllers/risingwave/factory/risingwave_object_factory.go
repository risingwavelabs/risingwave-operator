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

	"github.com/samber/lo"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/consts"
)

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
	return f.risingwave.Spec.ObjectStorage.S3 != nil
}

func (f *RisingWaveObjectFactory) storeParam() string {
	storage := f.risingwave.Spec.ObjectStorage
	switch {
	case storage.S3 != nil:
		var bucket = *storage.S3.Bucket
		return fmt.Sprintf("hummock+s3://%s", bucket)
	case storage.Memory:
		return "hummock+memory"
	case storage.MinIO != nil:
		return fmt.Sprintf("hummock+minio://hummock:12345678@%s:%d/hummock001", f.risingwave.Name+"-minio", risingwavev1alpha1.MinIOServerPort)
	default:
		return "not-supported"
	}
}

func (f *RisingWaveObjectFactory) componentName(component string) string {
	switch component {
	case consts.ComponentMeta:
		return f.risingwave.Name + "-meta"
	case consts.ComponentCompute:
		return f.risingwave.Name + "-compute"
	case consts.ComponentFrontend:
		return f.risingwave.Name + "-frontend"
	case consts.ComponentCompactor:
		return f.risingwave.Name + "-compactor"
	case consts.ComponentConfig:
		return f.risingwave.Name + "-config"
	default:
		panic("never reach here")
	}
}

func (f *RisingWaveObjectFactory) objectMeta(component string, noSync bool) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      f.componentName(component),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:       f.risingwave.Name,
			consts.LabelRisingWaveComponent:  component,
			consts.LabelRisingWaveGeneration: lo.If(noSync, consts.NoSync).Else(strconv.FormatInt(f.risingwave.Generation, 10)),
		},
	}
}

func (f *RisingWaveObjectFactory) objectMetaSync(component string) metav1.ObjectMeta {
	return f.objectMeta(component, false)
}

func (f *RisingWaveObjectFactory) objectMetaNoSync(component string) metav1.ObjectMeta {
	return f.objectMeta(component, true)
}

func (f *RisingWaveObjectFactory) podLabelsOrSelectors(component string) map[string]string {
	return map[string]string{
		consts.LabelRisingWaveName:      f.risingwave.Name,
		consts.LabelRisingWaveComponent: component,
	}
}

func (f *RisingWaveObjectFactory) convertContainerPortsToServicePorts(containerPorts []corev1.ContainerPort) []corev1.ServicePort {
	return lo.Map(containerPorts, func(p corev1.ContainerPort, _ int) corev1.ServicePort {
		return corev1.ServicePort{
			Protocol:   corev1.ProtocolTCP,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromString(p.Name),
			Name:       p.Name,
		}
	})
}

func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	metaService := &corev1.Service{
		ObjectMeta: f.objectMetaSync(consts.ComponentMeta),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentMeta),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.MetaNode.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, metaService, f.scheme)
}

type containerPatch func(c *corev1.Container)

func (f *RisingWaveObjectFactory) newContainerFor(name string, descriptor *risingwavev1alpha1.DeployDescriptor, patches ...containerPatch) corev1.Container {
	image := fmt.Sprintf("%s:%s", *descriptor.Image.Repository, lo.If(descriptor.Image.Tag != nil, *descriptor.Image.Tag).Else("latest"))

	c := corev1.Container{
		Name:            name,
		Resources:       *descriptor.Resources.DeepCopy(),
		Image:           image,
		ImagePullPolicy: *descriptor.Image.PullPolicy,
		Ports:           descriptor.Ports,
		Command:         lo.If(len(descriptor.CMD) > 0, descriptor.DeepCopy().CMD).Else([]string{"/risingwave/bin/risingwave"}),
	}

	for _, patch := range patches {
		patch(&c)
	}

	return c
}

func (f *RisingWaveObjectFactory) patchArgsForMeta(c *corev1.Container) {
	metaNodeSpec := f.risingwave.Spec.MetaNode

	args := []string{
		"meta-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--listen-addr", fmt.Sprintf("0.0.0.0:%d", risingwavev1alpha1.MetaServerPort),
		"--host", "$(POD_IP)",
		"--dashboard-host", fmt.Sprintf("0.0.0.0:%d", risingwavev1alpha1.MetaDashboardPort),
		"--prometheus-host", fmt.Sprintf("0.0.0.0:%d", risingwavev1alpha1.MetaMetricsPort),
	}

	// TODO support other storages.
	if metaNodeSpec.Storage.Type == risingwavev1alpha1.InMemory {
		args = append(args, "--backend", "mem")
	} else if metaNodeSpec.Storage.Type == risingwavev1alpha1.ETCD {
		args = append(args, "--backend", "etcd", "--etcd-endpoints", metaNodeSpec.Storage.EtcdEndpoint)
	} else {
		panic("unknown storage type")
	}

	c.Args = args
}

func (f *RisingWaveObjectFactory) NewMetaDeployment() *appsv1.Deployment {
	metaNodeSpec := f.risingwave.Spec.MetaNode

	metaDeployment := &appsv1.Deployment{
		ObjectMeta: f.objectMetaSync(consts.ComponentMeta),
		Spec: appsv1.DeploymentSpec{
			Replicas: metaNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentMeta),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentMeta),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						f.risingWaveConfigVolume(),
					},
					NodeSelector: metaNodeSpec.NodeSelector,
					Affinity:     metaNodeSpec.Affinity,
					Containers: []corev1.Container{
						f.newContainerFor("meta-node", &metaNodeSpec.DeployDescriptor,
							f.patchArgsForMeta,
							f.patchRisingWaveConfigVolumeMount,
							f.patchReadinessProbe(&corev1.Probe{
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString(risingwavev1alpha1.MetaServerPortName),
									},
								},
							}),
						),
					},
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, metaDeployment, f.scheme)
}

func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	frontendService := &corev1.Service{
		ObjectMeta: f.objectMetaSync(consts.ComponentFrontend),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentFrontend),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.Frontend.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, frontendService, f.scheme)
}

func (f *RisingWaveObjectFactory) patchStartupProbe(probe *corev1.Probe) containerPatch {
	return func(c *corev1.Container) {
		c.StartupProbe = probe
	}
}

func (f *RisingWaveObjectFactory) patchLivenessProbe(probe *corev1.Probe) containerPatch {
	return func(c *corev1.Container) {
		c.LivenessProbe = probe
	}
}

func (f *RisingWaveObjectFactory) patchReadinessProbe(probe *corev1.Probe) containerPatch {
	return func(c *corev1.Container) {
		c.ReadinessProbe = probe
	}
}

func (f *RisingWaveObjectFactory) patchPodIPEnv(c *corev1.Container) {
	for _, envVar := range c.Env {
		if envVar.Name == "POD_IP" {
			return
		}
	}
	c.Env = append(c.Env, corev1.EnvVar{
		Name: "POD_IP",
		ValueFrom: &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "status.podIP",
			},
		},
	})
}

func (f *RisingWaveObjectFactory) patchArgsForFrontend(c *corev1.Container) {
	args := []string{
		"frontend-node",
		"--host",
		fmt.Sprintf("$(POD_IP):%d", risingwavev1alpha1.FrontendPort),
		"--meta-addr",
		fmt.Sprintf("http://%s:%d", f.componentName(consts.ComponentMeta), risingwavev1alpha1.MetaServerPort),
	}

	c.Args = args
}

func (f *RisingWaveObjectFactory) NewFrontendDeployment() *appsv1.Deployment {
	frontendSpec := f.risingwave.Spec.Frontend

	frontendDeployment := &appsv1.Deployment{
		ObjectMeta: f.objectMetaSync(consts.ComponentFrontend),
		Spec: appsv1.DeploymentSpec{
			Replicas: frontendSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentFrontend),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentFrontend),
				},
				Spec: corev1.PodSpec{
					NodeSelector: frontendSpec.NodeSelector,
					Affinity:     frontendSpec.Affinity,
					Containers: []corev1.Container{
						f.newContainerFor("frontend", &frontendSpec.DeployDescriptor,
							f.patchPodIPEnv,
							f.patchArgsForFrontend,
							f.patchReadinessProbe(&corev1.Probe{
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString(risingwavev1alpha1.FrontendPortName),
									},
								},
							}),
						),
					},
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, frontendDeployment, f.scheme)
}

func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	computeService := &corev1.Service{
		ObjectMeta: f.objectMetaSync(consts.ComponentCompute),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentCompute),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.ComputeNode.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, computeService, f.scheme)
}

func (f *RisingWaveObjectFactory) s3EnvVars() []corev1.EnvVar {
	objectStorage := f.risingwave.Spec.ObjectStorage
	secretRef := corev1.LocalObjectReference{
		Name: objectStorage.S3.SecretName,
	}
	return []corev1.EnvVar{
		{
			Name: "AWS_REGION",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.AWSS3Region,
				},
			},
		},
		{
			Name: "AWS_ACCESS_KEY_ID",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.AWSS3AccessKeyID,
				},
			},
		},
		{
			Name: "AWS_SECRET_ACCESS_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: secretRef,
					Key:                  consts.AWSS3SecretAccessKey,
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) patchStorageEnvs(c *corev1.Container) {
	if f.isObjectStorageS3() {
		c.Env = append(c.Env, f.s3EnvVars()...)
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
					Name: f.componentName(consts.ComponentConfig),
				},
			},
		},
	}
}

func (f *RisingWaveObjectFactory) patchRisingWaveConfigVolumeMount(c *corev1.Container) {
	c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
		Name:      risingWaveConfigVolume,
		MountPath: "/risingwave/config",
		ReadOnly:  true,
	})
}

func (f *RisingWaveObjectFactory) patchArgsForCompute(c *corev1.Container) {
	c.Args = []string{ // TODO: mv args -> configuration file
		"compute-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--host", fmt.Sprintf("$(POD_IP):%d", risingwavev1alpha1.ComputeNodePort),
		fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", risingwavev1alpha1.ComputeNodeMetricsPort),
		"--metrics-level=1",
		fmt.Sprintf("--state-store=%s", f.storeParam()),
		fmt.Sprintf("--meta-address=http://%s:%d", f.componentName(consts.ComponentMeta), risingwavev1alpha1.MetaServerPort),
	}
}

func (f *RisingWaveObjectFactory) NewComputeStatefulSet() *appsv1.StatefulSet {
	computeNodeSpec := f.risingwave.Spec.ComputeNode

	computeStatefulSet := &appsv1.StatefulSet{
		ObjectMeta: f.objectMetaSync(consts.ComponentCompute),
		Spec: appsv1.StatefulSetSpec{
			Replicas: computeNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentCompute),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentCompute),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						f.risingWaveConfigVolume(),
					},
					NodeSelector: computeNodeSpec.NodeSelector,
					Affinity:     computeNodeSpec.Affinity,
					Containers: []corev1.Container{
						f.newContainerFor("compute-node", &computeNodeSpec.DeployDescriptor,
							f.patchPodIPEnv,
							f.patchArgsForCompute,
							f.patchStorageEnvs,
							f.patchRisingWaveConfigVolumeMount,
							f.patchReadinessProbe(&corev1.Probe{
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString(risingwavev1alpha1.ComputeNodePortName),
									},
								},
							}),
						),
					},
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, computeStatefulSet, f.scheme)
}

func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	compactorService := &corev1.Service{
		ObjectMeta: f.objectMetaSync(consts.ComponentCompactor),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentCompactor),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.ComputeNode.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, compactorService, f.scheme)
}

func (f *RisingWaveObjectFactory) patchArgsForCompactor(c *corev1.Container) {
	c.Args = []string{
		"compactor-node",
		"--config-path", "/risingwave/config/risingwave.toml",
		"--host", fmt.Sprintf("$(POD_IP):%d", risingwavev1alpha1.CompactorNodePort),
		fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", risingwavev1alpha1.CompactorNodeMetricsPort),
		"--metrics-level=1",
		fmt.Sprintf("--state-store=%s", f.storeParam()),
		fmt.Sprintf("--meta-address=http://%s:%d", f.componentName(consts.ComponentMeta), risingwavev1alpha1.MetaServerPort),
	}
}

func (f *RisingWaveObjectFactory) NewCompactorDeployment() *appsv1.Deployment {
	compactorNodeSpec := f.risingwave.Spec.CompactorNode

	compactorDeployment := &appsv1.Deployment{
		ObjectMeta: f.objectMetaSync(consts.ComponentCompactor),

		Spec: appsv1.DeploymentSpec{
			Replicas: compactorNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentCompactor),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentCompactor),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						f.risingWaveConfigVolume(),
					},
					NodeSelector: compactorNodeSpec.NodeSelector,
					Affinity:     compactorNodeSpec.Affinity,
					Containers: []corev1.Container{
						f.newContainerFor("compactor-node", &compactorNodeSpec.DeployDescriptor,
							f.patchPodIPEnv,
							f.patchArgsForCompactor,
							f.patchStorageEnvs,
							f.patchRisingWaveConfigVolumeMount,
							f.patchReadinessProbe(&corev1.Probe{
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								ProbeHandler: corev1.ProbeHandler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.FromString(risingwavev1alpha1.CompactorNodePortName),
									},
								},
							}),
						),
					},
				},
			},
		},
	}

	return mustSetControllerReference(f.risingwave, compactorDeployment, f.scheme)
}

func (f *RisingWaveObjectFactory) NewConfigConfigMap() *corev1.ConfigMap {
	// Not going to sync.
	risingwaveConfigConfigMap := &corev1.ConfigMap{
		ObjectMeta: f.objectMetaNoSync(consts.ComponentConfig),
		Data: map[string]string{
			risingWaveConfigMapKey: risingWaveConfigTemplate,
		},
	}

	return mustSetControllerReference(f.risingwave, risingwaveConfigConfigMap, f.scheme)
}

func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave: risingwave,
		scheme:     scheme,
	}
}
