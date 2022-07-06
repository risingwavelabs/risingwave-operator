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

package install

import (
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	installExample = `  # install the latest risingwave operator into the cluster
  kubectl rw install

  # install the specified version risingwave operator into the cluster
  kubectl rw install --version v0.0.1
`
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme))
}

// InstallOptions contains the options to the install command.
type InstallOptions struct {
	version string

	genericclioptions.IOStreams
}

// NewInstallOptions returns a InstallOptions.
func NewInstallOptions(streams genericclioptions.IOStreams) *InstallOptions {
	return &InstallOptions{
		version:   "latest",
		IOStreams: streams,
	}
}

// NewInstallCommand creates the installation command which can install the operator in the kubernetes cluster.
func NewInstallCommand(ctx *context.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewInstallOptions(streams)

	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Install the risingwave operator in the cluster",
		Long:    "Install the risingwave operator in the cluster",
		Example: installExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVarP(&o.version, "version", "v", o.version, "the version of risingwave operator to install.")

	return cmd
}

func (o *InstallOptions) Complete(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	if len(o.version) == 0 {
		o.version = "latest"
	}

	return nil
}

// 1. check cert-manager
// 2. install cert-manager or give the installation guide
// 3. wait cert-manager ready
// 4. install risingwave operator

func (o *InstallOptions) Run(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	return nil
}
