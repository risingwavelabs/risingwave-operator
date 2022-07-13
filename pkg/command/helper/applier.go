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

package helper

import (
	"bytes"
	"io/ioutil"

	"k8s.io/client-go/rest"

	"k8s.io/cli-runtime/pkg/resource"
)

type Applier struct {
	builder *resource.Builder
}

func NewApplier(config *rest.Config) *Applier {
	i := &InnerRESTClientGetter{
		Config: config,
	}
	a := &Applier{
		builder: resource.NewBuilder(i).Unstructured(),
	}
	return a
}

func (a *Applier) Apply(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	r := a.builder.Stream(bytes.NewReader(data), "parser").Do()
	if r.Err() != nil {
		return r.Err()
	}
	infos, err := r.Infos()
	if err != nil {
		return err
	}

	for _, info := range infos {
		err = createObject(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *Applier) UnApply(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	r := a.builder.Stream(bytes.NewReader(data), "parser").Do()
	if r.Err() != nil {
		return r.Err()
	}
	infos, err := r.Infos()
	if err != nil {
		return err
	}

	for _, info := range infos {
		err = deleteObject(info)
		if err != nil {
			return err
		}
	}
	return nil
}

func deleteObject(info *resource.Info) error {
	h := resource.NewHelper(info.Client, info.Mapping)
	ns := info.Namespace
	if len(ns) == 0 {
		ns = "default"
	}
	_, err := h.Delete(ns, info.Name)
	if err != nil {
		return err
	}
	return nil
}

func createObject(info *resource.Info) error {
	h := resource.NewHelper(info.Client, info.Mapping)
	ns := info.Namespace
	if len(ns) == 0 {
		ns = "default"
	}
	_, err := h.Create(ns, true, info.Object)
	if err != nil {
		return err
	}
	return nil
}
