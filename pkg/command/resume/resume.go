package resume

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/stop"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Start the risingwave instances.
`
	Example = `  # Resume risingwave named example-rw in default namespace.
  kubectl rw resume example-rw

  # Resume risingwave named example-rw in foo namespace.
  kubectl rw resume example-rw -n foo
`
)

type Options struct {
	name string

	namespace string

	genericclioptions.IOStreams
}

// NewOptions returns a resume Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the resume command.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "resume",
		Short:   "Resume risingwave instances",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"start"},
	}

	return cmd
}

func (o *Options) Complete(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	if len(ctx.Namespace()) == 0 {
		o.namespace = "default"
	} else {
		o.namespace = ctx.Namespace()
	}

	if len(args) == 0 {
		return fmt.Errorf("name of risingwave cannot be nil")
	} else {
		o.name = args[0]
	}
	return nil
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw := &v1alpha1.RisingWave{}

	operatorKey := client.ObjectKey{
		Name:      o.name,
		Namespace: o.namespace,
	}

	err := ctx.Client().Get(context.Background(), operatorKey, rw)
	if err != nil {
		if errors.IsNotFound(err) {
			fmt.Fprint(o.Out, "Risingwave instance not exists")
			return nil
		}
		return err
	}

	err = ResumeRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}

	return nil
}

func ResumeRisingWave(instance *v1alpha1.RisingWave) error {
	// deserialize the annotation
	// TODO: move this to utils
	replicas := stop.GroupReplicas{}

	if instance.Annotations == nil {
		return fmt.Errorf("error replica information. are you trying to resume an instance that was not stopped?")
	}

	err := json.Unmarshal([]byte(instance.Annotations["replicas.old"]), &replicas)
	if err != nil {
		return fmt.Errorf("failed to unmarshal replicas, %v; are you trying to resume an instance that was not stopped?", err)
	}

	for _, replicaInfo := range replicas.Compute {
		for i, group := range instance.Spec.Components.Compute.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Compute.Groups[i].Replicas = replicaInfo.Replicas
				break
			}
		}
	}

	for _, replicaInfo := range replicas.Compactor {
		for i, group := range instance.Spec.Components.Compactor.Groups {
			if group.Name == replicaInfo.GroupName {
				instance.Spec.Components.Compactor.Groups[i].Replicas = replicaInfo.Replicas
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
	delete(instance.Annotations, "replicas.old")

	return nil
}
