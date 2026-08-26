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

	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/silence-operator/silence-operator/internal/config"
)

// parseFlags registers the operator's flags plus zap's on fs, parses args into a config.Config.
// zapOpts is bound to fs (not flag.CommandLine) so callers can pass an isolated FlagSet in tests.
func parseFlags(fs *flag.FlagSet, zapOpts *zap.Options, args []string) (*config.Config, error) {
	cfg := &config.Config{}

	fs.StringVar(&cfg.MetricsAddr, "metrics-bind-address", "0", "The address the metrics endpoint binds to. "+
		"Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	fs.StringVar(&cfg.ProbeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	fs.BoolVar(&cfg.EnableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	fs.BoolVar(&cfg.SecureMetrics, "metrics-secure", false,
		"If set, the metrics endpoint is served securely via HTTPS.")
	fs.StringVar(&cfg.WebhookCertPath, "webhook-cert-path", "", "The directory that contains the webhook certificate.")
	fs.StringVar(&cfg.WebhookCertName, "webhook-cert-name", "tls.crt", "The name of the webhook certificate file.")
	fs.StringVar(&cfg.WebhookCertKey, "webhook-cert-key", "tls.key", "The name of the webhook key file.")
	fs.StringVar(&cfg.MetricsCertPath, "metrics-cert-path", "",
		"The directory that contains the metrics server certificate.")
	fs.StringVar(&cfg.MetricsCertName, "metrics-cert-name", "tls.crt", "The name of the metrics server certificate file.")
	fs.StringVar(&cfg.MetricsCertKey, "metrics-cert-key", "tls.key", "The name of the metrics server key file.")
	fs.BoolVar(&cfg.EnableHTTP2, "enable-http2", false,
		"If set, HTTP/2 will be enabled for the metrics and webhook servers")
	fs.StringVar(&cfg.InstanceName, "instance-name", config.DefaultInstanceName, "Name of the silence operator instance.")
	fs.StringVar(&cfg.SilenceAuthor, "silence-author", config.DefaultSilenceAuthor,
		"This string will be used as 'Created by' field in AM silence.")
	fs.StringVar(&cfg.AlertManagerURL, "alertmanager-url", "", "AlertManager URL.")
	fs.DurationVar(&cfg.Interval, "interval", config.DefaultInterval, "The interval between reconciliations.")
	fs.DurationVar(&cfg.SilenceDuration, "silence-duration", config.DefaultSilenceDuration,
		"The duration for the silence.")
	fs.IntVar(&cfg.GetSilenceAttempts, "get-silence-attempts", config.DefaultGetSilenceAttempts,
		"Number of attempts to get the silence.")
	fs.DurationVar(&cfg.GetSilenceInterval, "get-silence-interval", config.DefaultGetSilenceInterval,
		"The interval between get silence attempts.")
	fs.IntVar(&cfg.Concurrency, "concurrency", config.DefaultConcurrency,
		"Amount of silences to be processed in parallel.")

	zapOpts.BindFlags(fs)

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
