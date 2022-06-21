/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	"go.uber.org/zap/zapcore"
	clouddns "google.golang.org/api/dns/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	capg "sigs.k8s.io/cluster-api-provider-gcp/api/v1beta1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/giantswarm/dns-operator-gcp/controllers"
	"github.com/giantswarm/dns-operator-gcp/pkg/k8sclient"
	"github.com/giantswarm/dns-operator-gcp/pkg/registrar"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(capg.AddToScheme(scheme))
	utilruntime.Must(capi.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	var baseDomain string
	var parentDNSZone string
	var gcpProject string
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	flag.StringVar(&gcpProject, "gcp-project", "",
		"The gcp project id where the dns records will be created.")
	flag.StringVar(&baseDomain, "base-domain", "",
		"The base domain to use when creating dns records.")
	flag.StringVar(&parentDNSZone, "parent-dns-zone", "",
		"The parent Cloud DNS zone, where the base domain is registered.")
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080",
		"The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")

	opts := zap.Options{
		Development: true,
		TimeEncoder: zapcore.RFC3339TimeEncoder,
	}

	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "c6d2deb7.giantswarm.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	service, err := clouddns.NewService(context.Background())
	if err != nil {
		setupLog.Error(err, "failed to create Cloud DNS client")
		os.Exit(1)
	}

	runtimeClient := mgr.GetClient()
	client := k8sclient.NewGCPCluster(runtimeClient)
	bastionsClient := k8sclient.NewBastions(runtimeClient, controllers.FinalizerDNS)
	loadBalancerClient := k8sclient.NewLoadBalancer(registrar.IngressNamespace, runtimeClient)
	zoneRegistrar := registrar.NewZone(baseDomain, parentDNSZone, gcpProject, service)
	apiRegistrar := registrar.NewAPI(baseDomain, service)
	bastionRegistrar := registrar.NewBastion(baseDomain, bastionsClient, service)
	ingressRegistrar := registrar.NewIngress(baseDomain, service, loadBalancerClient)
	wildcardRegistrar := registrar.NewWildcard(baseDomain, service)
	registrars := []controllers.Registrar{
		zoneRegistrar,
		apiRegistrar,
		bastionRegistrar,
		ingressRegistrar,
		wildcardRegistrar,
	}
	controller := controllers.NewGCPClusterReconciler(client, registrars)
	err = controller.SetupWithManager(mgr)
	if err != nil {
		setupLog.Error(err, "failed to setup controller", "controller", "GCPCluster")
		os.Exit(1)
	}

	//+kubebuilder:scaffold:builder

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
