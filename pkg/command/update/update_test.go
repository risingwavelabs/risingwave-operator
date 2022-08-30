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

package update

import (
	goctx "context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/stretchr/testify/assert"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util"
)

var ctx = context.Fake

func TestOptions_Validate(t *testing.T) {
	var o = Options{
		BasicOptions: &context.BasicOptions{},
		component:    "fake-component",
	}
	err := o.Validate(ctx, nil, []string{})
	assert.Equal(t, err.Error(), "component should be in [compactor,compute,frontend,meta,global]")

	createTestInstance(o, t)

	o.component = util.Meta
	o.group = util.DefaultGroup
	o.memoryLimit.requestedQty = "256Mi"
	err = o.Validate(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, o.memoryLimit.convertedQty, resource.MustParse("256Mi"))

	o.group = "fake-group"
	for k, _ := range componentSet {
		o.component = k
		err = o.Validate(ctx, nil, []string{})
		ss := fmt.Sprintf("invalid risingwave group: %s for component: %s", o.group, o.component)
		assert.Equal(t, err.Error(), ss)
	}
}

func TestOptions_Run(t *testing.T) {
	// test update DefaultGroup
	var o = Options{
		BasicOptions: &context.BasicOptions{},
		memoryLimit: Request{
			requestedQty: "256Mi",
		},
		memoryRequest: Request{
			requestedQty: "128Mi",
		},
		cpuLimit: Request{
			requestedQty: "1000m",
		},
		cpuRequest: Request{
			requestedQty: "200m",
		},
	}

	createTestInstance(o, t)

	o.component = util.Global
	o.group = util.DefaultGroup
	err := o.Validate(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = o.Run(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	var fakeRW = util.FakeRW()
	var rw = v1alpha1.RisingWave{}
	err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: fakeRW.Namespace, Name: fakeRW.Name}, &rw)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, rw.Namespace, rw.Namespace)
	assert.Equal(t, len(rw.Spec.Components.Frontend.Groups), 3)

	var globalGroup v1alpha1.RisingWaveComponentGroup
	for _, g := range rw.Spec.Components.Frontend.Groups {
		if g.Name == util.DefaultGroup {
			globalGroup = g
			break
		}
	}
	assert.Equal(t, globalGroup.Resources.Limits[corev1.ResourceMemory], resource.MustParse("256Mi"))

	var cg v1alpha1.RisingWaveComputeGroup
	for _, g := range rw.Spec.Components.Compute.Groups {
		if g.Name == util.DefaultGroup {
			cg = g
			break
		}
	}
	assert.Equal(t, cg.Resources.Requests[corev1.ResourceCPU], resource.MustParse("200m"))
	ctx.Client().Delete(goctx.Background(), &rw)

	// test update meta resource
	o.component = util.Meta
	o.group = fakeRW.Spec.Components.Meta.Groups[0].Name
	createTestInstance(o, t)
	err = o.Validate(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = o.Run(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: fakeRW.Namespace, Name: fakeRW.Name}, &rw)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(rw.Spec.Components.Meta.Groups), 2)

	var metaG v1alpha1.RisingWaveComponentGroup
	for _, g := range rw.Spec.Components.Meta.Groups {
		if g.Name == o.group {
			metaG = g
			break
		}
	}
	assert.Equal(t, metaG.Resources.Limits[corev1.ResourceMemory], resource.MustParse("256Mi"))
	assert.Equal(t, metaG.Resources.Requests[corev1.ResourceCPU], resource.MustParse("200m"))

	// test update continuously
	o.cpuLimit.requestedQty = "3333m"
	_ = o.Validate(ctx, nil, []string{})
	_ = o.Run(ctx, nil, []string{})
	_ = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: fakeRW.Namespace, Name: fakeRW.Name}, &rw)

	for _, g := range rw.Spec.Components.Meta.Groups {
		if g.Name == o.group {
			metaG = g
			break
		}
	}
	assert.Equal(t, metaG.Resources.Limits[corev1.ResourceCPU], resource.MustParse("3333m"))
}

func createTestInstance(o Options, t *testing.T) {
	rw := util.FakeRW()
	var ns = rw.Namespace
	var name = rw.Name
	ctx.SetNamespace(ns)
	err := o.Complete(ctx, nil, []string{name})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, o.Namespace, ns)
	assert.Equal(t, o.Name, name)

	ctx.SetClient(util.NewFakeClient())
	err = ctx.Client().Create(goctx.Background(), rw)
	if err != nil {
		t.Fatal(err)
	}
}
