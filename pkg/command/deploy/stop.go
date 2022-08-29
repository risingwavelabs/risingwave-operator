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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	StopLongDesc = `
Stop the risingwave instances.
`
	StopExample = `  # Stop risingwave named example-rw in default namespace.
  kubectl rw stop example-rw

  # Stop risingwave named example-rw in foo namespace.
  kubectl rw stop example-rw -n foo
`
)

type StopOptions struct {
	*cmdcontext.BasicOptions
}

// NewCommand creates the stop command.
func NewStopCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := StopOptions{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Stop risingwave instances",
		Long:    StopLongDesc,
		Example: StopExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.ExitOnErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"pause"},
	}

	return cmd
}

func (o *StopOptions) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	if doesReplicaAnnotationExist(rw) {
		return fmt.Errorf("instance already stopped")
	}

	return nil
}

func (o *StopOptions) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	err = stopRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}

	return nil
}

func stopRisingWave(instance *v1alpha1.RisingWave) error {
	replicas := GroupReplicas{}

	// record current replica values in annotation
	for _, group := range instance.Spec.Components.Compactor.Groups {
		compactorReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		updateReplicas(instance, util.Compactor, group.Name, 0)
		replicas.Compactor = append(replicas.Compactor, compactorReplica)
	}

	for _, group := range instance.Spec.Components.Compute.Groups {
		computeReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		updateReplicas(instance, util.Compute, group.Name, 0)
		replicas.Compute = append(replicas.Compute, computeReplica)
	}

	for _, group := range instance.Spec.Components.Frontend.Groups {
		frontendReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		updateReplicas(instance, util.Frontend, group.Name, 0)
		replicas.Frontend = append(replicas.Frontend, frontendReplica)
	}

	for _, group := range instance.Spec.Components.Meta.Groups {
		metaReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		updateReplicas(instance, util.Meta, group.Name, 0)
		replicas.Meta = append(replicas.Meta, metaReplica)
	}

	global, err := json.Marshal(instance.Spec.Global.Replicas)
	if err != nil {
		return fmt.Errorf("failed to serialize replicas, %v", err)
	}
	stopGlobal(instance)

	// serialize replica struct to annotation
	annotation, err := json.Marshal(replicas)
	if err != nil {
		return fmt.Errorf("failed to serialize replicas, %v", err)
	}

	// set annotation
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations[ReplicaAnnotation] = string(annotation)
	instance.Annotations[GlobalReplicaAnnotation] = string(global)

	return nil
}

func stopGlobal(instance *v1alpha1.RisingWave) {
	instance.Spec.Global.Replicas = v1alpha1.RisingWaveGlobalReplicas{
		Meta:      0,
		Compactor: 0,
		Compute:   0,
		Frontend:  0,
	}
}
