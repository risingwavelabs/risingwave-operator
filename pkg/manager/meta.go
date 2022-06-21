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
	logger "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type MetaNodeManager struct {
}

func NewMetaMetaNodeManager() *MetaNodeManager {
	return &MetaNodeManager{}
}

func (m *MetaNodeManager) Name() string {
	return MetaNodeName
}

func (m *MetaNodeManager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	log := logger.FromContext(ctx)

	deployment := generateMetaDeployment(rw)

	log.Info("Render deploy succeed", "key", deployment.Namespace+"/"+deployment.Name)
	err := CreateIfNotFound(ctx, c, deployment)
	if err != nil && !errors.IsAlreadyExists(err) {
		log.Error(err, "create or update object failed", "name", deployment.Namespace+"/"+deployment.Name)
		return err
	}

	service := generateMetaService(rw)
	err = CreateIfNotFound(ctx, c, service)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	if ServiceMonitorFlagFromContext(ctx) {
		sm := GenerateServiceMonitor(MetaNodeComponentName(rw.Name), v1alpha1.MetaMetricsPortName, rw)
		err = CreateIfNotFound(ctx, c, sm)
		if err != nil && !errors.IsAlreadyExists(err) {
			return fmt.Errorf("create service monitor failed, %w", err)
		}
	}

	return nil
}

func (m *MetaNodeManager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	newDeploy := generateMetaDeployment(rw)
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

	// if deployment spec different. update it
	// TODO: add image change event for upgrading
	if newDeploy.Spec.Replicas != deploy.Spec.Replicas {
		log := logger.FromContext(ctx)
		log.Info("Need update deployment", "key", newDeploy.Namespace+"/"+newDeploy.Name)
		return true, CreateOrUpdateObject(ctx, c, newDeploy)
	}

	return false, nil
}

// DeleteService do deletion for meta node.
// TODO: add some func calls to ensure stop gracefully.
func (m *MetaNodeManager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	log := logger.FromContext(ctx)
	log.V(1).Info("Delete meta service")

	var namespacedName = types.NamespacedName{
		Namespace: rw.GetNamespace(),
		Name:      MetaNodeComponentName(rw.Name),
	}
	err := DeleteObjectByObjectKey(ctx, c, namespacedName, &corev1.Service{})
	if err != nil {
		return err
	}

	if ServiceMonitorFlagFromContext(ctx) {
		err := DeleteServiceMonitor(ctx, c, MetaNodeComponentName(rw.Name), rw)
		if err != nil {
			return err
		}
	}

	return DeleteObjectByObjectKey(ctx, c, namespacedName, &v1.Deployment{})
}

func (m *MetaNodeManager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var deployment v1.Deployment

	var namespacedName = types.NamespacedName{
		Namespace: rw.GetNamespace(),
		Name:      MetaNodeComponentName(rw.Name),
	}

	// if stats.Replicas == spec.Replicas, means ready
	// TODO: add health check
	err := wait.PollImmediate(RetryPeriod, RetryTimeout, func() (bool, error) {
		err := c.Get(ctx, namespacedName, &deployment)
		if err != nil {
			return false, fmt.Errorf("get deploy failed, %w", err)
		}

		if deployment.Status.AvailableReplicas == deployment.Status.Replicas &&
			deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("could not ensure service, %w", err)
	}
	return nil
}

func (m *MetaNodeManager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var deployment v1.Deployment

	var namespacedName = types.NamespacedName{
		Namespace: rw.GetNamespace(),
		Name:      MetaNodeComponentName(rw.Name),
	}

	err := c.Get(ctx, namespacedName, &deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get deploy failed, %w", err)
	}

	if deployment.Status.AvailableReplicas == deployment.Status.Replicas &&
		deployment.Status.AvailableReplicas == *deployment.Spec.Replicas {
		return true, nil
	}

	return false, nil
}

var _ ComponentManager = &MetaNodeManager{}

func generateMetaDeployment(rw *v1alpha1.RisingWave) *v1.Deployment {
	spec := rw.Spec.MetaNode

	var tag = "latest"
	if spec.Image.Tag != nil {
		tag = *spec.Image.Tag
	}

	container := corev1.Container{
		Name:            "meta-node",
		Resources:       *spec.Resources,
		Image:           fmt.Sprintf("%s:%s", *spec.Image.Repository, tag),
		ImagePullPolicy: *spec.Image.PullPolicy,
		Ports:           spec.Ports,
		Command: []string{
			"/risingwave/bin/risingwave",
		},
		Args: []string{
			"meta-node",
			"--listen-addr",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaServerPort),
			"--host",
			"$(POD_IP)",
			"--dashboard-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaDashboardPort),
			"--prometheus-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaMetricsPort),
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
					Port: intstr.FromInt(v1alpha1.MetaServerPort),
				},
			},
		},
	}

	var storage []string
	if spec.Storage.Type == v1alpha1.InMemory {
		storage = []string{"--backend", "mem"}
	}

	// TODO: maybe support other storage
	container.Args = append(container.Args, storage...)

	if len(spec.CMD) != 0 {
		container.Command = make([]string, len(spec.CMD))
		copy(container.Command, spec.CMD)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			container,
		},
	}

	if len(spec.NodeSelector) != 0 {
		podSpec.NodeSelector = spec.NodeSelector
	}

	if spec.Affinity != nil {
		podSpec.Affinity = spec.Affinity
	}

	var deploy = &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      MetaNodeComponentName(rw.Name),
		},
		Spec: v1.DeploymentSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					RisingWaveKey:  RisingWaveMetaValue,
					RisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RisingWaveKey:  RisingWaveMetaValue,
						RisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return deploy
}

func generateMetaService(rw *v1alpha1.RisingWave) *corev1.Service {
	spec := corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			RisingWaveKey:  RisingWaveMetaValue,
			RisingWaveName: rw.Name,
		},
	}

	var ports []corev1.ServicePort
	for _, p := range rw.Spec.MetaNode.Ports {
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
			Name:      MetaNodeComponentName(rw.Name),

			Labels: map[string]string{
				ServiceNameKey: MetaNodeComponentName(rw.Name),
				UIDKey:         string(rw.UID),
			},
		},
		Spec: spec,
	}
	return &s
}
