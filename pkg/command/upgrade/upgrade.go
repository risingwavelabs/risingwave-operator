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

package upgrade

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

type Options struct {
	*cmdcontext.BasicOptions

	version string
}

const (
	LongDesc = `
Upgrade a risingwave instance to the latest version.
`
	Example = `  # Upgrade risingwave named example-rw to the latest version.
  kubectl rw upgrade example-rw

  # Upgrade risingwave named example-rw in namespace foo to the latest version.
  kubectl rw upgrade example-rw -n foo
`
)

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

	return cmd
}

// check if the version is valid.
func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	return nil
}
