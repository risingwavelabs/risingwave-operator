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

package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	prometheusv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	kruiseappsv1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseappsv1beta1 "github.com/openkruise/kruise-api/apps/v1beta1"

	risingwavev1alpha1 "github.com/risingwavelabs/risingwave-operator/apis/risingwave/v1alpha1"
	risingwavecontroller "github.com/risingwavelabs/risingwave-operator/pkg/controller"
	"github.com/risingwavelabs/risingwave-operator/pkg/features"
	"github.com/risingwavelabs/risingwave-operator/pkg/metrics"
	risingwavewebhook "github.com/risingwavelabs/risingwave-operator/pkg/webhook"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(risingwavev1alpha1.AddToScheme(scheme))
	utilruntime.Must(apiextensionsv1.AddToScheme(scheme))
	utilruntime.Must(prometheusv1.AddToScheme(scheme))
	utilruntime.Must(kruiseappsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(kruiseappsv1beta1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

var (
	metricsAddr          string
	probeAddr            string
	configPath           string
	enableLeaderElection bool
	featureGates         string
	operatorVersion      string
)

func requireKubernetesVersion(serverVersion *version.Info, minMajor, minMinor int) {
	major, _ := strconv.Atoi(serverVersion.Major)
	minor, _ := strconv.Atoi(strings.TrimRight(serverVersion.Minor, "+"))
	if major < minMajor || (major == minMajor && minor < minMinor) {
		setupLog.Error(nil, "Kubernetes version is too low", "expected", fmt.Sprintf("%d.%d+", minMajor, minMinor),
			"actual", serverVersion.GitVersion)
		os.Exit(1)
	}
}

func main() {
	metrics.InitMetrics()
	metrics.ReceivingMetricsFromOperator.Inc()
	flag.StringVar(&configPath, "config-file", "/config/config.yaml", "The file path of the configuration file.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&featureGates, "feature-gates", "", "The feature gates arguments for the operator.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	featureManager := features.InitFeatureManager(features.SupportedFeatureList, featureGates)

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	config := ctrl.GetConfigOrDie()

	mgr, err := ctrl.NewManager(config, ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		WebhookServer:          webhook.NewServer(webhook.Options{}),
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "02bd7444.risingwavelabs.com",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Barrier to ensure that the operator is not started on Kubernetes lower than 1.21.
	// This is to avoid the issue that the operator will be stuck in a crash loop if it is started on Kubernetes lower than 1.21.
	kubernetesVersion, err := discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig()).ServerVersion()
	if err != nil {
		setupLog.Error(err, "Unable to get Kubernetes version")
		os.Exit(1)
	}
	requireKubernetesVersion(kubernetesVersion, 1, 21)

	if err = risingwavewebhook.SetupWebhooksWithManager(mgr, featureManager.IsFeatureEnabled(features.EnableOpenKruiseFeature)); err != nil {
		setupLog.Error(err, "unable to setup webhooks")
		os.Exit(1)
	}

	if err = risingwavecontroller.NewMetaPodRoleLabeler(mgr.GetClient()).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "meta-pod-role-labeler")
		os.Exit(1)
	}

	if err = risingwavecontroller.NewRisingWaveController(
		mgr.GetClient(),
		mgr.GetEventRecorderFor("risingwave-controller"),
		featureManager.IsFeatureEnabled(features.EnableOpenKruiseFeature),
		featureManager.IsFeatureEnabled(features.EnableForceUpdate),
		operatorVersion,
	).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RisingWave")
		os.Exit(1)
	}

	if err = risingwavecontroller.NewRisingWaveScaleViewController(mgr.GetClient()).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RisingWaveScaleView")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
