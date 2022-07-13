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
	"fmt"
	"os"
	"sync"

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/command/helper"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
}

// RWContext wraps the configuration and credential for tidb cluster accessing.
type RWContext struct {
	*genericclioptions.ConfigFlags

	lock sync.Mutex

	c client.Client

	b *resource.Builder

	restConfig *rest.Config

	a *helper.Applier
}

// we need put client init code when get func called, because the flag not parsed when NewContext

func NewContext(f *genericclioptions.ConfigFlags) *RWContext {
	o := &RWContext{
		ConfigFlags: f,
	}
	return o
}

func (o *RWContext) Scheme() *runtime.Scheme {
	return scheme
}

func (o *RWContext) Builder() *resource.Builder {
	if o.b == nil {
		o.lazyInit()
	}
	return o.b
}

func (o *RWContext) Namespace() string {
	return *o.ConfigFlags.Namespace
}

func (o *RWContext) Client() client.Client {
	if o.b == nil {
		o.lazyInit()
	}
	return o.c
}

func (o *RWContext) getKubeClient() (client.Client, error) {
	fmt.Println(*o.KubeConfig)
	config, err := o.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	o.restConfig = config

	c, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, err
	}

	return NewBypassClient(c), nil
}

func (o *RWContext) lazyInit() {
	o.lock.Lock()
	defer o.lock.Unlock()

	c, err := o.getKubeClient()
	if err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
	o.c = c
	o.b = resource.NewBuilder(o)
	o.a = helper.NewApplier(o.restConfig)
}

func (o *RWContext) Applier() *helper.Applier {
	if o.a == nil {
		o.lazyInit()
	}
	return o.a
}
