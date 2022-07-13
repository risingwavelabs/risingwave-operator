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

package delete

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Delete the risingwave instance.
`
	Example = `  # Delete the risingwave named example-rw in the test namespace.
  kubectl rw delete example-rw -n test

  # Force delete the risingwave named example-rw in the test namespace.
  kubectl rw delete example-rw -n test -f
`
)

type Options struct {
	name string

	namespace string

	// TODO: maybe can support force delete
	force bool

	genericclioptions.IOStreams
}

// NewOptions returns a delete Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the delete command which can create a risingwave instance.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "delete",
		Short:   "Delete a risingwave instance",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"delete"},
	}

	cmd.Flags().BoolVarP(&o.force, "force", "f", false, "force delete the tenant")

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

	return nil
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {

	rw := &v1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: o.namespace,
			Name:      o.name,
		},
	}

	err := ctx.Client().Delete(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}
	return nil
}
