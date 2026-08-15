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
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestParseFlagsDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	opts, err := parseFlags(fs, &zapOpts, nil)
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	want := &operatorOptions{
		metricsAddr:        "0",
		probeAddr:          ":8081",
		webhookCertName:    "tls.crt",
		webhookCertKey:     "tls.key",
		metricsCertName:    "tls.crt",
		metricsCertKey:     "tls.key",
		instanceName:       defaultInstanceName,
		silenceAuthor:      defaultSilenceAuthor,
		interval:           defaultInterval,
		silenceDuration:    defaultDuration,
		getSilenceAttempts: defaultGetSilenceAttempts,
		getSilenceInterval: defaultGetSilenceInterval,
		concurrency:        defaultConcurrency,
	}
	if *opts != *want {
		t.Fatalf("parseFlags() = %+v, want %+v", *opts, want)
	}
}

func TestParseFlagsOverrides(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	opts, err := parseFlags(fs, &zapOpts, []string{
		"-alertmanager-url", "http://alertmanager:9093",
		"-concurrency", "42",
		"-interval", "30s",
		"-enable-http2",
		"-metrics-secure",
	})
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	switch {
	case opts.alertManagerURL != "http://alertmanager:9093":
		t.Errorf("alertManagerURL = %q", opts.alertManagerURL)
	case opts.concurrency != 42:
		t.Errorf("concurrency = %d", opts.concurrency)
	case opts.interval != 30*time.Second:
		t.Errorf("interval = %v", opts.interval)
	case !opts.enableHTTP2:
		t.Error("enableHTTP2 = false, want true")
	case !opts.secureMetrics:
		t.Error("secureMetrics = false, want true")
	}
}

func TestParseFlagsBindsZapFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	if _, err := parseFlags(fs, &zapOpts, []string{"-zap-devel"}); err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}
	if !zapOpts.Development {
		t.Error("Development = false, want true after -zap-devel")
	}
}

func TestParseFlagsInvalid(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	if _, err := parseFlags(fs, &zapOpts, []string{"-concurrency", "not-a-number"}); err == nil {
		t.Fatal("expected an error for an invalid -concurrency value")
	}
}
