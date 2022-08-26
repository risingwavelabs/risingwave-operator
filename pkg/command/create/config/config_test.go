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
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"

	"github.com/stretchr/testify/assert"
)

func Test_ApplyConfigFile(t *testing.T) {
	var path = ""
	config, _ := ApplyConfigFile(path)
	assert.Equal(t, config, DefaultConfig)

	path = "fake.toml"
	_, err := ApplyConfigFile(path)
	assert.Equal(t, strings.Contains(err.Error(), "no such file or directory"), true)

	config, err = ApplyConfigFile("example.toml")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, config.Image, Image)
}

func Test_constructGroup(t *testing.T) {
	var groups = []group{
		{Replicas: 0},
	}
	g := constructGroup(groups)
	assert.Equal(t, g[0].Replicas, int32(1))
	groups[0].Replicas = 2
	g = constructGroup(groups)
	assert.Equal(t, g[0].Replicas, int32(2))
}

func Test_constructResource(t *testing.T) {
	var limit = resource{
		CPU:    "",
		Memory: "",
	}

	var request = resource{
		CPU:    "",
		Memory: "",
	}

	r := constructResource(limit, request)
	assert.Equal(t, r.Limits[corev1.ResourceCPU], k8sresource.MustParse(DefaultLimitCPU))
	assert.Equal(t, r.Limits[corev1.ResourceMemory], k8sresource.MustParse(DefaultLimitMemory))
	assert.Equal(t, r.Requests[corev1.ResourceCPU], k8sresource.MustParse(DefaultRequestCPU))
	assert.Equal(t, r.Requests[corev1.ResourceMemory], k8sresource.MustParse(DefaultRequestMemory))
}
