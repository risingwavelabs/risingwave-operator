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

type FrontendManager struct {
}

func NewFrontendManager() *FrontendManager {
	return &FrontendManager{}
}

func (m *FrontendManager) Name() string {
	return FrontendName
}

func (m *FrontendManager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	err := CreateIfNotFound(ctx, c, generateFrontendDeployment(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = CreateIfNotFound(ctx, c, generateFrontendService(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m *FrontendManager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	newDeploy := generateFrontendDeployment(rw)
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

	// if statefulSet spec different. update it
	// TODO: add image change event for upgrading
	if newDeploy.Spec.Replicas != deploy.Spec.Replicas {
		return true, CreateOrUpdateObject(ctx, c, newDeploy)
	}

	return false, nil
}

func (m *FrontendManager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      FrontendComponentName(rw.Name),
	}
	err := DeleteObjectByObjectKey(ctx, c, namespacedName, &corev1.Service{})
	if err != nil {
		return err
	}

	return DeleteObjectByObjectKey(ctx, c, namespacedName, &v1.Deployment{})
}

func (m *FrontendManager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var oldD v1.Deployment
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      FrontendComponentName(rw.Name),
	}

	// if stats.Replicas == spec.Replicas, means ready
	// TODO: add health check
	err := wait.PollImmediate(RetryPeriod, RetryTimeout, func() (bool, error) {
		err := c.Get(ctx, namespacedName, &oldD)
		if err != nil {
			return false, fmt.Errorf("get deploy failed, %w", err)
		}

		if oldD.Status.ReadyReplicas == oldD.Status.Replicas &&
			oldD.Status.ReadyReplicas == *oldD.Spec.Replicas {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		return fmt.Errorf("could not ensure frontend service, %w", err)
	}
	return nil
}

func (m *FrontendManager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var oldD v1.Deployment
	var namespacedName = types.NamespacedName{
		Namespace: rw.Namespace,
		Name:      FrontendComponentName(rw.Name),
	}
	err := c.Get(ctx, namespacedName, &oldD)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		return false, fmt.Errorf("get sts failed, %w", err)
	}

	if oldD.Status.ReadyReplicas == oldD.Status.Replicas &&
		oldD.Status.ReadyReplicas == *oldD.Spec.Replicas {
		return true, nil
	}

	return false, nil
}

var _ ComponentManager = &FrontendManager{}

func generateFrontendDeployment(rw *v1alpha1.RisingWave) *v1.Deployment {
	spec := rw.Spec.Frontend
	var tag = "latest"
	if spec.Image.Tag != nil {
		tag = *spec.Image.Tag
	}

	var c = corev1.Container{
		Name:      FrontendContainerName,
		Image:     fmt.Sprintf("%s:%s", *spec.Image.Repository, tag),
		Resources: *spec.Resources,
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
			fmt.Sprintf("http://%s:%d", MetaNodeComponentName(rw.Name), v1alpha1.MetaServerPort),
		},
		Ports: spec.Ports,
	}

	if len(spec.CMD) != 0 {
		c.Command = make([]string, len(spec.CMD))
		copy(c.Command, spec.CMD)
	}

	podSpec := corev1.PodSpec{
		Containers: []corev1.Container{
			c,
		},
	}

	if len(spec.NodeSelector) != 0 {
		podSpec.NodeSelector = spec.NodeSelector
	}

	if spec.Affinity != nil {
		podSpec.Affinity = spec.Affinity
	}

	d := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FrontendComponentName(rw.Name),
			Namespace: rw.Namespace,
		},
		Spec: v1.DeploymentSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					RisingWaveKey:  RisingWaveFrontendValue,
					RisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RisingWaveKey:  RisingWaveFrontendValue,
						RisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return d
}

func generateFrontendService(rw *v1alpha1.RisingWave) *corev1.Service {
	spec := corev1.ServiceSpec{
		Type: corev1.ServiceTypeNodePort,
		Selector: map[string]string{
			RisingWaveKey:  RisingWaveFrontendValue,
			RisingWaveName: rw.Name,
		},
	}

	var ports []corev1.ServicePort
	for _, p := range rw.Spec.Frontend.Ports {
		ports = append(ports, corev1.ServicePort{
			Protocol:   corev1.ProtocolTCP,
			Port:       p.ContainerPort,
			TargetPort: intstr.FromInt(int(p.ContainerPort)),
			Name:       p.Name,
		})
	}
	spec.Ports = ports

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      FrontendComponentName(rw.Name),
		},
		Spec: spec,
	}
	return s
}
