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

package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	listLongDesc = `
Print the client version information for the current context.
`
	listExample = `  # Print the client versions for the current context
  kubectl rw version
`
)

// Options is a struct to support version command.
type Options struct {
	genericclioptions.IOStreams
}

// NewVersionCommand will run a cmd of version.
func NewVersionCommand(ctx *context.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := &Options{
		IOStreams: streams,
	}
	cmd := &cobra.Command{
		Use:     "version",
		Short:   "Print the client version",
		Long:    listLongDesc,
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"ls"},
	}

	return cmd
}

// TODO: add version file

func (o *Options) Run(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	fmt.Fprintf(o.Out, "Client Version: %s\n", "0.0.1")

	return nil
}
