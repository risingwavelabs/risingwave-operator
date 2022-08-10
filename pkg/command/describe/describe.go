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

package describe

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

type Options struct {
	*cmdcontext.BasicOptions
}

const (
	LongDesc = `
 Describe a risingwave instance.
 `
	Example = `  # Describe risingwave named example-rw.
   kubectl rw describe example-rw
 
   # Describe risingwave instance named example-rw in namespace foo.
   kubectl rw describe example-rw -n foo
 `
)

func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := Options{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "describe",
		Short:   "Describe a risingwave instance",
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

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(ctx)
	if err != nil {
		return err
	}

	return o.printRisingwaveInfo(rw)
}

func (o *Options) printRisingwaveInfo(rw *v1alpha1.RisingWave) error {
	fmt.Fprintf(o.Out, "Name: %s\n", rw.Name)
	fmt.Fprintf(o.Out, "Namespace: %s\n", rw.Namespace)
	fmt.Fprintf(o.Out, "Labels: %v\n", rw.Labels)
	fmt.Fprintf(o.Out, "Annotations: %v\n", rw.Annotations)
	fmt.Fprintf(o.Out, "API Version: %s\n", rw.APIVersion)
	fmt.Fprintf(o.Out, "Kind: %s\n", rw.Kind)
	fmt.Fprintf(o.Out, "Metadata:\n")
	fmt.Fprintf(o.Out, "  Creation Timestamp: %s\n", rw.CreationTimestamp.String())
	fmt.Fprintf(o.Out, "  Generation: %d\n", rw.Generation)
	fmt.Fprintf(o.Out, "  Resource Version: %s\n", rw.ResourceVersion)
	fmt.Fprintf(o.Out, "  UID: %s\n", rw.UID)

	// spec
	o.describeSpec(rw)

	// configuration
	fmt.Fprintf(o.Out, "Configuration:\n")
	if rw.Spec.Configuration.ConfigMap != nil {
		fmt.Fprintf(o.Out, "  ConfigMap: %s\n", rw.Spec.Configuration.ConfigMap.String())
	}

	// security
	fmt.Fprintf(o.Out, "  Security:\n")
	if rw.Spec.Security.TLS != nil {
		fmt.Fprintf(o.Out, "    Enabled: %t\n", rw.Spec.Security.TLS.Enabled)
		fmt.Fprintf(o.Out, "    Secret: %s\n", rw.Spec.Security.TLS.Secret)
	}

	// storage
	fmt.Fprintf(o.Out, "  Storages:\n")
	fmt.Fprintf(o.Out, "    Meta:\n")

	if !*rw.Spec.Storages.Meta.Memory {
		fmt.Fprintf(o.Out, "      ETCD Endpoint: %s\n", rw.Spec.Storages.Meta.Etcd.Endpoint)
		fmt.Fprintf(o.Out, "      ETCD Secret: %s\n", rw.Spec.Storages.Meta.Etcd.Secret)
	} else {
		fmt.Fprintf(o.Out, "      Memory:%v\n", *rw.Spec.Storages.Meta.Memory)
	}

	fmt.Fprintf(o.Out, "    Object:\n")
	if *rw.Spec.Storages.Object.Memory {
		fmt.Fprintf(o.Out, "      Memory:%v\n", *rw.Spec.Storages.Object.Memory)
	} else if rw.Spec.Storages.Object.S3 != nil {
		fmt.Fprintf(o.Out, "      S3 Bucket: %s\n", rw.Spec.Storages.Object.S3.Bucket)
		fmt.Fprintf(o.Out, "      S3 Secret: %s\n", rw.Spec.Storages.Object.S3.Secret)
	} else {
		fmt.Fprintf(o.Out, "      MinIO Endpoint: %s\n", rw.Spec.Storages.Object.MinIO.Endpoint)
		fmt.Fprintf(o.Out, "      MinIO Bucket: %s\n", rw.Spec.Storages.Object.MinIO.Bucket)
		fmt.Fprintf(o.Out, "      MinIO Secret: %s\n", rw.Spec.Storages.Object.MinIO.Secret)
	}

	// status
	o.describeStatus(rw)
	return nil
}

func (o *Options) describeSpec(rw *v1alpha1.RisingWave) {
	fmt.Fprintf(o.Out, "Spec:\n")
	fmt.Fprintf(o.Out, "  Compactor:\n")
	for _, group := range rw.Spec.Components.Compactor.Groups {
		fmt.Fprintf(o.Out, "    Group: %s\n", group.Name)
		o.describeGroupSpec(*group.RisingWaveComponentGroupTemplate)
	}
	o.describePorts(rw.Spec.Components.Compactor.Ports)

	fmt.Fprintf(o.Out, "  Compute:\n")
	for _, group := range rw.Spec.Components.Compute.Groups {
		fmt.Fprintf(o.Out, "    Group: %s\n", group.Name)
		o.describeGroupSpec(group.RisingWaveComponentGroupTemplate)
	}
	o.describePorts(rw.Spec.Components.Compute.Ports)

	fmt.Fprintf(o.Out, "  Frontend:\n")
	for _, group := range rw.Spec.Components.Frontend.Groups {
		fmt.Fprintf(o.Out, "    Group: %s\n", group.Name)
		o.describeGroupSpec(*group.RisingWaveComponentGroupTemplate)
	}
	o.describePorts(rw.Spec.Components.Frontend.Ports)

	fmt.Fprintf(o.Out, "  Meta:\n")
	for _, group := range rw.Spec.Components.Meta.Groups {
		fmt.Fprintf(o.Out, "    Group: %s\n", group.Name)
		o.describeGroupSpec(*group.RisingWaveComponentGroupTemplate)
	}
	o.describePorts(rw.Spec.Components.Meta.Ports.RisingWaveComponentCommonPorts)

	fmt.Fprintf(o.Out, "  Global:\n")
	o.describeGroupSpec(rw.Spec.Global.RisingWaveComponentGroupTemplate)

}

