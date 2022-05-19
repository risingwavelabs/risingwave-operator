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

package options

import (
	"io/ioutil"
	"reflect"
	"strings"
	"unsafe"

	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
)

type ResourceList map[corev1.ResourceName]string

type InnerBaseOptions struct {
	Image map[v1alpha1.Arch]v1alpha1.ImageOptions `yaml:"image"`

	Replicas   int32             `yaml:"replicas"`
	PullPolicy corev1.PullPolicy `yaml:"pullPolicy"`
	Limits     ResourceList      `yaml:"limits"`
	Requests   ResourceList      `yaml:"requests"`

	innerResources *corev1.ResourceRequirements
}

func (b *InnerBaseOptions) Unmarshal() {
	b.innerResources = &corev1.ResourceRequirements{}
	if len(b.Limits) != 0 {
		b.innerResources.Limits = make(map[corev1.ResourceName]resource.Quantity)
		for n, v := range b.Limits {
			b.innerResources.Limits[n] = resource.MustParse(v)
		}
	}

	if len(b.Requests) != 0 {
		b.innerResources.Requests = make(map[corev1.ResourceName]resource.Quantity)
		for n, v := range b.Requests {
			b.innerResources.Requests[n] = resource.MustParse(v)
		}
	}
}

func (b *InnerBaseOptions) DeepCopy() *InnerBaseOptions {
	return &InnerBaseOptions{
		Image:          copyImage(b.Image),
		PullPolicy:     b.PullPolicy,
		Replicas:       b.Replicas,
		innerResources: b.innerResources.DeepCopy(),
	}
}

func (b *InnerBaseOptions) DeepCopyToBaseOptions(arch v1alpha1.Arch) v1alpha1.BaseOptions {

	return v1alpha1.BaseOptions{
		Image:      copyImage(b.Image),
		PullPolicy: b.PullPolicy,
		Replicas:   b.Replicas,
		Resources:  *b.innerResources.DeepCopy(),
	}
}

func copyImage(im map[v1alpha1.Arch]v1alpha1.ImageOptions) map[v1alpha1.Arch]v1alpha1.ImageOptions {
	var image = make(map[v1alpha1.Arch]v1alpha1.ImageOptions)
	for k, v := range im {
		image[k] = v
	}
	return image
}

type InnerRisingWaveOptions struct {
	Arch v1alpha1.Arch

	Default       *InnerBaseOptions `yaml:"default"`
	MetaNode      *InnerBaseOptions `yaml:"metaNode"`
	ComputeNode   *InnerBaseOptions `yaml:"computeNode"`
	CompactorNode *InnerBaseOptions `yaml:"compactorNode"`
	MinIO         *InnerBaseOptions `yaml:"minIO"`
	Frontend      *InnerBaseOptions `yaml:"frontend"`
}

func (o *InnerRisingWaveOptions) BuildConfigFromFile(filePath string) error {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = o.unmarshal(data)
	if err != nil {
		return err
	}
	o.install()
	v1alpha1.SetDefaultOption(v1alpha1.RisingWaveOptions{
		Arch:          o.Arch,
		MetaNode:      o.MetaNode.DeepCopyToBaseOptions(o.Arch),
		MinIO:         o.MinIO.DeepCopyToBaseOptions(o.Arch),
		Frontend:      o.Frontend.DeepCopyToBaseOptions(o.Arch),
		ComputeNode:   o.ComputeNode.DeepCopyToBaseOptions(o.Arch),
		CompactorNode: o.CompactorNode.DeepCopyToBaseOptions(o.Arch),
	})

	return nil
}

func (o *InnerRisingWaveOptions) unmarshal(in []byte) error {
	err := yaml.Unmarshal(in, o)
	if err != nil {
		return err
	}
	t := reflect.ValueOf(*o)
	l := t.NumField()
	for i := 0; i < l; i++ {
		inner := t.Field(i)
		// find BaseOptions
		if inner.Type().Kind() == reflect.Ptr && inner.Type().Elem().Name() == "InnerBaseOptions" {
			if inner.Pointer() == uintptr(unsafe.Pointer(nil)) {
				continue
			}
			inner.MethodByName("Unmarshal").Call([]reflect.Value{})
		}
	}
	return nil
}

func (o *InnerRisingWaveOptions) install() {
	if len(o.Arch) == 0 {
		o.Arch = v1alpha1.AMD64Arch
	}

	for k, v := range o.Default.Image {
		o.Default.Image[k] = v1alpha1.ImageOptions{
			Repository: replaceArch(v.Repository, string(k)),
			Tag:        replaceArch(v.Tag, string(k)),
		}
	}

	o.MetaNode = o.installCommonDefaultValue(o.MetaNode)
	o.ComputeNode = o.installCommonDefaultValue(o.ComputeNode)
	o.MinIO = o.installCommonDefaultValue(o.MinIO)
	o.Frontend = o.installCommonDefaultValue(o.Frontend)
	o.CompactorNode = o.installCommonDefaultValue(o.CompactorNode)
}

func (o *InnerRisingWaveOptions) installCommonDefaultValue(targetOption *InnerBaseOptions) *InnerBaseOptions {
	defaultOption := o.Default
	if targetOption == nil {
		targetOption = defaultOption.DeepCopy()
		return targetOption
	}

	if targetOption.Image == nil {
		targetOption.Image = copyImage(o.Default.Image)
	}

	if targetOption.Replicas == 0 {
		targetOption.Replicas = defaultOption.Replicas
	}
	if len(targetOption.PullPolicy) == 0 {
		targetOption.PullPolicy = defaultOption.PullPolicy
	}
	if targetOption.innerResources == nil {
		targetOption.innerResources = defaultOption.innerResources.DeepCopy()
	} else {
		if targetOption.innerResources.Limits == nil {
			targetOption.innerResources.Limits = defaultOption.innerResources.Limits.DeepCopy()
		}
		if targetOption.innerResources.Requests == nil {
			targetOption.innerResources.Requests = defaultOption.innerResources.Requests.DeepCopy()
		}
	}

	for k, v := range targetOption.Image {
		v.Repository = replaceArch(v.Repository, string(k))
		v.Tag = replaceArch(v.Tag, string(k))
		targetOption.Image[k] = v
	}

	return targetOption
}

func replaceArch(source, target string) string {
	return strings.Replace(source, "${ARCH}", target, 1)
}
