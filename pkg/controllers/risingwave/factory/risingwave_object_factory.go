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

func (f *RisingWaveObjectFactory) NewMetaService() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: f.risingwave.Namespace,
			Name:      f.risingwave.Name + "-meta",
			Labels: map[string]string{
				consts.LabelRisingWaveUID:        string(f.risingwave.UID),
				consts.LabelRisingWaveGeneration: strconv.FormatInt(f.risingwave.Generation, 10),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Selector: map[string]string{
				consts.LabelRisingWaveRole: consts.MetaNode,
				consts.LabelRisingWaveName: f.risingwave.Name,
			},
			Ports: lo.Map(f.risingwave.Spec.MetaNode.Ports, func(p corev1.ContainerPort, _ int) corev1.ServicePort {
				return corev1.ServicePort{
					Protocol:   corev1.ProtocolTCP,
					Port:       p.ContainerPort,
					TargetPort: intstr.FromInt(int(p.ContainerPort)),
					Name:       p.Name,
				}
			}),
		},
	}

	return mustSetControllerReference(f.risingwave, service, f.scheme)
}

func (f *RisingWaveObjectFactory) NewMetaDeployment() *appsv1.Deployment {
	rw := f.risingwave

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
			"--host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaServerPort),
			"--dashboard-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaDashboardPort),
			"--prometheus-host",
			fmt.Sprintf("0.0.0.0:%d", v1alpha1.MetaMetricsPort),
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

	deploy := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: rw.Namespace,
			Name:      rw.Name + "-meta",
		},
		Spec: v1.DeploymentSpec{
			Replicas: spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					consts.LabelRisingWaveRole: consts.MetaNode,
					consts.LabelRisingWaveName: rw.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						consts.LabelRisingWaveRole: consts.MetaNode,
						consts.LabelRisingWaveName: rw.Name,
					},
				},
				Spec: podSpec,
			},
		},
	}

	return mustSetControllerReference(f.risingwave, deploy, f.scheme)
}

func (f *RisingWaveObjectFactory) NewFrontendService() *corev1.Service {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewFrontendDeployment() *appsv1.Deployment {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewComputeService() *corev1.Service {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewComputeDeployment() *appsv1.Deployment {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewCompactorService() *corev1.Service {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewCompactorDeployment() *appsv1.Deployment {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewMinIOService() *corev1.Service {
	panic("unimplemented")
}

func (f *RisingWaveObjectFactory) NewMinIODeployment() *appsv1.Deployment {
	panic("unimplemented")
}

func NewRisingWaveObjectFactory(risingwave *risingwavev1alpha1.RisingWave, scheme *runtime.Scheme) *RisingWaveObjectFactory {
	return &RisingWaveObjectFactory{
		risingwave: risingwave,
		scheme:     scheme,
	}
}
