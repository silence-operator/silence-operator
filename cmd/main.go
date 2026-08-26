/*
Copyright 2026 Silence-Operator Maintainers.

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
	"flag"
	"fmt"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/config"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	monitoringv1alpha1 "github.com/silence-operator/silence-operator/api/v1alpha1"
	"github.com/silence-operator/silence-operator/internal/alertmanager"
	"github.com/silence-operator/silence-operator/internal/controller"
	o11y "github.com/silence-operator/silence-operator/internal/observability"
	"github.com/silence-operator/silence-operator/internal/tlsutil"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(monitoringv1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	err := run(os.Args[1:])
	if err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// run parses flags, wires up the manager and controller, and blocks until the manager stops.
func run(args []string) error {
	zapOpts := zap.Options{Development: false}

	opts, err := parseFlags(flag.CommandLine, &zapOpts, args)
	if err != nil {
		return err
	}

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zapOpts)))

	// Pure validation first: fail before touching any OS resource (cert watchers open fsnotify watches).
	alertManagerClient, err := alertmanager.New(&alertmanager.Config{
		URL:             opts.alertManagerURL,
		Author:          opts.silenceAuthor,
		InstanceName:    opts.instanceName,
		SilenceDuration: opts.silenceDuration,
	})
	if err != nil {
		return fmt.Errorf("invalid alertmanager configuration: %w", err)
	}

	if !opts.enableHTTP2 {
		setupLog.Info("disabling http/2")
	}

	tlsOpts := tlsutil.Options(opts.enableHTTP2)

	webhookCertWatcher, err := setupCertWatcher("webhook", opts.webhookCertPath, opts.webhookCertName, opts.webhookCertKey)
	if err != nil {
		return err
	}

	webhookServer := webhook.NewServer(webhook.Options{TLSOpts: tlsutil.WithCertificate(tlsOpts, webhookCertWatcher)})

	// Without a certificate, controller-runtime self-signs one for the metrics endpoint (fine for dev, not prod).
	metricsCertWatcher, err := setupCertWatcher("metrics", opts.metricsCertPath, opts.metricsCertName, opts.metricsCertKey)
	if err != nil {
		return err
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                o11y.MetricsOptions(opts.metricsAddr, opts.secureMetrics, tlsOpts, metricsCertWatcher),
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: opts.probeAddr,
		LeaderElection:         opts.enableLeaderElection,
		LeaderElectionID:       "silence-operator-leader-election",
		// LeaderElectionReleaseOnCancel is left off: it's only safe once shutdown never lingers.
		Controller: config.Controller{
			MaxConcurrentReconciles: opts.concurrency,
			RecoverPanic:            ptr.To(true),
		},
	})
	if err != nil {
		return fmt.Errorf("unable to start manager: %w", err)
	}

	err = (&controller.SilenceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		AlertManager:       alertManagerClient,
		Interval:           opts.interval,
		GetSilenceAttempts: opts.getSilenceAttempts,
		GetSilenceInterval: opts.getSilenceInterval,
	}).SetupWithManager(mgr)
	if err != nil {
		return fmt.Errorf("unable to create Silence controller: %w", err)
	}
	// +kubebuilder:scaffold:builder

	err = addCertWatcher(mgr, "metrics", metricsCertWatcher)
	if err != nil {
		return err
	}

	err = addCertWatcher(mgr, "webhook", webhookCertWatcher)
	if err != nil {
		return err
	}

	err = mgr.AddHealthzCheck("healthz", healthz.Ping)
	if err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}

	err = mgr.AddReadyzCheck("readyz", healthz.Ping)
	if err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	setupLog.Info("starting manager")

	return mgr.Start(ctrl.SetupSignalHandler())
}
