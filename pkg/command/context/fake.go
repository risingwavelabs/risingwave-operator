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

package context

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/pkg/command/helper"
)

var FakerContext = &FakeContext{
	namespace: "test",
}

type FakeContext struct {
	namespace string
	client    client.Client
}

func (f *FakeContext) SetNamespace(ns string) {
	f.namespace = ns
}

func (f *FakeContext) Scheme() *runtime.Scheme {
	//TODO implement me
	panic("implement me")
}

func (f *FakeContext) Namespace() string {
	return f.namespace
}

func (f *FakeContext) Builder() *resource.Builder {
	//TODO implement me
	panic("implement me")
}

func (f *FakeContext) Client() client.Client {
	return f.client
}

func (f *FakeContext) SetClient(c client.Client) {
	f.client = c
}

func (f *FakeContext) Applier() *helper.Applier {
	//TODO implement me
	panic("implement me")
}

var _ Context = &FakeContext{}
