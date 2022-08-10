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
	RestartLongDesc = `
Restart the risingwave instances.
`
	RestartExample = `  # Restart risingwave named example-rw in default namespace.
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
		Long:    RestartLongDesc,
		Example: RestartExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	return cmd
}

func (o *RestartOptions) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return err
	}

	if !doesReplicaAnnotationExist(rw) {
		return fmt.Errorf("instance already stopped")
	}

	return nil
}

func (o *RestartOptions) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return err
	}

	err = stopRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to stop instance, %w", err)
	}

	// check that all replicas have scaled down
	err = o.verifyStopped(ctx)
	if err != nil {
		return err
	}

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

func (o *RestartOptions) verifyStopped(ctx *cmdcontext.RWContext) error {
	timeout := time.After(10 * time.Minute)
	for {
		select {
		case <-timeout:
			return fmt.Errorf("failed to stop instance, timed out")
		default:
			count, err := o.getRunningCount(ctx)
			if err != nil {
				return err
			}
			if count == 0 {
				return nil
			}
			time.Sleep(5 * time.Second)
		}
	}
}

func (o *RestartOptions) getRunningCount(ctx *cmdcontext.RWContext) (int32, error) {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return -1, err
	}

	status := rw.Status.ComponentReplicas
	count := status.Compute.Running + status.Frontend.Running + status.Compactor.Running + status.Meta.Running

	return count, nil
}
