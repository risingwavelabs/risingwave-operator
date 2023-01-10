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

package install

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	apiadmissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cmdcontext "github.com/risingwavelabs/risingwave-operator/pkg/command/context"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/install/version"
	"github.com/risingwavelabs/risingwave-operator/pkg/command/util/errors"
)

const (
	installExample = `  # install the latest risingwave operator into the cluster
  kubectl rw install

  # install the specified version risingwave operator into the cluster
  kubectl rw install --version v0.0.1
`
)

// InstallOptions contains the options to the installation command.
type InstallOptions struct {
	version string

	genericclioptions.IOStreams
}

// NewInstallOptions returns a InstallOptions.
func NewInstallOptions(streams genericclioptions.IOStreams) *InstallOptions {
	return &InstallOptions{
		version:   "latest",
		IOStreams: streams,
	}
}

// NewInstallCommand creates the installation command which can install the operator in the kubernetes cluster.
func NewInstallCommand(ctx *cmdcontext.RWContext, streams genericclioptions.IOStreams) *cobra.Command {
	o := NewInstallOptions(streams)

	cmd := &cobra.Command{
		Use:     "install",
		Short:   "Install the risingwave operator in the cluster",
		Long:    "Install the risingwave operator in the cluster",
		Example: installExample,
		Run: func(cmd *cobra.Command, args []string) {
			errors.CheckErr(o.Complete(ctx, cmd, args))
			errors.CheckErr(o.Run(ctx, cmd, args))
		},
	}

	cmd.Flags().StringVarP(&o.version, "version", "v", version.DefaultVersion, "the version of risingwave operator to install.")

	return cmd
}