func (o *Options) describeGroupSpec(component v1alpha1.RisingWaveComponentGroupTemplate) {
	// cpu in millicores and memory in bibytes.
	cpuRequest := component.Resources.Requests.Cpu().MilliValue()
	memRequest := formatMem(component.Resources.Requests.Memory())
	cpuLimit := component.Resources.Limits.Cpu().MilliValue()
	memLimit := formatMem(component.Resources.Limits.Memory())

	fmt.Fprintf(o.Out, "      Image:              %s\n", component.Image)
	fmt.Fprintf(o.Out, "      Image Pull Policy:  %s\n", component.ImagePullPolicy)
	fmt.Fprintf(o.Out, "      Resources:\n")
	fmt.Fprintf(o.Out, "        Limits:\n")
	fmt.Fprintf(o.Out, "          Cpu:     %dm\n", cpuRequest)
	fmt.Fprintf(o.Out, "          Memory:  %s\n", memRequest)
	fmt.Fprintf(o.Out, "        Requests:\n")
	fmt.Fprintf(o.Out, "          Cpu:     %dm\n", cpuLimit)
	fmt.Fprintf(o.Out, "          Memory:  %s\n", memLimit)
	fmt.Fprintf(o.Out, "      Upgrade Strategy:\n")
	fmt.Fprintf(o.Out, "        Type:  %s\n", component.UpgradeStrategy.Type)
}

func (o *Options) describePorts(ports v1alpha1.RisingWaveComponentCommonPorts) {
	fmt.Fprintf(o.Out, "    Ports:\n")
	fmt.Fprintf(o.Out, "      Metrics: %d\n", ports.MetricsPort)
	fmt.Fprintf(o.Out, "      Service: %d\n", ports.ServicePort)
}

// convert into bibytes and add suffix.
func formatMem(mem *resource.Quantity) string {
	suffix := []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}
	value := float64(mem.Value())
	for _, s := range suffix {
		if value < 1024 {
			return fmt.Sprintf("%d%s", int(value), s)
		}
		value /= 1024
	}
	return fmt.Sprintf("%d%s", int(value), suffix[len(suffix)-1])
}

func (o *Options) describeStatus(rw *v1alpha1.RisingWave) {
	fmt.Fprintf(o.Out, "Status:\n")
	fmt.Fprintf(o.Out, "  Component Replicas:\n")

	fmt.Fprintf(o.Out, "    Compactor:\n")
	o.describeComponentReplicas(rw.Status.ComponentReplicas.Compactor)
	fmt.Fprintf(o.Out, "    Compute:\n")
	o.describeComponentReplicas(rw.Status.ComponentReplicas.Compute)
	fmt.Fprintf(o.Out, "    Frontend:\n")
	o.describeComponentReplicas(rw.Status.ComponentReplicas.Frontend)
	fmt.Fprintf(o.Out, "    Meta:\n")
	o.describeComponentReplicas(rw.Status.ComponentReplicas.Meta)

	fmt.Fprintf(o.Out, "  Conditions:\n")
	for _, condition := range rw.Status.Conditions {
		fmt.Fprintf(o.Out, "    %s:\n", condition.Type)
		fmt.Fprintf(o.Out, "      Last Transition Time:  %s\n", condition.LastTransitionTime)
		fmt.Fprintf(o.Out, "      Status:                %s\n", condition.Status)
		fmt.Fprintf(o.Out, "      Type:                  %s\n", condition.Type)
	}
	fmt.Fprintf(o.Out, "  Observed Generation:       %d\n", rw.Status.ObservedGeneration)
	fmt.Fprintf(o.Out, "  Storages:\n")
	fmt.Fprintf(o.Out, "    Meta:\n")
	fmt.Fprintf(o.Out, "      Type:  %s\n", rw.Status.Storages.Meta.Type)
	fmt.Fprintf(o.Out, "    Object:\n")
	fmt.Fprintf(o.Out, "      Type:  %s\n", rw.Status.Storages.Object.Type)
}

func (o *Options) describeComponentReplicas(component v1alpha1.ComponentReplicasStatus) {
	fmt.Fprintf(o.Out, "      Groups:\n")
	for _, group := range component.Groups {
		fmt.Fprintf(o.Out, "        %s (%d/%d)\n", group.Name, group.Running, group.Target)
	}
	fmt.Fprintf(o.Out, "      Total (%d/%d)\n", component.Running, component.Target)
}
