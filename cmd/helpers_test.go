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
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

const (
	webhookComponent = "webhook"
	metricsComponent = "metrics"
)

func TestSetupCertWatcher(t *testing.T) {
	tests := []struct {
		name         string
		component    string
		path         func(t *testing.T) string
		wantNil      bool
		wantErrMatch string
	}{
		{
			name:      "empty path returns nil watcher and no error",
			component: webhookComponent,
			path:      func(*testing.T) string { return "" },
			wantNil:   true,
		},
		{
			name:         "missing files returns a wrapped error naming the component",
			component:    metricsComponent,
			path:         func(t *testing.T) string { return t.TempDir() },
			wantErrMatch: "metrics certificate watcher",
		},
		{
			name:      "valid cert pair returns a non-nil watcher",
			component: webhookComponent,
			path: func(t *testing.T) string {
				dir := t.TempDir()
				writeSelfSignedCert(t, dir, "tls.crt", "tls.key")
				return dir
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := setupCertWatcher(tt.component, tt.path(t), "tls.crt", "tls.key")

			if tt.wantErrMatch != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErrMatch) {
					t.Fatalf("error = %v, want it to mention %q", err, tt.wantErrMatch)
				}
				return
			}
			if err != nil {
				t.Fatalf("setupCertWatcher() error = %v", err)
			}
			if got := w != nil; got == tt.wantNil {
				t.Fatalf("setupCertWatcher() watcher = %v, want nil = %v", w, tt.wantNil)
			}
		})
	}
}

func TestAddCertWatcher(t *testing.T) {
	tests := []struct {
		name         string
		component    string
		watcher      func(t *testing.T) *certwatcher.CertWatcher
		addErr       error
		wantAddCalls int
		wantErrMatch string
	}{
		{
			name:      "nil watcher is a no-op",
			component: webhookComponent,
			watcher:   func(*testing.T) *certwatcher.CertWatcher { return nil },
		},
		{
			name:         "non-nil watcher is registered on the adder",
			component:    webhookComponent,
			watcher:      newTestCertWatcher,
			wantAddCalls: 1,
		},
		{
			name:         "Add error is wrapped with the component name",
			component:    metricsComponent,
			watcher:      newTestCertWatcher,
			addErr:       errors.New("boom"),
			wantAddCalls: 1,
			wantErrMatch: "metrics certificate watcher",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adder := &fakeRunnableAdder{addErr: tt.addErr}

			err := addCertWatcher(adder, tt.component, tt.watcher(t))
			switch {
			case tt.wantErrMatch != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErrMatch)):
				t.Fatalf("error = %v, want it to mention %q", err, tt.wantErrMatch)
			case tt.wantErrMatch == "" && err != nil:
				t.Fatalf("addCertWatcher() error = %v", err)
			}

			if adder.addCalls != tt.wantAddCalls {
				t.Fatalf("Add() called %d times, want %d", adder.addCalls, tt.wantAddCalls)
			}
		})
	}
}

// newTestCertWatcher returns a watcher over a throwaway self-signed cert/key pair.
func newTestCertWatcher(t *testing.T) *certwatcher.CertWatcher {
	t.Helper()

	dir := t.TempDir()
	writeSelfSignedCert(t, dir, "tls.crt", "tls.key")

	w, err := setupCertWatcher(webhookComponent, dir, "tls.crt", "tls.key")
	if err != nil {
		t.Fatalf("setupCertWatcher() error = %v", err)
	}
	return w
}

// fakeRunnableAdder is a runnableAdder stub that records how many times Add was called.
type fakeRunnableAdder struct {
	addCalls int
	addErr   error
}

func (f *fakeRunnableAdder) Add(manager.Runnable) error {
	f.addCalls++
	return f.addErr
}

// writeSelfSignedCert writes a throwaway self-signed cert/key pair into dir for use with certwatcher.
func writeSelfSignedCert(t *testing.T, dir, certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "cmd-test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	if err := os.WriteFile(filepath.Join(dir, certFile), certPEM, 0o600); err != nil {
		t.Fatalf("write cert: %v", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey() error = %v", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	if err := os.WriteFile(filepath.Join(dir, keyFile), keyPEM, 0o600); err != nil {
		t.Fatalf("write key: %v", err)
	}
}
