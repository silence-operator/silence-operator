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

package config

import "testing"

func validConfig() *Config {
	return &Config{
		GetSilenceAttempts: DefaultGetSilenceAttempts,
		GetSilenceInterval: DefaultGetSilenceInterval,
		Interval:           DefaultInterval,
		SilenceDuration:    DefaultSilenceDuration,
		Concurrency:        DefaultConcurrency,
	}
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr bool
	}{
		{name: "valid defaults", mutate: func(*Config) {}},
		{name: "zero get-silence-attempts", mutate: func(c *Config) { c.GetSilenceAttempts = 0 }, wantErr: true},
		{name: "negative get-silence-attempts", mutate: func(c *Config) { c.GetSilenceAttempts = -1 }, wantErr: true},
		{name: "zero get-silence-interval", mutate: func(c *Config) { c.GetSilenceInterval = 0 }, wantErr: true},
		{name: "zero interval", mutate: func(c *Config) { c.Interval = 0 }, wantErr: true},
		{name: "zero silence-duration", mutate: func(c *Config) { c.SilenceDuration = 0 }, wantErr: true},
		{name: "zero concurrency", mutate: func(c *Config) { c.Concurrency = 0 }, wantErr: true},
		{
			name: "multiple invalid fields",
			mutate: func(c *Config) {
				c.GetSilenceAttempts = 0
				c.GetSilenceInterval = 0
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := validConfig()
			tt.mutate(c)

			err := c.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
