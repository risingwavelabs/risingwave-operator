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
	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logger "sigs.k8s.io/controller-runtime/pkg/log"
)

type MinIOManager struct {
}

func (m MinIOManager) Name() string {
	return MinIOName
}

func (m MinIOManager) CreateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	log := logger.FromContext(ctx)
	log.V(1).Info("Begin to create minIO")

	err := CreateIfNotFound(ctx, c, generateMinIODeployment(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	err = CreateIfNotFound(ctx, c, generateMinIOService(rw))
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (m MinIOManager) UpdateService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var deploymentNamespacedName = minIONamespacedName(rw)
	var oldDeploy v1.Deployment
	err := c.Get(ctx, deploymentNamespacedName, &oldDeploy)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
	}

	deployment := generateMinIODeployment(rw)
	// if deployment spec different. update it
	// TODO: add image change event for upgrading
	if deployment.Spec.Replicas != oldDeploy.Spec.Replicas {
		return true, CreateOrUpdateObject(ctx, c, deployment)
	}

	return false, nil
}

func (m MinIOManager) DeleteService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	deploymentNamespacedName := minIONamespacedName(rw)

	err := DeleteObjectByObjectKey(ctx, c, deploymentNamespacedName, &corev1.Service{})
	if err != nil {
		return err
	}

	return DeleteObjectByObjectKey(ctx, c, deploymentNamespacedName, &v1.Deployment{})
}

func (m MinIOManager) CheckService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) (bool, error) {
	var deployment v1.Deployment
	var deploymentNamespacedName = minIONamespacedName(rw)
	err := c.Get(ctx, deploymentNamespacedName, &deployment)
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

func (m MinIOManager) EnsureService(ctx context.Context, c client.Client, rw *v1alpha1.RisingWave) error {
	var deployment v1.Deployment
	var deploymentNamespacedName = minIONamespacedName(rw)

	//var servicePort int32 = 9301
	//if rw.Spec.ObjectStorage.MinIO.ServicePort != nil {
	//	servicePort = *rw.Spec.ObjectStorage.MinIO.ServicePort
	//}

	err := wait.PollImmediate(RetryPeriod, RetryTimeout, func() (bool, error) {
		err := c.Get(ctx, deploymentNamespacedName, &deployment)
		if err != nil {
			return false, fmt.Errorf("get deploy failed, %w", err)
		}

		if deployment.Status.AvailableReplicas != deployment.Status.Replicas || deployment.Status.AvailableReplicas != *deployment.Spec.Replicas {
			return false, nil
		}

		// TODO: maybe lead to minio cannot be ready
		//var healthCheckUrl = fmt.Sprintf("http://%s:%d/minio/health/live", MinIOComponentName(rw.Name), servicePort)
		//resp, err := http.Head(healthCheckUrl)
		//if err != nil {
		//	return false, fmt.Errorf("fail to ensure service with url, %s, %w", healthCheckUrl, err)
		//}
		//if resp.Status != "200 OK" {
		//	return false, nil
		//}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("could not ensure service, %w", err)
	}
	return nil
}

func NewMinIOManager() *MinIOManager {
	return &MinIOManager{}
}

func minIONamespacedName(rw *v1alpha1.RisingWave) types.NamespacedName {
	return types.NamespacedName{
		Namespace: rw.GetNamespace(),
		Name:      MinIOComponentName(rw.Name),
	}
}

func generateMinIODeployment(rw *v1alpha1.RisingWave) *v1.Deployment {
	spec := rw.Spec.ObjectStorage.MinIO

	var tag = "latest"
	if spec.Image.Tag != nil {
		tag = *spec.Image.Tag
	}

	container := corev1.Container{
		Name:            "minio",
		Resources:       *spec.Resources,
		Image:           fmt.Sprintf("%s:%s", *spec.Image.Repository, tag),
		Ports:           spec.Ports,
		ImagePullPolicy: *spec.Image.PullPolicy,
		Env: []corev1.EnvVar{
			{
				Name:  "MINIO_SERVER_PORT",
				Value: fmt.Sprintf("%d", v1alpha1.MinIOServerPort),
			},
			{
				Name:  "MINIO_CONSOLE_PORT",
				Value: fmt.Sprintf("%d", v1alpha1.MinIOConsolePort),
			},
		},
		Command: []string{"/bin/bash", "-c"},
		Args: []string{
			"${PREFIX_BIN}/set_minio.sh",
		},
		ReadinessProbe: &corev1.Probe{
			Handler: corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"/bin/bash",
						"-c",
						"ls",
						"${PREFIX_LOG}/minio_server_ready",
					},
				},
			},
		},
		LivenessProbe: &corev1.Probe{
			// URL: curl -I http://127.0.0.1:9301/minio/health/live
			Handler: corev1.Handler{
				HTTPGet: &corev1.HTTPGetAction{
					Port: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: v1alpha1.MinIOServerPort,
					},
					Path: "/minio/health/live",
				},
			},
		},
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

	// TODO(KexiangWang): Add some check for `spec.Replicas`
	// Currently minio is coped as in-memory storage, which is not distributed,
	// so the `spec.Replicas` value should always be 1.
	var deploy = &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      MinIOComponentName(rw.Name),
		},
		Spec: v1.DeploymentSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					RisingWaveKey:  RisingWaveMinIOValue,
					RisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						RisingWaveKey:  RisingWaveMinIOValue,
						RisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return deploy
}

func generateMinIOService(rw *v1alpha1.RisingWave) *corev1.Service {
	spec := corev1.ServiceSpec{
		Type: corev1.ServiceTypeClusterIP,
		Selector: map[string]string{
			RisingWaveKey:  RisingWaveMinIOValue,
			RisingWaveName: rw.Name,
		},
	}

	var ports []corev1.ServicePort
	for _, p := range rw.Spec.ObjectStorage.MinIO.Ports {
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
			Name:      MinIOComponentName(rw.Name),
		},
		Spec: spec,
	}
	return s
}
