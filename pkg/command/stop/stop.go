package stop

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
	"github.com/singularity-data/risingwave-operator/pkg/command/scale"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	LongDesc = `
Stop the risingwave instances.
`
	Example = `  # Stop risingwave named example-rw in default namespace.
  kubectl rw stop example-rw

  # Stop risingwave named example-rw in foo namespace.
  kubectl rw stop example-rw -n foo
`
)

type Options struct {
	name string

	namespace string

	genericclioptions.IOStreams
}

// NewOptions returns a stop Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the stop command.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "stop",
		Short:   "Stop risingwave instances",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"stp"},
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

	err = StopRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %w", err)
	}

	fmt.Fprint(o.Out, "Risingwave instance updated")
	return nil
}

// TODO: move to common package
type GroupReplicas struct {
	Compute   []ReplicaInfo
	Frontend  []ReplicaInfo
	Compactor []ReplicaInfo
	Meta      []ReplicaInfo
}

type ReplicaInfo struct {
	GroupName string
	Replicas  int32
}

func StopRisingWave(instance *v1alpha1.RisingWave) error {
	replicas := GroupReplicas{}

	// record current replica values in annotation
	for _, group := range instance.Spec.Components.Compute.Groups {
		computeReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		// TODO: use constants for component naming
		scale.UpdateReplicas(instance, "compute", group.Name, 0)
		replicas.Compute = append(replicas.Compute, computeReplica)
	}

	for _, group := range instance.Spec.Components.Frontend.Groups {
		frontendReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		scale.UpdateReplicas(instance, "frontend", group.Name, 0)
		replicas.Frontend = append(replicas.Frontend, frontendReplica)
	}

	for _, group := range instance.Spec.Components.Compactor.Groups {
		compactorReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		scale.UpdateReplicas(instance, "compactor", group.Name, 0)
		replicas.Compactor = append(replicas.Compactor, compactorReplica)
	}

	for _, group := range instance.Spec.Components.Meta.Groups {
		metaReplica := ReplicaInfo{
			GroupName: group.Name,
			Replicas:  group.Replicas,
		}
		scale.UpdateReplicas(instance, "meta", group.Name, 0)
		replicas.Meta = append(replicas.Meta, metaReplica)
	}

	// serialise replica struct to annotation
	annotation, err := json.Marshal(replicas)
	if err != nil {
		return fmt.Errorf("failed to serialise replicas, %v", err)
	}

	// set annotation
	// TODO: create map in create command
	if instance.Annotations == nil {
		instance.Annotations = make(map[string]string)
	}
	instance.Annotations["replicas.old"] = string(annotation)

	return nil
}
