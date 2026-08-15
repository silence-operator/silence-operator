// Copyright 2026 Silence-Operator Maintainers
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

const (
	defaultInterval           = time.Minute * 5
	defaultDuration           = time.Hour
	defaultInstanceName       = "silence-operator"
	defaultSilenceAuthor      = "silence-operator"
	defaultConcurrency        = 10
	defaultGetSilenceAttempts = 3
	defaultGetSilenceInterval = time.Second * 10
)

// operatorOptions holds every CLI-configurable setting for the silence-operator binary.
type operatorOptions struct {
	metricsAddr          string
	metricsCertPath      string
	metricsCertName      string
	metricsCertKey       string
	webhookCertPath      string
	webhookCertName      string
	webhookCertKey       string
	probeAddr            string
	instanceName         string
	silenceAuthor        string
	alertManagerURL      string
	interval             time.Duration
	silenceDuration      time.Duration
	getSilenceInterval   time.Duration
	getSilenceAttempts   int
	concurrency          int
	enableLeaderElection bool
	secureMetrics        bool
	enableHTTP2          bool
}

// parseFlags registers the operator's flags plus zap's on fs, parses args, and returns the result.
// zapOpts is bound to fs (not flag.CommandLine) so callers can pass an isolated FlagSet in tests.
func parseFlags(fs *flag.FlagSet, zapOpts *zap.Options, args []string) (*operatorOptions, error) {
	o := &operatorOptions{}

	fs.StringVar(&o.metricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	fs.StringVar(&o.probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&o.enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.BoolVar(&o.secureMetrics, "metrics-secure", false,
		"If set, the metrics endpoint is served securely via HTTPS.")
	fs.StringVar(&o.webhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	fs.StringVar(&o.webhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	fs.StringVar(&o.webhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	fs.StringVar(&o.metricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	fs.StringVar(&o.metricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	fs.StringVar(&o.metricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	fs.BoolVar(&o.enableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	fs.StringVar(&o.instanceName, "instance-name", defaultInstanceName, "Name of the silence operator instance.")
	fs.StringVar(&o.silenceAuthor, "silence-author", defaultSilenceAuthor,
		"This string will be used as 'Created by' field in AM silence.")
	fs.StringVar(&o.alertManagerURL, "alertmanager-url", "", "AlertManager URL.")
	fs.DurationVar(&o.interval, "interval", defaultInterval, "The interval between reconciliations.")
	fs.DurationVar(&o.silenceDuration, "silence-duration", defaultDuration,
		"The duration for the silence.")
	fs.IntVar(&o.getSilenceAttempts, "get-silence-attempts", defaultGetSilenceAttempts,
		"Number of attempts to get the silence.")
	fs.DurationVar(&o.getSilenceInterval, "get-silence-interval", defaultGetSilenceInterval,
		"The interval between get silence attempts.")
	fs.IntVar(&o.concurrency, "concurrency", defaultConcurrency,
		"Amount of silences to be processed in parallel.")

	zapOpts.BindFlags(fs)

	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return o, nil
}
