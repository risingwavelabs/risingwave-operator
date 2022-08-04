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

package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Restart the risingwave instances.
`
	Example = `  # Restart risingwave named example-rw in default namespace.
  kubectl rw restart example-rw

  # Restart risingwave named example-rw in foo namespace.
  kubectl rw restart example-rw -n foo
`
)

type RestartOptions struct {
	*cmdcontext.BasicOptions
}

// NewCommand creates the restart command.
func NewRestartCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := RestartOptions{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "restart",
		Short:   "Restart risingwave instances",
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

func (o *RestartOptions) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return err
	}

	err = StopRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to stop instance, %w", err)
	}

	// check that all replicas have scaled down
	// TODO: use client go to achieve this
	time.Sleep(time.Minute * 3)

	// update rw
	rw, _ = o.GetRwInstance(ctx)

	err = resumeRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to resume instance, %w", err)
	}

	return nil
}
