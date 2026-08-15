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

// Package tlsutil provides TLS config and certificate-watcher helpers shared by the metrics and
// webhook servers.
package tlsutil

import (
	"crypto/tls"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

// Options returns TLS config mutators with HTTP/2 disabled unless enableHTTP2 is set. Disabling
// HTTP/2 avoids the Stream Cancellation and Rapid Reset CVEs (GHSA-qppj-fm5r-hxr3, GHSA-4374-p667-p6c8).
func Options(enableHTTP2 bool) []func(*tls.Config) {
	if enableHTTP2 {
		return nil
	}
	return []func(*tls.Config){func(c *tls.Config) {
		c.NextProtos = []string{"http/1.1"}
	}}
}

// WithCertificate returns a copy of tlsOpts extended to serve certificates from watcher, or
// tlsOpts unchanged if watcher is nil.
func WithCertificate(tlsOpts []func(*tls.Config), watcher *certwatcher.CertWatcher) []func(*tls.Config) {
	if watcher == nil {
		return tlsOpts
	}
	out := make([]func(*tls.Config), len(tlsOpts), len(tlsOpts)+1)
	copy(out, tlsOpts)
	return append(out, func(c *tls.Config) {
		c.GetCertificate = watcher.GetCertificate
	})
}

// Watcher returns a CertWatcher for the cert/key pair named certName/keyName in path, or nil if
// path is empty, meaning no certificate was configured.
func Watcher(path, certName, keyName string) (*certwatcher.CertWatcher, error) {
	if path == "" {
		return nil, nil
	}
	return certwatcher.New(filepath.Join(path, certName), filepath.Join(path, keyName))
}
