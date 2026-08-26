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

	"github.com/silence-operator/silence-operator/internal/config"
)

func TestParseFlagsDefaults(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	cfg, err := parseFlags(fs, &zapOpts, nil)
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	want := &config.Config{
		MetricsAddr:        "0",
		ProbeAddr:          ":8081",
		WebhookCertName:    "tls.crt",
		WebhookCertKey:     "tls.key",
		MetricsCertName:    "tls.crt",
		MetricsCertKey:     "tls.key",
		InstanceName:       config.DefaultInstanceName,
		SilenceAuthor:      config.DefaultSilenceAuthor,
		Interval:           config.DefaultInterval,
		SilenceDuration:    config.DefaultSilenceDuration,
		GetSilenceAttempts: config.DefaultGetSilenceAttempts,
		GetSilenceInterval: config.DefaultGetSilenceInterval,
		Concurrency:        config.DefaultConcurrency,
	}
	if *cfg != *want {
		t.Fatalf("parseFlags() = %+v, want %+v", *cfg, want)
	}
}

func TestParseFlagsOverrides(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	cfg, err := parseFlags(fs, &zapOpts, []string{
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
	case cfg.AlertManagerURL != "http://alertmanager:9093":
		t.Errorf("AlertManagerURL = %q", cfg.AlertManagerURL)
	case cfg.Concurrency != 42:
		t.Errorf("Concurrency = %d", cfg.Concurrency)
	case cfg.Interval != 30*time.Second:
		t.Errorf("Interval = %v", cfg.Interval)
	case !cfg.EnableHTTP2:
		t.Error("EnableHTTP2 = false, want true")
	case !cfg.SecureMetrics:
		t.Error("SecureMetrics = false, want true")
	}
}

func TestParseFlagsBindsZapFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	_, err := parseFlags(fs, &zapOpts, []string{"-zap-devel"})
	if err != nil {
		t.Fatalf("parseFlags() error = %v", err)
	}

	if !zapOpts.Development {
		t.Error("Development = false, want true after -zap-devel")
	}
}

func TestParseFlagsInvalid(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	zapOpts := zap.Options{}

	_, err := parseFlags(fs, &zapOpts, []string{"-concurrency", "not-a-number"})
	if err == nil {
		t.Fatal("expected an error for an invalid -concurrency value")
	}
}
