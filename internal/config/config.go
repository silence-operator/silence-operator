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

// Package config holds the operator's settings and their validation, source-agnostic (flags today).
package config

import (
	"errors"
	"fmt"
	"time"
)

// Defaults for every Config field that doesn't require an operator-supplied value.
const (
	DefaultInterval           = time.Minute * 5
	DefaultSilenceDuration    = time.Hour
	DefaultInstanceName       = "silence-operator"
	DefaultSilenceAuthor      = "silence-operator"
	DefaultConcurrency        = 10
	DefaultGetSilenceAttempts = 3
	DefaultGetSilenceInterval = time.Second * 10
)

// Config holds every setting the silence-operator binary needs to run.
type Config struct {
	MetricsAddr          string
	MetricsCertPath      string
	MetricsCertName      string
	MetricsCertKey       string
	WebhookCertPath      string
	WebhookCertName      string
	WebhookCertKey       string
	ProbeAddr            string
	InstanceName         string
	SilenceAuthor        string
	AlertManagerURL      string
	Interval             time.Duration
	SilenceDuration      time.Duration
	GetSilenceInterval   time.Duration
	GetSilenceAttempts   int
	Concurrency          int
	EnableLeaderElection bool
	SecureMetrics        bool
	EnableHTTP2          bool
}

// Validate rejects settings that would misbehave deep in the reconcile loop (e.g. GetSilenceAttempts=0).
func (c *Config) Validate() error {
	return errors.Join(
		atLeast("get-silence-attempts", c.GetSilenceAttempts, 1),
		positive("get-silence-interval", c.GetSilenceInterval),
		positive("interval", c.Interval),
		positive("silence-duration", c.SilenceDuration),
		atLeast("concurrency", c.Concurrency, 1),
	)
}

// atLeast rejects v below minimum, naming the setting it came from.
func atLeast(name string, v, minimum int) error {
	if v < minimum {
		return fmt.Errorf("%s must be >= %d, got %d", name, minimum, v)
	}

	return nil
}

// positive rejects a non-positive duration, naming the setting it came from.
func positive(name string, v time.Duration) error {
	if v <= 0 {
		return fmt.Errorf("%s must be > 0, got %s", name, v)
	}

	return nil
}
