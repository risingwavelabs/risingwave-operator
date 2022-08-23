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

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Update the CPU and memory configuration for risingwave instances.

Limits and requests for CPU resources are measured in cpu units while that of memory resources are measured in bytes.

Accepted values for resources:

  CPU: 	  Plain integer or using millicpu. For example, 1.0 or 100m, these are equivalent.

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

type Options struct {
	*cmdcontext.BasicOptions

	computeRequest Request
	computeLimit   Request
	memoryRequest  Request
	memoryLimit    Request
	group          string
	component      string
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
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.ExitOnErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVar(&o.computeRequest.requestedQty, "cpu-request", "", "The target cpu request.")
	cmd.Flags().StringVar(&o.computeLimit.requestedQty, "cpu-limit", "", "The target cpu limit.")
	cmd.Flags().StringVar(&o.memoryRequest.requestedQty, "memory-request", "", "The target memory request.")
	cmd.Flags().StringVar(&o.memoryLimit.requestedQty, "memory-limit", "", "The target memory limit.")
	cmd.Flags().StringVarP(&o.group, "group", "g", util.DefaultGroup, "The group to be updated. If not set, update the default group.")
	cmd.Flags().StringVarP(&o.component, "component", "c", util.GLOBAL, "The component to be updated. If not set, update global resources.")

	return cmd
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	// parse the resource requests
	if o.computeRequest.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.computeRequest.requestedQty)
		if err != nil {
			return err
		}
		o.computeRequest.convertedQty = qty
	}

	if o.computeLimit.requestedQty != "" {
		qty, err := k8sresource.ParseQuantity(o.computeLimit.requestedQty)
		if err != nil {
			return err
		}
		o.computeLimit.convertedQty = qty
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

	// validate group
	if o.group == "" {
		o.group = util.DefaultGroup
	} else {
		rw, err := o.GetRwInstance(ctx)
		if err != nil {
			return err
		}

		if !util.IsValidGroup(rw, o.group) {
			return fmt.Errorf("invalid group name %s", o.group)
		}
	}

	// validate component
	if o.component == "" {
		return fmt.Errorf("component name is required")
	}

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return err
	}

	o.updateConfig(rw)

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %v", err)
	}

	return nil
}

func (o *Options) updateConfig(rw *v1alpha1.RisingWave) error {
	components := &rw.Spec.Components

	switch o.component {
	case util.COMPUTE:
		for _, group := range components.Compute.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.META:
		for _, group := range components.Meta.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.COMPACTOR:
		for _, group := range components.Compactor.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.FRONTEND:
		for _, group := range components.Frontend.Groups {
			if group.Name == o.group {
				o.updateComponentResources(&group.Resources)
				break
			}
		}

	case util.GLOBAL:
		o.updateComponentResources(&rw.Spec.Global.Resources)

	default:
		return fmt.Errorf("invalid component name %s, will do nothing", o.component)
	}

	return nil
}

func (o *Options) updateComponentResources(resourceMap *corev1.ResourceRequirements) {
	if o.computeRequest.requestedQty != "" {
		resourceMap.Requests[corev1.ResourceCPU] = o.computeRequest.convertedQty
	}

	if o.computeLimit.requestedQty != "" {
		resourceMap.Limits[corev1.ResourceCPU] = o.computeLimit.convertedQty
	}

	if o.memoryRequest.requestedQty != "" {
		resourceMap.Requests[corev1.ResourceMemory] = o.memoryRequest.convertedQty
	}

	if o.memoryLimit.requestedQty != "" {
		resourceMap.Limits[corev1.ResourceMemory] = o.memoryLimit.convertedQty
	}
}
