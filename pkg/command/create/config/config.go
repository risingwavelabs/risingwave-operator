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

package config

import (
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

const (
	Image = "ghcr.io/risingwavelabs/risingwave"

	DefaultLimitCPU      = "1"
	DefaultLimitMemory   = "1Gi"
	DefaultRequestCPU    = "100m"
	DefaultRequestMemory = "100Mi"
)

// Config contain the fields needed that creating a instance.
type Config struct {
	BaseConfig

	MetaConfig      ComponentConfig
	ComputeConfig   ComponentConfig
	CompactorConfig ComponentConfig
	FrontendConfig  ComponentConfig
}

type ComponentConfig struct {
	Groups []Group
}

type Group struct {
	Name      string
	Replicas  int32
	Resources corev1.ResourceRequirements
}

func (g Group) deepCopy() Group {
	return Group{
		Name:      g.Name,
		Replicas:  g.Replicas,
		Resources: *g.Resources.DeepCopy(),
	}
}

type BaseConfig struct {
	Image    string
	Replicas int32

	Resources corev1.ResourceRequirements
}

var defaultResource = corev1.ResourceRequirements{
	Limits: map[corev1.ResourceName]k8sresource.Quantity{
		corev1.ResourceCPU:    k8sresource.MustParse(DefaultLimitCPU),
		corev1.ResourceMemory: k8sresource.MustParse(DefaultLimitMemory),
	},
	Requests: map[corev1.ResourceName]k8sresource.Quantity{
		corev1.ResourceCPU:    k8sresource.MustParse(DefaultRequestCPU),
		corev1.ResourceMemory: k8sresource.MustParse(DefaultRequestMemory),
	}}

var defaultGroup = Group{
	Name:      "default",
	Replicas:  1,
	Resources: *defaultResource.DeepCopy(),
}

var DefaultConfig = Config{
	BaseConfig: BaseConfig{
		Image:     Image,
		Resources: *defaultResource.DeepCopy(),
	},
	MetaConfig: ComponentConfig{
		Groups: []Group{
			defaultGroup.deepCopy(),
		},
	},
	ComputeConfig: ComponentConfig{
		Groups: []Group{
			defaultGroup.deepCopy(),
		},
	},
	CompactorConfig: ComponentConfig{
		Groups: []Group{
			defaultGroup.deepCopy(),
		},
	},
	FrontendConfig: ComponentConfig{
		Groups: []Group{
			defaultGroup.deepCopy(),
		},
	},
}

// ApplyConfigFile will construct a config by config file.
func ApplyConfigFile(path string) (Config, error) {
	DefaultConfig.Image = Image
	if len(path) == 0 {
		return DefaultConfig, nil
	}

	c, err := parse(path)
	if err != nil {
		return DefaultConfig, err
	}

	return constructConfig(c), nil
}

func constructConfig(c innerConfig) Config {
	var conf = Config{
		BaseConfig: BaseConfig{
			Image:     Image,
			Replicas:  int32(c.Global.Replicas),
			Resources: constructResource(c.Global.Limit, c.Global.Request),
		},
		ComputeConfig: ComponentConfig{
			Groups: constructGroup(c.Compute.Groups),
		},
		CompactorConfig: ComponentConfig{
			Groups: constructGroup(c.Compactor.Groups),
		},
		FrontendConfig: ComponentConfig{
			Groups: constructGroup(c.Frontend.Groups),
		},
		MetaConfig: ComponentConfig{
			Groups: constructGroup(c.Meta.Groups),
		},
	}
	return conf
}

func constructGroup(groups []group) []Group {
	var newGroups []Group
	for _, g := range groups {
		var r int32
		if g.Replicas == 0 {
			r = 1
		} else {
			r = int32(g.Replicas)
		}
		var newG = Group{
			Name:      g.Name,
			Replicas:  r,
			Resources: constructResource(g.Limit, g.Request),
		}

		newGroups = append(newGroups, newG)
	}
	return newGroups
}

func constructResource(limit, request resource) corev1.ResourceRequirements {
	var cpuLimit = limit.CPU
	if len(limit.CPU) == 0 {
		cpuLimit = DefaultLimitCPU
	}

	var lMemory = limit.Memory
	if len(limit.Memory) == 0 {
		lMemory = DefaultLimitMemory
	}
	var cpuRequest = request.CPU
	if len(request.CPU) == 0 {
		cpuRequest = DefaultRequestCPU
	}
	var rMemory = request.Memory
	if len(request.Memory) == 0 {
		rMemory = DefaultRequestMemory
	}
	return corev1.ResourceRequirements{
		Limits: map[corev1.ResourceName]k8sresource.Quantity{
			corev1.ResourceCPU:    k8sresource.MustParse(cpuLimit),
			corev1.ResourceMemory: k8sresource.MustParse(lMemory),
		},
		Requests: map[corev1.ResourceName]k8sresource.Quantity{
			corev1.ResourceCPU:    k8sresource.MustParse(cpuRequest),
			corev1.ResourceMemory: k8sresource.MustParse(rMemory),
		},
	}
}
