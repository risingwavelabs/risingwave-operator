/*
 * Copyright 2023 RisingWave Labs
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
	"reflect"
	"strings"
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
	assert.Equal(t, err.Error(), "component should be in [compactor,compute,connector,frontend,meta,global]")

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
	for k := range componentSet {
		o.component = k
		err = o.Validate(ctx, nil, []string{})
		ss := fmt.Sprintf("invalid risingwave group: %s for component: %s", o.group, o.component)
		assert.Equal(t, err.Error(), ss)
	}

	o = Options{
		BasicOptions: &context.BasicOptions{},
		component:    util.Meta,
		group:        util.DefaultGroup,
		cpuRequest:   Request{requestedQty: "a123"},
	}
	err = o.Validate(ctx, nil, []string{})
	assert.Equal(t, err == nil, false)

	o = Options{
		BasicOptions: &context.BasicOptions{},
		component:    util.Meta,
		group:        util.DefaultGroup,
		cpuLimit:     Request{requestedQty: "a123"},
	}
	err = o.Validate(ctx, nil, []string{})
	assert.Equal(t, err == nil, false)

	o = Options{
		BasicOptions:  &context.BasicOptions{},
		component:     util.Meta,
		group:         util.DefaultGroup,
		memoryRequest: Request{requestedQty: "a123"},
	}
	err = o.Validate(ctx, nil, []string{})
	assert.Equal(t, err == nil, false)

	o = Options{
		BasicOptions: &context.BasicOptions{},
		component:    util.Meta,
		group:        util.DefaultGroup,
		memoryLimit:  Request{requestedQty: "a123"},
	}
	err = o.Validate(ctx, nil, []string{})
	assert.Equal(t, err == nil, false)
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

	// test global
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
	var rw = &v1alpha1.RisingWave{}
	err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: fakeRW.Namespace, Name: fakeRW.Name}, rw)
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
	ctx.Client().Delete(goctx.Background(), rw)

	// test compute
	o.component = util.Compute
	o.group = rw.Spec.Components.Compute.Groups[0].Name

	rw = createTestInstance(o, t)
	oldLen := len(rw.Spec.Components.Compute.Groups)
	err = o.Validate(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = o.Run(ctx, nil, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: rw.Namespace, Name: rw.Name}, rw)
	if err != nil {
		t.Fatal(err)
	}

	newGroups := rw.Spec.Components.Compute.Groups

	assert.Equal(t, len(newGroups), oldLen+1)

	for _, g := range newGroups {
		if g.Name == o.group {
			cg = g
			break
		}
	}
	assert.Equal(t, cg.Resources.Limits[corev1.ResourceMemory], resource.MustParse("256Mi"))
	assert.Equal(t, cg.Resources.Requests[corev1.ResourceCPU], resource.MustParse("200m"))
	ctx.Client().Delete(goctx.Background(), rw)

	// test other component
	testCases := []struct {
		component string
		groups    []v1alpha1.RisingWaveComponentGroup
	}{
		{
			util.Meta,
			fakeRW.Spec.Components.Meta.Groups,
		}, {
			util.Compactor,
			fakeRW.Spec.Components.Compactor.Groups,
		}, {
			util.Frontend,
			fakeRW.Spec.Components.Frontend.Groups,
		}, {
			util.Connector,
			fakeRW.Spec.Components.Connector.Groups,
		},
	}

	var doTest = func(component string, oldGroup []v1alpha1.RisingWaveComponentGroup) {
		defer ctx.Client().Delete(goctx.Background(), rw)

		o.component = component
		o.group = oldGroup[0].Name
		rw := createTestInstance(o, t)
		err = o.Validate(ctx, nil, []string{})
		if err != nil {
			t.Fatal(err)
		}

		err = o.Run(ctx, nil, []string{})
		if err != nil {
			t.Fatal(err)
		}

		err = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: rw.Namespace, Name: rw.Name}, rw)
		if err != nil {
			t.Fatal(err)
		}

		newGroups := reflectGroup(component, rw)

		assert.Equal(t, len(newGroups), len(oldGroup)+1)

		var cg v1alpha1.RisingWaveComponentGroup
		for _, g := range newGroups {
			if g.Name == o.group {
				cg = g
				break
			}
		}
		assert.Equal(t, cg.Resources.Limits[corev1.ResourceMemory], resource.MustParse("256Mi"))
		assert.Equal(t, cg.Resources.Requests[corev1.ResourceCPU], resource.MustParse("200m"))

		// test update continuously
		o.cpuLimit.requestedQty = "3333m"
		_ = o.Validate(ctx, nil, []string{})
		_ = o.Run(ctx, nil, []string{})
		_ = ctx.Client().Get(goctx.Background(), client.ObjectKey{Namespace: fakeRW.Namespace, Name: fakeRW.Name}, rw)

		for _, g := range reflectGroup(component, rw) {
			if g.Name == o.group {
				cg = g
				break
			}
		}
		assert.Equal(t, cg.Resources.Limits[corev1.ResourceCPU], resource.MustParse("3333m"))
	}

	for _, c := range testCases {
		doTest(c.component, c.groups)
	}
}

func reflectGroup(component string, rw *v1alpha1.RisingWave) []v1alpha1.RisingWaveComponentGroup {
	c := rw.Spec.Components
	v := reflect.ValueOf(c)
	nameValue := v.FieldByName(titleSize(component))
	g := nameValue.FieldByName("Groups")
	return g.Interface().([]v1alpha1.RisingWaveComponentGroup)
}

func titleSize(s string) string { return strings.ToTitle(s[:1]) + strings.ToLower(s[1:]) }

func createTestInstance(o Options, t *testing.T) *v1alpha1.RisingWave {
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
	return rw
}
