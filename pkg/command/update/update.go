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
	"k8s.io/cli-runtime/pkg/genericclioptions"

	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"

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
	Example = `  # Update compute request and limit config of risingwave named example-rw to 200m and 1000m
  kubectl rw update example-rw -cr 200m -cl 1000m

  # Update memory request config of risingwave named example-rw in namespace foo to 256Mi
  kubectl rw update example-rw -n foo -mr 256Mi

  # Update memory request config of risingwave named example-rw in namespace foo and in group test to 256Mi.
  kubectl rw update example-rw -n foo -g test -mr 256Mi
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

// request holds the request string and converted resource quantity.
type Request struct {
	requested_qty string
	converted_qty k8sresource.Quantity
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
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVar(&o.computeRequest.requested_qty, "cpurequest", "", "The target cpu request.")
	cmd.Flags().StringVar(&o.computeLimit.requested_qty, "cpulimit", "", "The target cpu limit.")
	cmd.Flags().StringVar(&o.memoryRequest.requested_qty, "memoryrequest", "", "The target memory request")
	cmd.Flags().StringVar(&o.memoryLimit.requested_qty, "memorylimit", "", "The target memory limit")
	cmd.Flags().StringVarP(&o.group, "group", "g", "default", "The group to be updated. If not set, update the default group")
	cmd.Flags().StringVarP(&o.component, "component", "c", "", "The component to be updated. If not set, return error")

	return cmd
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if o.computeRequest.requested_qty != "" {
		qty, err := k8sresource.ParseQuantity(o.computeRequest.requested_qty)
		if err != nil {
			return err
		}
		o.computeRequest.converted_qty = qty
	}

	if o.computeLimit.requested_qty != "" {
		qty, err := k8sresource.ParseQuantity(o.computeLimit.requested_qty)
		if err != nil {
			return err
		}
		o.computeLimit.converted_qty = qty
	}

	if o.memoryRequest.requested_qty != "" {
		qty, err := k8sresource.ParseQuantity(o.memoryRequest.requested_qty)
		if err != nil {
			return err
		}
		o.memoryRequest.converted_qty = qty
	}

	if o.memoryLimit.requested_qty != "" {
		qty, err := k8sresource.ParseQuantity(o.memoryLimit.requested_qty)
		if err != nil {
			return err
		}
		o.memoryLimit.converted_qty = qty
	}

	if o.group == "" {
		o.group = "default"
	} else {
		rw, err := o.GetRwInstance(ctx)
		if err != nil {
			return err
		}

		if !util.IsValidGroup(rw, o.group) {
			return fmt.Errorf("invalid group name %s", o.group)
		}
	}

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

type RWGroupInterface interface {
	RWGroup
}

type RWGroup struct {
	Name      string
	Resources corev1.ResourceRequirements
}

func (o *Options) updateComponentHelper(resourceMap *corev1.ResourceRequirements) {
	if o.computeRequest.requested_qty != "" {
		resourceMap.Requests[corev1.ResourceCPU] = o.computeRequest.converted_qty
	}

	if o.computeLimit.requested_qty != "" {
		resourceMap.Limits[corev1.ResourceCPU] = o.computeLimit.converted_qty
	}

	if o.memoryRequest.requested_qty != "" {
		resourceMap.Requests[corev1.ResourceMemory] = o.memoryRequest.converted_qty
	}

	if o.memoryLimit.requested_qty != "" {
		resourceMap.Limits[corev1.ResourceMemory] = o.memoryLimit.converted_qty
	}

}

func (o *Options) updateConfig(rw *v1alpha1.RisingWave) error {
	components := &rw.Spec.Components

	switch o.component {
	case "compute":
		for _, group := range components.Compute.Groups {
			if group.Name == o.group {
				o.updateComponentHelper(&group.Resources)
				break
			}
		}

	case "meta":
		for _, group := range components.Meta.Groups {
			if group.Name == o.group {
				o.updateComponentHelper(&group.Resources)
				break
			}
		}

	case "compactor":
		for _, group := range components.Compactor.Groups {
			if group.Name == o.group {
				o.updateComponentHelper(&group.Resources)
				break
			}
		}

	case "frontend":
		for _, group := range components.Frontend.Groups {
			if group.Name == o.group {
				o.updateComponentHelper(&group.Resources)
				break
			}
		}

	default:
		return fmt.Errorf("invalid component name %s", o.component)
	}

	return nil
}
