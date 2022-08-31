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

package describe

import (
	context2 "context"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	cmdcontext "github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util/errors"
)

type Options struct {
	*cmdcontext.BasicOptions
	choice string
}

const (
	LongDesc = `
 Describe a risingwave instance.
 `
	Example = `  # Describe risingwave named example-rw.
  kubectl rw describe example-rw
 
  # Describe risingwave instance named example-rw in namespace foo.
  kubectl rw describe example-rw -n foo
 `
)

func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := Options{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "describe",
		Short:   "Describe a risingwave instance",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			errors.CheckErr(o.Complete(ctx, cmd, args))
			errors.CheckErr(o.Validate(ctx, cmd, args))
			errors.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVarP(&o.choice, "choice", "c", "spec", "The section of the risingwave instance you would like to describe. Spec, status and all are the only valid values.")

	return cmd
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if o.choice != "spec" && o.choice != "status" && o.choice != "all" {
		return fmt.Errorf("invalid section %s", o.choice)
	}

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRWInstance(context2.Background(), ctx)
	if err != nil {
		return err
	}

	switch o.choice {
	case "spec":
		o.describeMetadata(rw)
		o.describeSpec(rw)
	case "status":
		o.describeMetadata(rw)
		o.describeStatus(rw)
	case "all":
		o.describeMetadata(rw)
		o.describeRisingwaveVerbose(rw)
	}

	return nil
}
