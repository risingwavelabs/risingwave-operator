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
	"context"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util/errors"
)

const (
	LongDesc = `
Update the CPU and memory configuration for risingwave instances.

Limits and requests for CPU resources are measured in cpu units while that of memory resources are measured in bytes.

Accepted values for resources:

  CPU: 	  Plain integer or using milli-cpu. For example, 1.0 or 100m, these are equivalent.

  Memory: Plain integer or as a fixed-point number using one of these quantity suffixes: E, P, T, G, M, k. You can 
          also use the power-of-two equivalents: Ei, Pi, Ti, Gi, Mi, Ki. For example, 1G, 1Gi, 1024M or 128974848.
`
	Example = `  # Update compute request and limit config of global component in risingwave named example-rw.
  kubectl rw update example-rw -cr 200m -cl 1000m

  # Update memory request of global component in risingwave named example-rw in namespace foo.
  kubectl rw update example-rw -n foo -mr 256Mi

  # Update memory request of meta component in risingwave named example-rw in namespace foo and group test.
  kubectl rw update example-rw -n foo -c meta -g test -mr 256Mi
`
)

var componentSet = sets.NewString(util.Compute, util.Compactor, util.Meta, util.Frontend)

type Options struct {
	*cmdcontext.BasicOptions

	cpuRequest    Request
	cpuLimit      Request
	memoryRequest Request
	memoryLimit   Request
	group         string
	component     string
}

// Request holds the request string and converted resource quantity.
type Request struct {
	requestedQty string
	convertedQty k8sresource.Quantity
}

func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := Options{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update CPU and memory configuration for risingwave instances",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			errors.CheckErr(o.Complete(ctx, cmd, args))
			errors.ExitOnErr(o.Validate(ctx, cmd, args))
			errors.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVar(&o.cpuRequest.requestedQty, "cpu-request", "", "The target cpu request.")
	cmd.Flags().StringVar(&o.cpuLimit.requestedQty, "cpu-limit", "", "The target cpu limit.")
	cmd.Flags().StringVar(&o.memoryRequest.requestedQty, "memory-request", "", "The target memory request.")
	cmd.Flags().StringVar(&o.memoryLimit.requestedQty, "memory-limit", "", "The target memory limit.")
	cmd.Flags().StringVarP(&o.component, "component", "c", util.Global, "The component to be updated. If not set, update global resources.")
	cmd.Flags().StringVarP(&o.group, "group", "g", util.DefaultGroup, "The group to be updated. If not set, update the default group.")

	return cmd
}

func (o *Options) Validate(ctx cmdcontext.Context, cmd *cobra.Command, args []string) error {
	// validate component
	if o.component != util.Global && !componentSet.Has(strings.ToLower(o.component)) {
		return fmt.Errorf("component should be in [%s,global]", strings.Join(componentSet.List(), ","))
	}

	// validate group
	rw, err := o.GetRWInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	// if set DefaultGroup, not need check the name.
	if o.group == util.DefaultGroup {
		return o.validateResources()
	}

	switch o.component {
	case util.Meta:
		if !util.IsValidRWGroup(o.group, rw.Spec.Components.Meta.Groups) {
			return fmt.Errorf("invalid risingwave group: %s for component: %s", o.group, o.component)
		}

	case util.Frontend:
		if !util.IsValidRWGroup(o.group, rw.Spec.Components.Frontend.Groups) {
			return fmt.Errorf("invalid risingwave group: %s for component: %s", o.group, o.component)
		}

	case util.Compactor:
		if !util.IsValidRWGroup(o.group, rw.Spec.Components.Compactor.Groups) {
			return fmt.Errorf("invalid risingwave group: %s for component: %s", o.group, o.component)
		}

	case util.Compute:
		if !util.IsValidComputeGroup(o.group, rw.Spec.Components.Compute.Groups) {
			return fmt.Errorf("invalid risingwave group: %s for component: %s", o.group, o.component)
		}
	}

	return o.validateResources()
}

func (o *Options) validateResources() error {
	// parse the resource requests
	if o.cpuRequest.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.cpuRequest.requestedQty)
		if err != nil {
			return err
		}
		o.cpuRequest.convertedQty = qty
	}

	if o.cpuLimit.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.cpuLimit.requestedQty)
		if err != nil {
			return err
		}
		o.cpuLimit.convertedQty = qty
	}

	if o.memoryRequest.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.memoryRequest.requestedQty)
		if err != nil {
			return err
		}
		o.memoryRequest.convertedQty = qty
	}

	if o.memoryLimit.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.memoryLimit.requestedQty)
		if err != nil {
			return err
		}
		o.memoryLimit.convertedQty = qty
	}
	return nil
}
func (o *Options) Run(ctx cmdcontext.Context, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRWInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	// convert the global config into the components.
	rw = util.ConvertRisingwave(rw)

	o.updateConfig(rw)

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %v", err)
	}

	return nil
}

func (o *Options) updateConfig(rw *v1alpha1.RisingWave) {
	components := &rw.Spec.Components

	// only change the global resources
	if o.component == util.Global {
		o.updateInnerGlobalResources(rw)
		return
	}

	switch o.component {
	case util.Compute:
		for _, group := range components.Compute.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.Meta:
		for _, group := range components.Meta.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.Compactor:
		for _, group := range components.Compactor.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.Frontend:
		for _, group := range components.Frontend.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}
	}

	return
}

func (o *Options) updateInnerGlobalResources(rw *v1alpha1.RisingWave) {
	for _, group := range rw.Spec.Components.Compute.Groups {
		if group.Name == util.DefaultGroup {
			o.updateComponentResources(&group.Resources)
			break
		}
	}

	var componentGroups = [][]v1alpha1.RisingWaveComponentGroup{
		rw.Spec.Components.Meta.Groups,
		rw.Spec.Components.Frontend.Groups,
		rw.Spec.Components.Compactor.Groups,
	}

	var addResource = func(groups []v1alpha1.RisingWaveComponentGroup) {
		for i := range groups {
			if groups[i].Name == util.DefaultGroup {
				o.updateComponentResources(&groups[i].Resources)
				break
			}
		}
	}

	for _, groups := range componentGroups {
		addResource(groups)
	}
}

func (o *Options) updateComponentResources(resourceMap *corev1.ResourceRequirements) {
	if resourceMap.Requests == nil {
		resourceMap.Requests = make(corev1.ResourceList)
	}
	if resourceMap.Limits == nil {
		resourceMap.Limits = make(corev1.ResourceList)
	}

	if o.cpuRequest.requestedQty != "" {
		resourceMap.Requests[corev1.ResourceCPU] = o.cpuRequest.convertedQty
	}

	if o.cpuLimit.requestedQty != "" {
		resourceMap.Limits[corev1.ResourceCPU] = o.cpuLimit.convertedQty
	}

	if o.memoryRequest.requestedQty != "" {
		resourceMap.Requests[corev1.ResourceMemory] = o.memoryRequest.convertedQty
	}

	if o.memoryLimit.requestedQty != "" {
		resourceMap.Limits[corev1.ResourceMemory] = o.memoryLimit.convertedQty
	}
}
