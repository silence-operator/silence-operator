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

package tlsutil

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
)

func TestOptions(t *testing.T) {
	tests := []struct {
		name           string
		enableHTTP2    bool
		wantNextProtos []string
	}{
		{name: "disabled by default", enableHTTP2: false, wantNextProtos: []string{"http/1.1"}},
		{name: "left enabled", enableHTTP2: true, wantNextProtos: nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &tls.Config{}
			for _, opt := range Options(tt.enableHTTP2) {
				opt(cfg)
			}

			if !slices.Equal(cfg.NextProtos, tt.wantNextProtos) {
				t.Fatalf("NextProtos = %v, want %v", cfg.NextProtos, tt.wantNextProtos)
			}
		})
	}
}

func TestWatcher(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T) string
		wantNil bool
		wantErr bool
	}{
		{name: "empty path", path: func(*testing.T) string { return "" }, wantNil: true},
		{name: "missing files", path: func(t *testing.T) string { return t.TempDir() }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, err := Watcher(tt.path(t), "tls.crt", "tls.key")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantNil && w != nil {
				t.Fatalf("Watcher() = %v, want nil", w)
			}
		})
	}
}

func TestWithCertificate(t *testing.T) {
	tests := []struct {
		name        string
		watcher     func(t *testing.T) *certwatcher.CertWatcher
		wantLenDiff int
		wantCert    bool
	}{
		{name: "nil watcher leaves input unchanged", watcher: func(*testing.T) *certwatcher.CertWatcher { return nil }},
		{name: "watcher appends GetCertificate", watcher: newTestWatcher, wantLenDiff: 1, wantCert: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := Options(false)

			got := WithCertificate(base, tt.watcher(t))
			if len(got) != len(base)+tt.wantLenDiff {
				t.Fatalf("len(got) = %d, want %d", len(got), len(base)+tt.wantLenDiff)
			}

			if len(base) != 1 {
				t.Fatalf("base was mutated: len = %d, want 1", len(base))
			}

			if tt.wantCert {
				cfg := &tls.Config{}
				got[len(got)-1](cfg)

				if cfg.GetCertificate == nil {
					t.Fatal("GetCertificate was not set")
				}
			}
		})
	}
}

// newTestWatcher writes a throwaway self-signed cert/key pair and returns a Watcher for it.
func newTestWatcher(t *testing.T) *certwatcher.CertWatcher {
	t.Helper()
	dir := t.TempDir()
	writeSelfSignedCert(t, dir, "tls.crt", "tls.key")

	w, err := Watcher(dir, "tls.crt", "tls.key")
	if err != nil {
		t.Fatalf("Watcher() error = %v", err)
	}

	return w
}

func writeSelfSignedCert(t *testing.T, dir, certFile, keyFile string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "tlsutil-test"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("CreateCertificate() error = %v", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	err = os.WriteFile(filepath.Join(dir, certFile), certPEM, 0o600)
	if err != nil {
		t.Fatalf("write cert: %v", err)
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey() error = %v", err)
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	err = os.WriteFile(filepath.Join(dir, keyFile), keyPEM, 0o600)
	if err != nil {
		t.Fatalf("write key: %v", err)
	}
}
