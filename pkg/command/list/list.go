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

package list

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	"github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

const (
	listLongDesc = `
List all risingwave instances.

By specifying namespace or label-selectors, you can filter instances.
`
	listExample = `  # list all clusters and sync to local config
  kubectl rw list

  # filter by namespace
  kubectl rw list --namespace=foo

  # get risingwave instances by selector
  kubectl rw list -l foo=bar
`
)

// Options contains the input to the list command.
type Options struct {
	allNamespaces bool

	namespace string

	selector string

	// TODO: we can print the rw by some print flags, just like "-w, -o yaml"
	//printFlags *genericclioptions.PrintFlags

	//printer printers.ResourcePrinter

	genericclioptions.IOStreams
}

// NewListOptions returns a ListOptions.
func NewListOptions(streams genericclioptions.IOStreams) *Options {
	return &Options{
		IOStreams: streams,
	}
}

// NewCommand creates the list command which lists all the risingwave instances
// in the specified kubernetes cluster and sync to local config file.
func NewCommand(ctx *context.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewListOptions(streams)

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List risingwave instances",
		Long:    listLongDesc,
		Example: listExample,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.CheckErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
		Aliases: []string{"ps"},
	}

	//o.printFlags = genericclioptions.NewPrintFlags("list").WithTypeSetter(scheme)
	//o.printFlags.AddFlags(cmd)

	cmd.Flags().StringVarP(&o.selector, "selector", "l", o.selector, "Selector (label query) to filter on, not including uninitialized ones, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2).")
	cmd.Flags().BoolVarP(&o.allNamespaces, "all-namespaces", "A", false, "Whether list instances in all namespaces.")

	return cmd
}

func (o *Options) Complete(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	namespace := ctx.Namespace()
	o.namespace = namespace

	//printer, err := o.printFlags.ToPrinter()
	//if err != nil {
	//	return err
	//}
	//o.printer = printer

	return nil
}

func (o *Options) Validate(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	if len(o.namespace) == 0 {
		o.namespace = "default"
	}
	return nil
}

func (o *Options) Run(ctx *context.RWContext, cmd *cobra.Command, args []string) error {
	r := ctx.
		Builder().
		LabelSelectorParam(o.selector).
		NamespaceParam(o.namespace).DefaultNamespace().AllNamespaces(o.allNamespaces).
		SingleResourceType().ResourceTypes("risingwaves").
		SelectAllParam(true).
		Unstructured().
		ContinueOnError().
		Latest().
		Flatten().
		Do()

	if err := r.Err(); err != nil {
		return err
	}

	infos, err := r.Infos()
	if err != nil {
		return err
	}

	var rwList []*v1alpha1.RisingWave
	for _, info := range infos {
		internalObj, _ := ctx.Scheme().ConvertToVersion(info.Object, v1alpha1.GroupVersion)
		rw := internalObj.(*v1alpha1.RisingWave)
		rwList = append(rwList, rw)
	}

	if len(rwList) == 0 {
		s := fmt.Sprintf("No resources found in %s namespace.\n", util.Bold(o.namespace))
		fmt.Fprint(o.Out, s)
		return nil
	}

	// TODO: sort rwList

	printTable(rwList, o.Out)

	return nil
}

// TODO(xinyu): readable print as table.
func printTable(rwList []*v1alpha1.RisingWave, w io.Writer) {
	for _, rw := range rwList {
		fmt.Fprintln(w, rw.Namespace+"/"+rw.Name+" "+rw.CreationTimestamp.String())
	}
}
