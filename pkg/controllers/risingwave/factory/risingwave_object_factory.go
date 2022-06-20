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
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	risingwavev1alpha1 "github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/controllers/risingwave/consts"
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
		return "in-memory"
	case storage.MinIO != nil:
		return fmt.Sprintf("hummock+minio://hummock:12345678@%s:%d/hummock001", f.risingwave.Name+"-minio", v1alpha1.MinIOServerPort)
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
	default:
		panic("never reach here")
	}
}

func (f *RisingWaveObjectFactory) objectMeta(component string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      f.componentName(component),
		Namespace: f.namespace(),
		Labels: map[string]string{
			consts.LabelRisingWaveName:       f.risingwave.Name,
			consts.LabelRisingWaveComponent:  component,
			consts.LabelRisingWaveGeneration: strconv.FormatInt(f.risingwave.Generation, 10),
		},
	}
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
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
			Name:       p.Name,
		}
	})
}

func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	metaService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentMeta),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: f.podLabelsOrSelectors(consts.ComponentMeta),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.MetaNode.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, metaService, f.scheme)
}

func (f *RisingWaveObjectFactory) NewMetaDeployment() *appsv1.Deployment {
	metaNodeSpec := f.risingwave.Spec.MetaNode

	container := corev1.Container{
		Name:            "meta-node",
		Resources:       *metaNodeSpec.Resources,
		Image:           fmt.Sprintf("%s:%s", *metaNodeSpec.Image.Repository, lo.If(metaNodeSpec.Image.Tag != nil, *metaNodeSpec.Image.Tag).Else("latest")),
		ImagePullPolicy: *metaNodeSpec.Image.PullPolicy,
		Ports:           metaNodeSpec.Ports,
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{
			"meta-node",
			"--host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaServerPort),
			"--dashboard-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaDashboardPort),
			"--prometheus-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaMetricsPort),
		},
	}

	var storage []string
	if metaNodeSpec.Storage.Type == v1alpha1.InMemory {
		storage = []string{"--backend", "mem"}
	}

	// TODO: maybe support other storage
	container.Args = append(container.Args, storage...)

	if len(metaNodeSpec.CMD) != 0 {
		container.Command = make([]string, len(metaNodeSpec.CMD))
		copy(container.Command, metaNodeSpec.CMD)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			container,
		},
	}

	if len(metaNodeSpec.NodeSelector) != 0 {
		podSpec.NodeSelector = metaNodeSpec.NodeSelector
	}

	if metaNodeSpec.Affinity != nil {
		podSpec.Affinity = metaNodeSpec.Affinity
	}

	metaDeployment := &v1.Deployment{
		ObjectMeta: f.objectMeta(consts.ComponentMeta),
		Spec: v1.DeploymentSpec{
			Replicas: metaNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentMeta),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentMeta),
				},
				Spec: podSpec,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, metaDeployment, f.scheme)
}

func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	frontendService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentFrontend),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: f.podLabelsOrSelectors(consts.ComponentFrontend),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.Frontend.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, frontendService, f.scheme)
}

func (f *RisingWaveObjectFactory) NewFrontendDeployment() *appsv1.Deployment {
	frontendSpec := f.risingwave.Spec.Frontend

	var c = corev1.Container{
		Name:      "frontend",
		Image:     fmt.Sprintf("%s:%s", *frontendSpec.Image.Repository, lo.If(frontendSpec.Image.Tag != nil, *frontendSpec.Image.Tag).Else("latest")),
		Resources: *frontendSpec.Resources,
		Env: []corev1.EnvVar{
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{
			"frontend-node",
			"--host",
			fmt.Sprintf("$(POD_IP):%d", v1alpha1.FrontendPort),
			"--meta-addr",
			fmt.Sprintf("http://%s:%d", f.componentName(consts.ComponentMeta), v1alpha1.MetaServerPort),
		},
		Ports: frontendSpec.Ports,
	}

	if len(frontendSpec.CMD) != 0 {
		c.Command = make([]string, len(frontendSpec.CMD))
		copy(c.Command, frontendSpec.CMD)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			c,
		},
	}

	if len(frontendSpec.NodeSelector) != 0 {
		podSpec.NodeSelector = frontendSpec.NodeSelector
	}

	if frontendSpec.Affinity != nil {
		podSpec.Affinity = frontendSpec.Affinity
	}

	frontendDeployment := &v1.Deployment{
		ObjectMeta: f.objectMeta(consts.ComponentFrontend),
		Spec: v1.DeploymentSpec{
			Replicas: frontendSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentFrontend),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentFrontend),
				},
				Spec: podSpec,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, frontendDeployment, f.scheme)
}

func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	computeService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentCompute),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
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

