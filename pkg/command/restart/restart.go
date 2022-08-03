package restart

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/resume"
	"github.com/singularity-data/risingwave-operator/pkg/command/stop"
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

type Options struct {
	name string

	namespace string

	genericclioptions.IOStreams
}

// TODO: create a generic option creater
// NewOptions returns a restart Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the restart command.
func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

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

	err = stop.StopRisingWave(rw)
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
	_ = ctx.Client().Get(context.Background(), operatorKey, rw)

	err = resume.ResumeRisingWave(rw)
	if err != nil {
		return err
	}

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to resume instance, %w", err)
	}

	return nil
}
