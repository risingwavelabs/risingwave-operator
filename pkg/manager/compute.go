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

package manager

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/singularity-data/risingwave-operator/pkg/rendor"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type ComputeNodeManager struct {
}

func NewComputeNodeManager() *ComputeNodeManager {
	return &ComputeNodeManager{}
}

func (m *ComputeNodeManager) Name() string {
	return ComputeNodeName
}

func (m *ComputeNodeManager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	cwd, _ := os.Getwd()
	var p = path.Join(cwd, TemplateFileDir, ComputeNodeConfigTemplate)

	var opt = map[string]interface{}{
		"NAME":       computeNodeConfigmapName(rw.Name),
		"NAME_SPACE": rw.Namespace,
	}
	err := rendor.CreateObjectByTem(p, opt)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("rendor parse file failed, %w", err)
	}

	err = CreateIfNotFound(ctx, c, generateComputeStatefulSet(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = CreateIfNotFound(ctx, c, generateComputeService(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if ServiceMonitorFlagFromContext(ctx) {
		sm := GenerateServiceMonitor(ComputeNodeComponentName(rw.Name), v1alpha1.ComputeNodeMetricsPortName, rw)
		err = CreateIfNotFound(ctx, c, sm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create service monitor failed, %w", err)
		}
	}
	return nil
}

func (m *ComputeNodeManager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	newSts := generateComputeStatefulSet(rw)
	var namespacedName = types.NamespacedName{
		Namespace: newSts.Namespace,
		Name:      newSts.Name,
	}
	var sts v1.StatefulSet
	err := c.Get(ctx, namespacedName, &sts)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
	}

	// if statefulSet rs different. update it
	// TODO: add image change event for upgrading
	if sts.Spec.Replicas != newSts.Spec.Replicas {
		return true, CreateOrUpdateObject(ctx, c, newSts)
	}

	return false, nil
}

func (m *ComputeNodeManager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      ComputeNodeComponentName(rw.Name),
	}
	err := DeleteObjectByObjectKey(ctx, c, namespacedName, &corev1.Service{})
	if err != nil {
		return err
	}

	err = DeleteObjectByObjectKey(ctx, c, types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      computeNodeConfigmapName(rw.Name),
	}, &corev1.ConfigMap{})
	if err != nil {
		return err
	}

	if ServiceMonitorFlagFromContext(ctx) {
		err := DeleteServiceMonitor(ctx, c, computeNodeConfigmapName(rw.Name), rw)
		if err != nil {
			return err
		}
	}

	return DeleteObjectByObjectKey(ctx, c, namespacedName, &v1.StatefulSet{})
}

func (m *ComputeNodeManager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var oldSts v1.StatefulSet
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      ComputeNodeComponentName(rw.Name),
	}

	// if stats.Replicas == spec.Replicas, means ready
	// TODO: add health check
	err := wait.PollImmediate(RetryPeriod, RetryTimeout, func() (bool, error) {
		err := c.Get(ctx, namespacedName, &oldSts)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("get statefulset failed, %w", err)
		}

		if oldSts.Status.ReadyReplicas == oldSts.Status.Replicas &&
			oldSts.Status.ReadyReplicas == *oldSts.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("could not ensure compute service, %w", err)
	}
	return nil
}

func (m *ComputeNodeManager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var sts v1.StatefulSet
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      ComputeNodeComponentName(rw.Name),
	}
	err := c.Get(ctx, namespacedName, &sts)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get sts failed, %w", err)
	}

	if sts.Status.ReadyReplicas == sts.Status.Replicas &&
		sts.Status.ReadyReplicas == *sts.Spec.Replicas {
		return true, nil
	}

	return false, nil
}

var _ ComponentManager = &ComputeNodeManager{}

func computeNodeConfigmapName(name string) string {
	return fmt.Sprintf("%s-compute-configmap", name)
}

func generateComputeStatefulSet(rw *v1alpha1.RisingWave) *v1.StatefulSet {
	spec := rw.Spec.ComputeNode

	var tag = "latest"
	if spec.Image.Tag != nil {
		tag = *spec.Image.Tag
	}

	container := corev1.Container{
		Name:            "compute-node",
		Resources:       *spec.Resources,
		Image:           fmt.Sprintf("%s:%s", *spec.Image.Repository, tag),
		ImagePullPolicy: *spec.Image.PullPolicy,
		Ports:           spec.Ports,
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
			fmt.Sprintf("--state-store=%s", StoreParam(rw)),
			fmt.Sprintf("--meta-address=http://%s:%d", MetaNodeComponentName(rw.Name), v1alpha1.MetaServerPort),
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
				Name:      ComputeNodeTomlName,
				MountPath: "/risingwave/config",
				ReadOnly:  true,
			},
		},
		// tcp readiness probe
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(v1alpha1.ComputeNodePort),
				},
			},
		},
	}

	if len(spec.CMD) != 0 {
		container.Command = make([]string, len(spec.CMD))
		copy(container.Command, spec.CMD)
	}

	if rw.Spec.ObjectStorage.S3 != nil {
		var env = []corev1.EnvVar{
			{
				Name: "AWS_REGION",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: rw.Spec.ObjectStorage.S3.SecretName,
						},
						Key: Region,
					},
				},
			},
			{
				Name: "AWS_ACCESS_KEY_ID",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: rw.Spec.ObjectStorage.S3.SecretName,
						},
						Key: AccessKeyID,
					},
				},
			},
			{
				Name: "AWS_SECRET_ACCESS_KEY",
				ValueFrom: &corev1.EnvVarSource{
					SecretKeyRef: &corev1.SecretKeySelector{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: rw.Spec.ObjectStorage.S3.SecretName,
						},
						Key: SecretAccessKey,
					},
				},
			},
		}
		container.Env = append(container.Env, env...)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			container,
		},
		Volumes: []corev1.Volume{
			{
				Name: ComputeNodeTomlName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						Items: []corev1.KeyToPath{
							{
								Key:  ComputeNodeTomlKey,
								Path: ComputeNodeTomlPath,
							},
						},
						LocalObjectReference: corev1.LocalObjectReference{
							Name: computeNodeConfigmapName(rw.Name),
						},
					},
				},
			},
		},
	}

	if len(spec.NodeSelector) != 0 {
		podSpec.NodeSelector = spec.NodeSelector
	}

	sts := &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      ComputeNodeComponentName(rw.Name),
		},

		Spec: v1.StatefulSetSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					RisingWaveKey:  RisingWaveComputeValue,
					RisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RisingWaveKey:  RisingWaveComputeValue,
						RisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return sts
}

func generateComputeService(rw *v1alpha1.RisingWave) *corev1.Service {
	spec := corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			RisingWaveKey:  RisingWaveComputeValue,
			RisingWaveName: rw.Name,
		},
		ClusterIP: "None",
	}

	var ports []corev1.ServicePort
	for _, p := range rw.Spec.ComputeNode.Ports {
		ports = append(ports, corev1.ServicePort{
			Protocol:   corev1.ProtocolTCP,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
			Name:       p.Name,
		})
	}
	spec.Ports = ports

	s := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      ComputeNodeComponentName(rw.Name),

			Labels: map[string]string{
				ServiceNameKey: ComputeNodeComponentName(rw.Name),
				UIDKey:         string(rw.UID),
			},
		},
		Spec: spec,
	}
	return &s
}
