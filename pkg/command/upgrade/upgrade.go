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

package upgrade

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/singularity-data/risingwave-operator/apis/risingwave/v1alpha1"
	cmdcontext "github.com/singularity-data/risingwave-operator/pkg/command/context"
	"github.com/singularity-data/risingwave-operator/pkg/command/util"
)

type Options struct {
	*cmdcontext.BasicOptions

	version string
}

const (
	LongDesc = `
Upgrade a risingwave instance to a specified version.
`
	Example = `  # Upgrade risingwave named example-rw to the latest version.
  kubectl rw upgrade example-rw

  # Upgrade risingwave named example-rw in namespace foo to the nightly version.
  kubectl rw upgrade example-rw -n foo -v nightly
`
)

func NewCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := Options{
		BasicOptions: cmdcontext.NewBasicOptions(streams),
	}

	cmd := &cobra.Command{
		Use:     "upgrade",
		Short:   "Upgrade the risingwave instance to a specified version",
		Long:    LongDesc,
		Example: Example,
		Run: func(cmd *cobra.Command, args []string) {
			util.CheckErr(o.Complete(ctx, cmd, args))
			util.ExitOnErr(o.Validate(ctx, cmd, args))
			util.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVarP(&o.version, "version", "v", "latest", "The version to upgrade to. If not specified, the latest version will be used.")

	return cmd
}

func (o *Options) Validate(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	err := o.verifyTag()
	if err != nil {
		return err
	}

	// parse current version
	rw, err := o.GetRwInstance(nil, ctx)
	if err != nil {
		return err
	}

	if rw.Spec.Global.Image == fmt.Sprintf("ghcr.io/singularity-data/risingwave:%s", o.version) {
		return fmt.Errorf("%s is already the current version", o.version)
	}

	return nil
}

func (o *Options) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	rw, err := o.GetRwInstance(context.Background(), ctx)
	if err != nil {
		return err
	}

	o.updateTag(rw)

	err = ctx.Client().Update(context.Background(), rw)
	if err != nil {
		return fmt.Errorf("failed to update instance, %v", err)
	}

	return nil
}

// compare it to the current version.
func (o *Options) updateTag(rw *v1alpha1.RisingWave) {
	image := fmt.Sprintf("ghcr.io/singularity-data/risingwave:%s", o.version)

	rw.Spec.Global.Image = image
	for i := range rw.Spec.Components.Compactor.Groups {
		rw.Spec.Components.Compactor.Groups[i].Image = image
	}
	for i := range rw.Spec.Components.Compute.Groups {
		rw.Spec.Components.Compute.Groups[i].Image = image
	}
	for i := range rw.Spec.Components.Frontend.Groups {
		rw.Spec.Components.Frontend.Groups[i].Image = image
	}
	for i := range rw.Spec.Components.Meta.Groups {
		rw.Spec.Components.Meta.Groups[i].Image = image
	}
}

// uses the github api to get the latest version.
// https://github.com/orgs/community/discussions/26279#discussioncomment-3251171
func (o *Options) verifyTag() error {
	// curl https://ghcr.io/token\?scope\="repository:singularity-data/risingwave:pull"
	token := "djE6c2luZ3VsYXJpdHktZGF0YS9yaXNpbmd3YXZlOjE2NjAyMDg1MTU1NDgyODMwMDU="

	// curl https://ghcr.io/v2/singularity-data/risingwave/tags/list -H "Authorization: Bearer"$token
	request, err := http.NewRequest("GET", "https://ghcr.io/v2/singularity-data/risingwave/tags/list", nil)
	if err != nil {
		return err
	}

	request.Header.Set("Authorization", "Bearer "+token)
	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	data := make(map[string]interface{})
	err = json.Unmarshal(body, &data)
	if err != nil {
		return err
	}

	for _, tag := range data["tags"].([]interface{}) {
		if tag == o.version {
			return nil
		}
	}

	return fmt.Errorf("version %s is not valid", o.version)
}
