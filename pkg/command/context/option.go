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
	"sync"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RWContext wraps the configuration and credential for tidb cluster accessing.
type RWContext struct {
	*genericclioptions.ConfigFlags

	lock sync.Mutex

	c client.Client
}

func NewContext(f *genericclioptions.ConfigFlags) *RWContext {
	return &RWContext{
		ConfigFlags: f,
	}
}

func (o *RWContext) Builder() *resource.Builder {
	return resource.NewBuilder(o)
}

func (o *RWContext) Namespace() string {
	return *o.ConfigFlags.Namespace
}
