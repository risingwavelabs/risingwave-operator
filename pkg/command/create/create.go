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
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/pkg/command/util"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/create/config"
)

const (
	LongDesc = `
Create a risingwave instance.
`
	Example = `  # Create a risingwave named example-rw in the test namespace by default configuration.
  kubectl rw create example-rw -n test

  # Create a risingwave named example-rw by configuration file.
  kubectl rw create -c rw.config
`
)

type Options struct {
	name string

	namespace string

	configFile string

	config config.Config

	genericclioptions.IOStreams
}

// NewOptions returns a create Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the create command which can create a risingwave instance.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a risingwave instance",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"new"},
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "The config file used when creating the instance.")

	return cmd
}

func (o *Options) Complete(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if len(ctx.Namespace()) == 0 {
		o.namespace = "default"
	} else {
		o.namespace = ctx.Namespace()
	}

	if len(args) != 0 {
		o.name = args[0]
	}

	if len(o.configFile) == 0 {
		o.config = config.DefaultConfig
	} else {
		c, err := config.ApplyConfigFile(o.configFile)
		if err != nil {
			return err
		}
		o.config = c
	}
	return nil
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if len(o.name) == 0 && len(o.configFile) == 0 {
		return fmt.Errorf("name should be set when using defalut config")
	}
	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.createInstance()
	if err != nil {
		return err
	}

	err = ctx.Client().Create(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}
	return nil
}

// TODO: will create new risingwave instance when PR(https://github.com/singularity-data/risingwave-operator/pull/105) merged.
func (o *Options) createInstance() (*v1alpha1.RisingWave, error) {
	rw := &v1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.namespace,
			Name:      o.name,
		},
	}
	return rw, nil
}
