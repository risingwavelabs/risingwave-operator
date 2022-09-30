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

package create

import (
	goctx "context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/create/config"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util"
)

var ctx = context.Fake

func Test_Complete(t *testing.T) {
	o := NewOptions(genericclioptions.IOStreams{})
	ctx.SetNamespace("test-ns")
	o.Complete(ctx, nil, []string{"test-name"})

	assert.Equal(t, o.namespace, "test-ns")
	assert.Equal(t, o.name, "test-name")

	ctx.SetNamespace("")
	o.Complete(ctx, nil, []string{"test-name"})
	assert.Equal(t, o.namespace, "default")
}

func Test_CreateInstance(t *testing.T) {
	var o = Options{
		name:       "test",
		namespace:  "test-ns",
		configFile: "config/example.toml",
	}
	c, err := config.ApplyConfigFile(o.configFile)
	if err != nil {
		t.Fatal(err)
	}
	o.config = c
	rw, err := o.createInstance()
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rw.Name, "test")
	assert.Equal(t, rw.Spec.Components.Compactor.Groups[0].Name, "compactor-group-1")
}

func TestOptions_Validate(t *testing.T) {
	var o = Options{
		namespace: "test-ns",
	}
	err := o.Validate(ctx, nil, []string{})
	assert.Equal(t, err.Error(), "name should be set when using default config")
}

func TestOptions_Run(t *testing.T) {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	var o = Options{
		configFile: "config/example.toml",
		IOStreams:  streams,
	}
	ctx.SetNamespace("test-ns")
	err := o.Complete(ctx, nil, []string{"test"})
	if err != nil {
		t.Fatal(err)
	}
	rw, _ := o.createInstance()

	ctx.SetClient(util.NewFakeClient())
	err = o.Run(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	var risingwave = risingwavev1alpha1.RisingWave{}
	err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: rw.Namespace, Name: rw.Name}, &risingwave)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, risingwave.Namespace, rw.Namespace)
	assert.Equal(t, risingwave.Spec.Global.Resources, rw.Spec.Global.Resources)
	assert.Equal(t, len(risingwave.Spec.Components.Meta.Groups), 2)
}

func Test_NewCommand(t *testing.T) {
	streams := genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr}
	cmd := NewCommand(ctx, streams)
	assert.Equal(t, cmd.Use, "create")
}
