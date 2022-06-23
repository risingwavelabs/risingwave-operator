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

package rendor

import (
	"bytes"
	"fmt"

	"k8s.io/cli-runtime/pkg/resource"
)

var parser *Parser

type Parser struct {
	Builder *resource.Builder
}

func NewParser(clientGetter resource.RESTClientGetter) {
	p := &Parser{
		Builder: resource.NewBuilder(clientGetter).Unstructured(),
	}

	parser = p
}

func CreateObjectByTem(path string, obj interface{}) error {
	infos, err := ParseFile(path, obj)
	if err != nil {
		return fmt.Errorf("parse file failed, %w", err)
	}
	for i := range infos {
		err := createObject(infos[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseFile(path string, obj interface{}) ([]*resource.Info, error) {
	data, err := Template(path, obj)
	if err != nil {
		return nil, err
	}
	r := parser.Builder.Stream(bytes.NewReader(data), "parser").Do()
	if r.Err() != nil {
		return nil, r.Err()
	}
	infos, err := r.Infos()
	if err != nil {
		return nil, err
	}
	return infos, nil
}

func createObject(info *resource.Info) error {
	h := resource.NewHelper(info.Client, info.Mapping)
	_, err := h.Create(info.Namespace, true, info.Object)
	if err != nil {
		return err
	}
	return nil
}
