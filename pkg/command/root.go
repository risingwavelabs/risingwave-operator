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

package command

import (
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/create"
	"github.com/singularity-data/risingwave-operator/pkg/command/delete"
	"github.com/singularity-data/risingwave-operator/pkg/command/deploy"
	"github.com/singularity-data/risingwave-operator/pkg/command/install"
	"github.com/singularity-data/risingwave-operator/pkg/command/list"
	"github.com/singularity-data/risingwave-operator/pkg/command/update"
	"github.com/singularity-data/risingwave-operator/pkg/command/version"
)

type RWOption struct {
	KubeConfigPath string
}

// NewCtlCommand creates the root `rw` command and its nested children.
func NewCtlCommand(streams genericclioptions.IOStreams) *cobra.Command {

	// Root command that all the subcommands are added to
	rootCmd := &cobra.Command{
		Use:   "kubectl rw",
		Short: "RisingWave management",
		Long:  RisingWaveCtlLongDescription,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Reuse kubectl global flags to provide namespace, context and credential options, include `namespace` option.
	kubeFlags := genericclioptions.NewConfigFlags(true)
	kubeFlags.AddFlags(rootCmd.PersistentFlags())

	ctx := context.NewContext(kubeFlags)

	groups := templates.CommandGroups{
		{
			Message: "Operator Management Commands:",
			Commands: []*cobra.Command{
				install.NewInstallCommand(ctx, streams),
				install.NewUninstallCommand(ctx, streams),
			},
		},
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				create.NewCommand(ctx, streams),
				delete.NewCommand(ctx, streams),
				list.NewCommand(ctx, streams),
			},
		},
		{
			Message: "Deploy Commands:",
			Commands: []*cobra.Command{
				deploy.NewScaleCommand(ctx, streams),
				deploy.NewStopCommand(ctx, streams),
				deploy.NewResumeCommand(ctx, streams),
				deploy.NewRestartCommand(ctx, streams),
			},
		},
		{
			Message: "Configuration Commands:",
			Commands: []*cobra.Command{
				update.NewCommand(ctx, streams),
			},
		},
	}
	groups.Add(rootCmd)

	// add `Other Commands`
	rootCmd.AddCommand(version.NewVersionCommand(ctx, streams))

	var skipCommandFilter = []string{"completion"}
	templates.ActsAsRootCommand(rootCmd, skipCommandFilter, groups...)

	return rootCmd
}
