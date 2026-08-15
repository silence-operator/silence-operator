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
	if err := run(os.Args[1:]); err != nil {
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

	if !opts.enableHTTP2 {
		setupLog.Info("disabling http/2")
	}
	tlsOpts := tlsutil.Options(opts.enableHTTP2)

	webhookCertWatcher, err := tlsutil.Watcher(opts.webhookCertPath, opts.webhookCertName, opts.webhookCertKey)
	if err != nil {
		return fmt.Errorf("unable to initialize webhook certificate watcher: %w", err)
	}
	if webhookCertWatcher != nil {
		setupLog.Info("Initializing webhook certificate watcher using provided certificates",
			"webhook-cert-path", opts.webhookCertPath, "webhook-cert-name", opts.webhookCertName,
			"webhook-cert-key", opts.webhookCertKey)
	}
	webhookServer := webhook.NewServer(webhook.Options{TLSOpts: tlsutil.WithCertificate(tlsOpts, webhookCertWatcher)})

	alertManagerClient, err := alertmanager.New(&alertmanager.Config{
		URL:             opts.alertManagerURL,
		Author:          opts.silenceAuthor,
		InstanceName:    opts.instanceName,
		SilenceDuration: opts.silenceDuration,
	})
	if err != nil {
		return fmt.Errorf("invalid alertmanager configuration: %w", err)
	}

	// Without a certificate, controller-runtime self-signs one for the metrics endpoint (fine for dev, not prod).
	metricsCertWatcher, err := tlsutil.Watcher(opts.metricsCertPath, opts.metricsCertName, opts.metricsCertKey)
	if err != nil {
		return fmt.Errorf("unable to initialize metrics certificate watcher: %w", err)
	}
	if metricsCertWatcher != nil {
		setupLog.Info("Initializing metrics certificate watcher using provided certificates",
			"metrics-cert-path", opts.metricsCertPath, "metrics-cert-name", opts.metricsCertName,
			"metrics-cert-key", opts.metricsCertKey)
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

	if err := (&controller.SilenceReconciler{
		Client:             mgr.GetClient(),
		Scheme:             mgr.GetScheme(),
		AlertManager:       alertManagerClient,
		Interval:           opts.interval,
		GetSilenceAttempts: opts.getSilenceAttempts,
		GetSilenceInterval: opts.getSilenceInterval,
	}).SetupWithManager(mgr); err != nil {
		return fmt.Errorf("unable to create Silence controller: %w", err)
	}
	// +kubebuilder:scaffold:builder

	if metricsCertWatcher != nil {
		setupLog.Info("Adding metrics certificate watcher to manager")
		if err := mgr.Add(metricsCertWatcher); err != nil {
			return fmt.Errorf("unable to add metrics certificate watcher to manager: %w", err)
		}
	}

	if webhookCertWatcher != nil {
		setupLog.Info("Adding webhook certificate watcher to manager")
		if err := mgr.Add(webhookCertWatcher); err != nil {
			return fmt.Errorf("unable to add webhook certificate watcher to manager: %w", err)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("unable to set up ready check: %w", err)
	}

	setupLog.Info("starting manager")
	return mgr.Start(ctrl.SetupSignalHandler())
}
