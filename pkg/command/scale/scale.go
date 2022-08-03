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

package scale

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Scale the risingwave instances.
`
	Example = `  # Scale compute-node of the risingwave named example-rw to 2.
  kubectl rw scale example-rw -t 2

  # Scale frontend of the risingwave named example-rw to 2 in the foo namespace.
  kubectl rw scale example-rw -n foo -c frontend -t 2

  # Scale frontend of the risingwave which named example-rw to 2 and in the foo namespace and in the test group.
  kubectl rw scale example-rw -n foo -c frontend -t 2 -g test
`
)

type Options struct {
	name string

	namespace string

	target int

	component string

	group string

	genericclioptions.IOStreams
}

// NewOptions returns a scale Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the scale command which can scale the risingwave components.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "scale",
		Short:   "Scale risingwave instances",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"sc"},
	}

	cmd.Flags().StringVarP(&o.component, "component", "c", "compute-node", "The component which you want to scale.")
	cmd.Flags().IntVarP(&o.target, "target", "t", -1, "The target number.")
	cmd.Flags().StringVarP(&o.group, "group", "g", "default", "The group name of the component. If not set, scale the default group")

	return cmd
}

func (o *Options) Complete(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if len(ctx.Namespace()) == 0 {
		o.namespace = "default"
	} else {
		o.namespace = ctx.Namespace()
	}

	if len(args) == 0 {
		return fmt.Errorf("name of risingwave cannot be nil")
	} else {
		o.name = args[0]
	}
	return nil
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if o.target < 0 {
		fmt.Fprint(o.Out, "No specific target or target is negative, will do nothing")
		return nil
	}

	rw := &v1alpha1.RisingWave{}

	operatorKey := client.ObjectKey{
		Name:      o.name,
		Namespace: o.namespace,
	}

	err := ctx.Client().Get(context.Background(), operatorKey, rw)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Fprint(o.Out, "Risingwave instance not exists")
			return nil
		}
		return err
	}

	err = UpdateReplicas(rw, o.component, o.group, int32(o.target))
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %v", err)
	}

	return nil
}

func UpdateReplicas(instance *v1alpha1.RisingWave, component, groupName string, target int32) error {
	switch component {
	case "compute":
		for i, group := range instance.Spec.Components.Compute.Groups {
			if group.Name == groupName {
				instance.Spec.Components.Compute.Groups[i].Replicas = target
				break
			}
		}
	case "frontend":
		for i, group := range instance.Spec.Components.Frontend.Groups {
			if group.Name == groupName {
				instance.Spec.Components.Frontend.Groups[i].Replicas = target
				break
			}
		}
	case "compactor":
		for i, group := range instance.Spec.Components.Compactor.Groups {
			if group.Name == groupName {
				instance.Spec.Components.Compactor.Groups[i].Replicas = target
				break
			}
		}
	case "meta":
		for i, group := range instance.Spec.Components.Meta.Groups {
			if group.Name == groupName {
				instance.Spec.Components.Meta.Groups[i].Replicas = target
				break
			}
		}
	default:
		return fmt.Errorf("no such component %v, will do nothing", component)
	}

	return nil
}
