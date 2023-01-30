/*
 * Copyright 2023 RisingWave Labs
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

package create

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/utils/pointer"

	"github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/create/config"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util/errors"
)

const (
	LongDesc = `
Create a risingwave instance.
`
	Example = `  # Create a risingwave named example-rw in the test namespace by default configuration.
  kubectl rw create example-rw -n test

  # Create a risingwave named example-rw by configuration file.
  kubectl rw create -c rw.config
`
)

type Options struct {
	name string

	namespace string

	configFile string

	config config.Config

	genericclioptions.IOStreams
}

// NewOptions returns a create Options.
func NewOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the create command which can create a risingwave instance.
func NewCommand(ctx cmdcontext.Context, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewOptions(streams)

	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Create a risingwave instance",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			errors.CheckErr(o.Complete(ctx, cmd, args))
			errors.CheckErr(o.Validate(ctx, cmd, args))
			errors.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"new"},
	}

	cmd.Flags().StringVarP(&o.configFile, "config", "c", o.configFile, "The config file used when creating the instance.")

	return cmd
}

func (o *Options) Complete(ctx cmdcontext.Context, cmd *cobra.Command, args []string) error {
	if len(ctx.Namespace()) == 0 {
		o.namespace = "default"
	} else {
		o.namespace = ctx.Namespace()
	}

	if len(args) != 0 {
		o.name = args[0]
	}

	c, err := config.ApplyConfigFile(o.configFile)
	if err != nil {
		return err
	}
	o.config = c
	return nil
}

func (o *Options) Validate(ctx cmdcontext.Context, cmd *cobra.Command, args []string) error {
	if len(o.name) == 0 && len(o.configFile) == 0 {
		return fmt.Errorf("name should be set when using default config")
	}
	return nil
}

func (o *Options) Run(ctx cmdcontext.Context, cmd *cobra.Command, args []string) error {
	rw, err := o.createInstance()
	if err != nil {
		return err
	}

	err = ctx.Client().Create(context.Background(), rw)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to create instance, %w", err)
	}

	s := fmt.Sprintf("Succeed to create risingwave %s in %s namespace.\n", o.name, o.namespace)
	fmt.Fprint(o.Out, s)
	return nil
}

// TODO: to support create different risingwave by config file
// TODO: to support different storage. issue: https://github.com/risingwavelabs/risingwave-operator/issues/182
func (o *Options) createInstance() (*v1alpha1.RisingWave, error) {
	c := o.config
	rw := &v1alpha1.RisingWave{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   o.namespace,
			Name:        o.name,
			Annotations: make(map[string]string),
		},
		Spec: v1alpha1.RisingWaveSpec{
			Global: v1alpha1.RisingWaveGlobalSpec{
				Replicas: v1alpha1.RisingWaveGlobalReplicas{
					Meta:      c.BaseConfig.Replicas,
					Frontend:  c.BaseConfig.Replicas,
					Compute:   c.BaseConfig.Replicas,
					Compactor: c.BaseConfig.Replicas,
				},
				RisingWaveComponentGroupTemplate: v1alpha1.RisingWaveComponentGroupTemplate{
					Image:     c.BaseConfig.Image,
					Resources: *c.BaseConfig.Resources.DeepCopy(),
				},
			},
			Storages: v1alpha1.RisingWaveStoragesSpec{
				Meta: v1alpha1.RisingWaveMetaStorage{
					Memory: pointer.Bool(true),
				},
				Object: v1alpha1.RisingWaveObjectStorage{
					Memory: pointer.Bool(true),
				},
			},
			Components: v1alpha1.RisingWaveComponentsSpec{
				Meta: v1alpha1.RisingWaveComponentMeta{
					Groups: o.createComponentGroups(c.MetaConfig),
				},
				Frontend: v1alpha1.RisingWaveComponentFrontend{
					Groups: o.createComponentGroups(c.FrontendConfig),
				},

				Compute: v1alpha1.RisingWaveComponentCompute{
					Groups: o.createComputeGroups(c.ComputeConfig),
				},
				Compactor: v1alpha1.RisingWaveComponentCompactor{
					Groups: o.createComponentGroups(c.CompactorConfig),
				},
			},
		},
	}

	return rw, nil
}

func (o *Options) createComponentGroups(c config.ComponentConfig) []v1alpha1.RisingWaveComponentGroup {
	var groups []v1alpha1.RisingWaveComponentGroup
	for _, g := range c.Groups {
		groups = append(groups, v1alpha1.RisingWaveComponentGroup{
			Name:     g.Name,
			Replicas: g.Replicas,
			RisingWaveComponentGroupTemplate: &v1alpha1.RisingWaveComponentGroupTemplate{
				Image:     o.config.Image,
				Resources: *g.Resources.DeepCopy(),
			},
		})
	}

	return groups
}

func (o *Options) createComputeGroups(c config.ComponentConfig) []v1alpha1.RisingWaveComputeGroup {
	var groups []v1alpha1.RisingWaveComputeGroup
	for _, g := range c.Groups {
		groups = append(groups, v1alpha1.RisingWaveComputeGroup{
			Name:     g.Name,
			Replicas: g.Replicas,
			RisingWaveComputeGroupTemplate: &v1alpha1.RisingWaveComputeGroupTemplate{
				RisingWaveComponentGroupTemplate: v1alpha1.RisingWaveComponentGroupTemplate{
					Image:     o.config.Image,
					Resources: *g.Resources.DeepCopy(),
				},
			},
		})
	}

	return groups
}