func (o *InstallOptions) Complete(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {
	v, err := version.ValidateVersion(o.version)
	if !v {
		if err != nil {
			return err
		}
		return fmt.Errorf("invalid version %s", o.version)
	}

	return nil
}

// Run will run the command as followed:
// 1. check cert-manager
// 2. install cert-manager or give the installation guide
// 3. wait cert-manager ready
// 4. install risingwave operator.
func (o *InstallOptions) Run(ctx *cmdcontext.RWContext, cmd *cobra.Command, args []string) error {

	exist, err := hasOperator(ctx)
	if err != nil {
		return err
	}
	if exist {
		fmt.Fprintln(o.Out, "RisingWave Operator already exists")
		return nil
	}
	fmt.Fprintln(o.Out, "RisingWave Operator not exists, need to install it!")

	// check cert-manager
	exist, err = hasCertManagerCR(ctx)
	if err != nil {
		return err
	}
	if !exist {
		fmt.Fprintln(o.Out, "Install the cert-manager!")
		err := installCertManager(ctx)
		if err != nil {
			return fmt.Errorf("install cert-manager failed, %w", err)
		}
		fmt.Fprintln(o.Out, "Wait the cert-manager ready!")
		err = waitCertManager(ctx)
		if err != nil {
			return fmt.Errorf("wait cert-manager failed, %w", err)
		}
	}

	fmt.Fprintf(o.Out, "Install the %s risingwave-operator!", o.version)
	err = installOperator(ctx, o.version)
	if err != nil {
		return fmt.Errorf("install risingwave failed, %w", err)
	}

	fmt.Fprintln(o.Out, "RisingWave Operator has been installed")
	fmt.Fprintln(o.Out, "Wait the operator ready!")
	err = waitOperator(ctx)
	if err != nil {
		return fmt.Errorf("wait risingwave-operator failed, %w", err)
	}

	return nil
}

func waitOperator(ctx *cmdcontext.RWContext) error {
	defer func() {
		_ = ctx.Client().Delete(context.Background(), &rw)
	}()

	err := wait.PollImmediate(time.Second, time.Minute*TimeOut, func() (bool, error) {
		ready, inErr := checkOperatorReady(ctx)
		if inErr != nil {
			return false, inErr
		}
		return ready, inErr
	})
	if err != nil {
		return err
	}
	return nil
}

func checkOperatorReady(ctx *cmdcontext.RWContext) (bool, error) {
	// check service port.
	svc := corev1.Service{}
	err := ctx.Client().Get(context.Background(),
		client.ObjectKey{Namespace: "risingwave-operator-system", Name: "risingwave-operator-webhook-service"}, &svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if len(svc.Spec.Ports) == 0 {
		return false, nil
	}

	if svc.Spec.Ports[0].Port == 0 {
		return false, nil
	}

	// Check create test rw.
	// The risingwave instance only has a name and namespace.
	// So will return invalid error.
	// if Invalid error, means the webhook is ready.
	err = ctx.Client().Create(context.Background(), &rw)
	if err != nil && !apierrors.IsAlreadyExists(err) && !apierrors.IsInvalid(err) {
		return false, nil
	}

	return true, nil
}

func waitCertManager(ctx *cmdcontext.RWContext) error {
	defer GCTestCert(ctx)
	err := wait.PollImmediate(time.Second, time.Minute*TimeOut, func() (bool, error) {
		ready, inErr := checkCertManagerReady(ctx)
		if inErr != nil {
			return false, inErr
		}
		return ready, inErr
	})
	if err != nil {
		return err
	}
	return nil
}

func checkCertManagerReady(ctx *cmdcontext.RWContext) (bool, error) {
	// check CABundle.
	conf := &apiadmissionregistrationv1.ValidatingWebhookConfiguration{}
	err := ctx.Client().Get(context.Background(), client.ObjectKey{Name: "cert-manager-webhook"}, conf)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if len(conf.Webhooks) == 0 {
		return false, nil
	}

	if len(conf.Webhooks[0].ClientConfig.CABundle) == 0 {
		return false, nil
	}

	// check service port.
	svc := corev1.Service{}
	err = ctx.Client().Get(context.Background(),
		client.ObjectKey{Namespace: "cert-manager", Name: "cert-manager-webhook"}, &svc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	if len(svc.Spec.Ports) == 0 {
		return false, nil
	}

	if svc.Spec.Ports[0].Port == 0 {
		return false, nil
	}

	// check cert-manager test case.
	err = ctx.Client().Create(context.Background(), &issuer)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return false, nil
	}

	err = ctx.Client().Create(context.Background(), &cf)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return false, nil
	}

	return true, nil
}

func GCTestCert(ctx *cmdcontext.RWContext) {
	_ = ctx.Client().Delete(context.Background(), &issuer)
	_ = ctx.Client().Delete(context.Background(), &cf)
	var secret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecretName,
			Namespace: "cert-manager",
		},
	}
	_ = ctx.Client().Delete(context.Background(), &secret)
}

func hasOperator(ctx *cmdcontext.RWContext) (bool, error) {
	deploy := &v1.Deployment{}

	operatorKey := client.ObjectKey{
		Namespace: OperatorNamespace,
		Name:      OperatorName,
	}
	err := ctx.Client().Get(context.Background(), operatorKey, deploy)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func hasCertManagerCR(ctx *cmdcontext.RWContext) (bool, error) {
	list := &apiextensionsv1.CustomResourceDefinitionList{}

	err := ctx.Client().List(context.Background(), list)
	if err != nil && !apierrors.IsNotFound(err) {
		return false, err
	}

	for _, item := range list.Items {
		if item.Spec.Group == "cert-manager.io" {
			return true, nil
		}
	}
	return false, nil
}

// download the cert-manager.yaml
// apply into the cluster.
func installCertManager(ctx *cmdcontext.RWContext) error {
	certFile, err := Download(CertManagerURL, TemDir+"/cert-manager")
	if err != nil {
		return err
	}

	err = ctx.Applier().Apply(certFile)
	if err != nil {
		return err
	}

	return nil
}

// download the operator.yaml
// apply into the cluster
// TODO(xinyu): add the version for risingwave.yaml.
func installOperator(ctx *cmdcontext.RWContext, version string) error {
	url := fmt.Sprintf(RisingWaveURLTemplate, version)
	yamlFile, err := Download(url, TemDir+"/operator")
	if err != nil {
		return err
	}

	err = ctx.Applier().Apply(yamlFile)
	if err != nil {
		return err
	}

	return nil
}
