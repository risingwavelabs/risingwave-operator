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

type CompactorNodeManager struct {
}

func NewCompactorNodeManager() *CompactorNodeManager {
	return &CompactorNodeManager{}
}

func (m *CompactorNodeManager) Name() string {
	return CompactorNodeName
}

func (m *CompactorNodeManager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	err := CreateIfNotFound(ctx, c, generateCompactorDeploy(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = CreateIfNotFound(ctx, c, generateCompactorService(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if ServiceMonitorFlagFromContext(ctx) {
		sm := GenerateServiceMonitor(CompactorNodeComponentName(rw.Name), v1alpha1.CompactorNodeMetricsPortName, rw)
		err = CreateIfNotFound(ctx, c, sm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create service monitor failed, %w", err)
		}
	}
	return nil
}

func (m *CompactorNodeManager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	newDeploy := generateCompactorDeploy(rw)
	var namespacedName = types.NamespacedName{
		Namespace: newDeploy.Namespace,
		Name:      newDeploy.Name,
	}
	var deploy v1.Deployment
	err := c.Get(ctx, namespacedName, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
	}

	if deploy.Spec.Replicas != newDeploy.Spec.Replicas {
		return true, CreateOrUpdateObject(ctx, c, newDeploy)
	}

	return false, nil
}

func (m *CompactorNodeManager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      CompactorNodeComponentName(rw.Name),
	}

	if ServiceMonitorFlagFromContext(ctx) {
		err := DeleteServiceMonitor(ctx, c, CompactorNodeComponentName(rw.Name), rw)
		if err != nil {
			return err
		}
	}

	err := DeleteObjectByObjectKey(ctx, c, namespacedName, &corev1.Service{})
	if err != nil {
		return err
	}

	return DeleteObjectByObjectKey(ctx, c, namespacedName, &v1.Deployment{})
}

func (m *CompactorNodeManager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var old v1.Deployment
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      CompactorNodeComponentName(rw.Name),
	}

	err := wait.PollImmediate(RetryPeriod, RetryTimeout, func() (bool, error) {
		err := c.Get(ctx, namespacedName, &old)
		if err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("get statefulset failed, %w", err)
		}

		if old.Status.ReadyReplicas == old.Status.Replicas &&
			old.Status.ReadyReplicas == *old.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("could not ensure compactor service, %w", err)
	}
	return nil
}

func (m *CompactorNodeManager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var deploy v1.Deployment
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      CompactorNodeComponentName(rw.Name),
	}
	err := c.Get(ctx, namespacedName, &deploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get sts failed, %w", err)
	}

	if deploy.Status.ReadyReplicas == deploy.Status.Replicas &&
		deploy.Status.ReadyReplicas == *deploy.Spec.Replicas {
		return true, nil
	}

	return false, nil
}

var _ ComponentManager = &CompactorNodeManager{}

func generateCompactorDeploy(rw *v1alpha1.RisingWave) *v1.Deployment {
	spec := rw.Spec.CompactorNode

	var tag = "latest"
	if spec.Image.Tag != nil {
		tag = *spec.Image.Tag
	}

	container := corev1.Container{
		Name:            "compactor-node",
		Resources:       *spec.Resources,
		Image:           fmt.Sprintf("%s:%s", *spec.Image.Repository, tag),
		ImagePullPolicy: *spec.Image.PullPolicy,
		Ports:           spec.Ports,
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{
			"compactor-node",
			"--host",
			fmt.Sprintf("$(POD_IP):%d", v1alpha1.CompactorNodePort),
			fmt.Sprintf("--prometheus-listener-addr=0.0.0.0:%d", v1alpha1.CompactorNodeMetricsPort),
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
		// tcp readiness probe
		ReadinessProbe: &corev1.Probe{
			InitialDelaySeconds: 10,
			PeriodSeconds:       10,
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt(v1alpha1.CompactorNodePort),
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
	}

	if len(spec.NodeSelector) != 0 {
		podSpec.NodeSelector = spec.NodeSelector
	}

	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      CompactorNodeComponentName(rw.Name),
		},

		Spec: v1.DeploymentSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					RisingWaveKey:  RisingWaveCompactorValue,
					RisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RisingWaveKey:  RisingWaveCompactorValue,
						RisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return deploy
}

func generateCompactorService(rw *v1alpha1.RisingWave) *corev1.Service {
	spec := corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			RisingWaveKey:  RisingWaveCompactorValue,
			RisingWaveName: rw.Name,
		},
		ClusterIP: "None",
	}

	var ports []corev1.ServicePort
	for _, p := range rw.Spec.CompactorNode.Ports {
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
			Name:      CompactorNodeComponentName(rw.Name),

			Labels: map[string]string{
				ServiceNameKey: CompactorNodeComponentName(rw.Name),
				UIDKey:         string(rw.UID),
			},
		},
		Spec: spec,
	}
	return &s
}
