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

package observability

import (
	"testing"

	"github.com/silence-operator/silence-operator/internal/tlsutil"
)

func TestMetricsOptionsInsecureByDefault(t *testing.T) {
	o := MetricsOptions(":8443", false, nil, nil)
	if o.BindAddress != ":8443" || o.SecureServing || o.FilterProvider != nil {
		t.Fatalf("unexpected options: %+v", o)
	}
}

func TestMetricsOptionsSecureSetsFilterProvider(t *testing.T) {
	o := MetricsOptions(":8443", true, nil, nil)
	if !o.SecureServing || o.FilterProvider == nil {
		t.Fatalf("expected secure serving with a filter provider, got %+v", o)
	}
}

func TestMetricsOptionsNilWatcherLeavesTLSOptsUntouched(t *testing.T) {
	tlsOpts := tlsutil.Options(false)

	o := MetricsOptions(":8443", false, tlsOpts, nil)
	if len(o.TLSOpts) != len(tlsOpts) {
		t.Fatalf("TLSOpts = %d entries, want %d", len(o.TLSOpts), len(tlsOpts))
	}
}
