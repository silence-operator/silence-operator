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

// Package observability builds the metrics server options for the operator's manager.
package observability

import (
	"crypto/tls"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/silence-operator/silence-operator/internal/tlsutil"
)

// MetricsOptions builds controller-runtime metrics server options for addr. secure enables an
// authn/authz filter provider; a non-nil watcher wires its certificate into tlsOpts.
func MetricsOptions(addr string, secure bool, tlsOpts []func(*tls.Config), watcher *certwatcher.CertWatcher) metricsserver.Options {
	o := metricsserver.Options{
		BindAddress:   addr,
		SecureServing: secure,
		TLSOpts:       tlsutil.WithCertificate(tlsOpts, watcher),
	}
	if secure {
		o.FilterProvider = filters.WithAuthenticationAndAuthorization
	}
	return o
}
