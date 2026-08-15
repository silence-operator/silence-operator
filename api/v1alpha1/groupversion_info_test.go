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

package v1alpha1

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
)

func TestAddToScheme(t *testing.T) {
	scheme := runtime.NewScheme()

	if err := AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() returned error: %v", err)
	}

	for _, kind := range []string{"Silence", "SilenceList"} {
		gvk := GroupVersion.WithKind(kind)
		if !scheme.Recognizes(gvk) {
			t.Errorf("scheme does not recognize %s after AddToScheme()", gvk)
		}
	}
}

func TestGroupVersion(t *testing.T) {
	if GroupVersion.Group != "monitoring.coreos.com" {
		t.Errorf("GroupVersion.Group = %q, want %q", GroupVersion.Group, "monitoring.coreos.com")
	}
	if GroupVersion.Version != "v1alpha1" {
		t.Errorf("GroupVersion.Version = %q, want %q", GroupVersion.Version, "v1alpha1")
	}
}
