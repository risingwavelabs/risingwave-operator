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

	"github.com/risingwavelabs/risingwave-operator/pkg/command/helper"
)

// Dummy is the default dummy context.
var Dummy = &dummyContext{
	namespace: "test",
}

type dummyContext struct {
	namespace string
	client    client.Client
}

// SetNamespace implements the Context.
func (f *dummyContext) SetNamespace(ns string) {
	f.namespace = ns
}

// Scheme implements the Context.
func (f *dummyContext) Scheme() *runtime.Scheme {
	panic("not supported")
}

// Namespace implements the Context.
func (f *dummyContext) Namespace() string {
	return f.namespace
}

// Builder  implements the Context.
func (f *dummyContext) Builder() *resource.Builder {
	panic("not supported")
}

// Client implements the Context.
func (f *dummyContext) Client() client.Client {
	return f.client
}

// SetClient implements the Context.
func (f *dummyContext) SetClient(c client.Client) {
	f.client = c
}

// Applier implements the Context.
func (f *dummyContext) Applier() *helper.Applier {
	panic("not supported")
}

var _ Context = &dummyContext{}
