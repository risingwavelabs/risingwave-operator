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
	ResumeLongDesc = `
Start the risingwave instances.
`
	ResumeExample = `  # Resume risingwave named example-rw in default namespace.
  kubectl rw resume example-rw

  # Resume risingwave named example-rw in foo namespace.
  kubectl rw resume example-rw -n foo
`
)

type ResumeOptions struct {
	*cmdcontext.BasicOptions
}

// NewCommand creates the resume command.
func NewResumeCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := &ResumeOptions{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "resume",
		Short:   "Resume risingwave instances",
		Long:    ResumeLongDesc,
		Example: ResumeExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"start"},
	}

	return cmd
}

func (o *ResumeOptions) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	err = resumeRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}

	return nil
}

func resumeRisingWave(instance *v1alpha1.RisingWave) error {
	// deserialize the annotation
	replicas := GroupReplicas{}

	if instance.Annotations == nil {
		return fmt.Errorf("error retrieving replica information; are you trying to resume an instance that was not stopped?")
	}

	err := json.Unmarshal([]byte(instance.Annotations[ReplicaAnnotation]), &replicas)
	if err != nil {
		return fmt.Errorf("failed to unmarshal replicas, %v; are you trying to resume an instance that was not stopped?", err)
	}

	global := v1alpha1.RisingWaveGlobalReplicas{}
	err = json.Unmarshal([]byte(instance.Annotations[GlobalReplicaAnnotation]), &global)
	if err != nil {
		return fmt.Errorf("failed to unmarshal global replicas, %v; are you trying to resume an instance that was not stopped?", err)
	}

	instance.Spec.Global.Replicas = global

	for _, replicaInfo := range replicas.Compactor {
		for i, group := range instance.Spec.Components.Compactor.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Compactor.Groups[i].Replicas = replicaInfo.Replicas
				break
			}
		}
	}

	for _, replicaInfo := range replicas.Compute {
		for i, group := range instance.Spec.Components.Compute.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Compute.Groups[i].Replicas = replicaInfo.Replicas
				break
			}
		}
	}

	for _, replicaInfo := range replicas.Frontend {
		for i, group := range instance.Spec.Components.Frontend.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Frontend.Groups[i].Replicas = replicaInfo.Replicas
				break
			}
		}
	}

	for _, replicaInfo := range replicas.Meta {
		for i, group := range instance.Spec.Components.Meta.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Meta.Groups[i].Replicas = replicaInfo.Replicas
				break
			}
		}
	}

	// delete annotation
	delete(instance.Annotations, ReplicaAnnotation)

	return nil
}