func (f *RisingWaveObjectFactory) NewComputeDeployment() *appsv1.StatefulSet {
	computeNodeSpec := f.risingwave.Spec.ComputeNode

	container := corev1.Container{
		Name:            "compute-node",
		Resources:       *computeNodeSpec.Resources,
		Image:           fmt.Sprintf("%s:%s", *computeNodeSpec.Image.Repository, lo.If(computeNodeSpec.Image.Tag != nil, *computeNodeSpec.Image.Tag).Else("latest")),
		ImagePullPolicy: *computeNodeSpec.Image.PullPolicy,
		Ports:           computeNodeSpec.Ports,
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{ // TODO: mv args -> configuration file
			"compute-node",
			"--config-path",
			"/risingwave/config/risingwave.toml",
			"--host",
			fmt.Sprintf("$(POD_IP):%d", v1alpha1.ComputeNodePort),
			fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", v1alpha1.ComputeNodeMetricsPort),
			"--metrics-level=1",
			fmt.Sprintf("--state-store=%s", f.storeParam()),
			fmt.Sprintf("--meta-address=http://%s:%d", f.componentName(consts.ComponentMeta), v1alpha1.MetaServerPort),
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "compute-config",
				MountPath: "/risingwave/config",
				ReadOnly:  true,
			},
		},
	}

	if len(computeNodeSpec.CMD) != 0 {
		container.Command = make([]string, len(computeNodeSpec.CMD))
		copy(container.Command, computeNodeSpec.CMD)
	}

	if f.isObjectStorageS3() {
		container.Env = append(container.Env, f.s3EnvVars()...)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			container,
		},
		Volumes: []corev1.Volume{
			{
				Name: "compute-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{
							{
								Key:  "risingwave.toml",
								Path: "risingwave.toml",
							},
						},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: f.risingwave.Name + "-compute-configmap",
						},
					},
				},
			},
		},
	}

	if len(computeNodeSpec.NodeSelector) != 0 {
		podSpec.NodeSelector = computeNodeSpec.NodeSelector
	}

	computeStatefulSet := &v1.StatefulSet{
		ObjectMeta: f.objectMeta(consts.ComponentCompute),
		Spec: v1.StatefulSetSpec{
			Replicas: computeNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentCompute),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentCompute),
				},
				Spec: podSpec,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, computeStatefulSet, f.scheme)
}

func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	compactorService := &corev1.Service{
		ObjectMeta: f.objectMeta(consts.ComponentCompactor),
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeNodePort,
			Selector: f.podLabelsOrSelectors(consts.ComponentCompactor),
			Ports:    f.convertContainerPortsToServicePorts(f.risingwave.Spec.ComputeNode.Ports),
		},
	}
	return mustSetControllerReference(f.risingwave, compactorService, f.scheme)
}

func (f *RisingWaveObjectFactory) NewCompactorDeployment() *appsv1.Deployment {
	compactorNodeSpec := f.risingwave.Spec.CompactorNode
	imageTag := lo.If(compactorNodeSpec.Image.Tag != nil, *compactorNodeSpec.Image.Tag).Else("latest")

	container := corev1.Container{
		Name:            "compactor-node",
		Resources:       *compactorNodeSpec.Resources,
		Image:           fmt.Sprintf("%s:%s", *compactorNodeSpec.Image.Repository, imageTag),
		ImagePullPolicy: *compactorNodeSpec.Image.PullPolicy,
		Ports:           compactorNodeSpec.Ports,
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{
			"compactor-node",
			"--host",
			fmt.Sprintf("$(POD_IP):%d", v1alpha1.CompactorNodePort),
			fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", v1alpha1.CompactorNodeMetricsPort),
			"--metrics-level=1",
			fmt.Sprintf("--state-store=%s", f.storeParam()),
			fmt.Sprintf("--meta-address=http://%s:%d", f.componentName(consts.ComponentMeta), v1alpha1.MetaServerPort),
		},
		Env: []corev1.EnvVar{
			{
				Name: "POD_IP",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						FieldPath: "status.podIP",
					},
				},
			},
		},
	}

	if len(compactorNodeSpec.CMD) != 0 {
		container.Command = make([]string, len(compactorNodeSpec.CMD))
		copy(container.Command, compactorNodeSpec.CMD)
	}

	if f.isObjectStorageS3() {
		container.Env = append(container.Env, f.s3EnvVars()...)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			container,
		},
	}

	if len(compactorNodeSpec.NodeSelector) != 0 {
		podSpec.NodeSelector = compactorNodeSpec.NodeSelector
	}

	compactorDeployment := &v1.Deployment{
		ObjectMeta: f.objectMeta(consts.ComponentCompactor),

		Spec: v1.DeploymentSpec{
			Replicas: compactorNodeSpec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: f.podLabelsOrSelectors(consts.ComponentCompactor),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: f.podLabelsOrSelectors(consts.ComponentCompactor),
				},
				Spec: podSpec,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, compactorDeployment, f.scheme)
}

func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave: risingwave,
		scheme:     scheme,
	}
}
