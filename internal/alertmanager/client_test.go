/*
Copyright 2026 Silence-Operator Maintainers.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package alertmanager

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/silence-operator/silence-operator/api/v1alpha1"
)

const (
	silencesPath      = "/api/v2/silences"
	existingSilenceID = "existing-id"
)

func newTestSilence(amID string) *v1alpha1.Silence {
	s := &v1alpha1.Silence{
		Spec: v1alpha1.SilenceSpec{
			Comment: "test silence",
			Matchers: v1alpha1.Matchers{
				{Name: "alertname", Value: "TestAlert", IsEqual: true, IsRegex: false},
			},
		},
	}
	s.Status.AlertManagerID = amID

	return s
}

func newTestAlertManager(t *testing.T, url string) *AlertManager {
	t.Helper()

	am, err := New(&Config{
		URL:             url,
		Author:          "test-author",
		InstanceName:    "test-instance",
		SilenceDuration: time.Hour,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	return am
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "url without scheme defaults to http", url: "alertmanager.default:9093"},
		{name: "url with explicit scheme", url: "https://alertmanager.default:9093"},
		{name: "invalid url", url: "http://[::1]:namedport", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&Config{URL: tt.url})

			if (err != nil) != tt.wantErr {
				t.Fatalf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// No AlertManagerID and no matching existing silence: a plain POST with an empty id.
func TestUpsertSilence_CreatesNewSilence(t *testing.T) {
	var postedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == silencesPath:
			writeJSON(t, w, []map[string]any{})
		case r.Method == http.MethodPost && r.URL.Path == silencesPath:
			var body struct {
				ID string `json:"id"`
			}
			decodeJSON(t, r, &body)
			postedID = body.ID
			writeJSON(t, w, map[string]any{"silenceID": "new-silence-id"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	id, err := am.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "new-silence-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "new-silence-id")
	}

	if postedID != "" {
		t.Errorf("PostSilences was called with id = %q, want empty (new silence)", postedID)
	}
}

// A matching, non-expired existing silence is reused: its id is filled in and posted back.
func TestUpsertSilence_ReusesMatchingExistingSilence(t *testing.T) {
	var postedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == silencesPath:
			writeJSON(t, w, []map[string]any{
				{
					"id":     existingSilenceID,
					"status": map[string]any{"state": "active"},
					"matchers": []map[string]any{
						{"name": "alertname", "value": "TestAlert", "isEqual": true, "isRegex": false},
					},
					"comment":   "old",
					"createdBy": "old-author",
					"startsAt":  time.Now().Format(time.RFC3339),
					"endsAt":    time.Now().Add(time.Hour).Format(time.RFC3339),
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == silencesPath:
			var body struct {
				ID string `json:"id"`
			}
			decodeJSON(t, r, &body)
			postedID = body.ID
			writeJSON(t, w, map[string]any{"silenceID": existingSilenceID})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	s := newTestSilence("")

	id, err := am.UpsertSilence(context.Background(), s, nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != existingSilenceID {
		t.Errorf("UpsertSilence() id = %q, want %q", id, existingSilenceID)
	}

	if postedID != existingSilenceID {
		t.Errorf("PostSilences was called with id = %q, want %q", postedID, existingSilenceID)
	}

	if s.Status.AlertManagerID != existingSilenceID {
		t.Errorf("Silence.Status.AlertManagerID = %q, want %q", s.Status.AlertManagerID, existingSilenceID)
	}
}

// An expired existing silence with matching matchers must not be reused.
func TestUpsertSilence_SkipsExpiredSilence(t *testing.T) {
	var postedID string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == silencesPath:
			writeJSON(t, w, []map[string]any{
				{
					"id":     "expired-id",
					"status": map[string]any{"state": "expired"},
					"matchers": []map[string]any{
						{"name": "alertname", "value": "TestAlert", "isEqual": true, "isRegex": false},
					},
					"comment":   "old",
					"createdBy": "old-author",
					"startsAt":  time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
					"endsAt":    time.Now().Add(-time.Hour).Format(time.RFC3339),
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == silencesPath:
			var body struct {
				ID string `json:"id"`
			}
			decodeJSON(t, r, &body)
			postedID = body.ID
			writeJSON(t, w, map[string]any{"silenceID": "brand-new-id"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	id, err := am.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "brand-new-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "brand-new-id")
	}

	if postedID != "" {
		t.Errorf("PostSilences was called with id = %q, want empty (expired silence must not be reused)", postedID)
	}
}

// A Silence that already carries an AlertManagerID skips the lookup and posts the known id directly.
func TestUpsertSilence_UpdatesKnownSilence(t *testing.T) {
	getCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == silencesPath:
			getCalled = true
			writeJSON(t, w, []map[string]any{})
		case r.Method == http.MethodPost && r.URL.Path == silencesPath:
			var body struct {
				ID string `json:"id"`
			}
			decodeJSON(t, r, &body)
			if body.ID != "known-id" {
				t.Errorf("PostSilences id = %q, want %q", body.ID, "known-id")
			}
			writeJSON(t, w, map[string]any{"silenceID": "known-id"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	id, err := am.UpsertSilence(context.Background(), newTestSilence("known-id"), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "known-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "known-id")
	}

	if getCalled {
		t.Errorf("GetSilences was called even though AlertManagerID was already known")
	}
}

func TestDeleteSilence(t *testing.T) {
	deletedID := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method %s", r.Method)
		}

		deletedID = r.URL.Path

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	if err := am.DeleteSilence("some-id"); err != nil {
		t.Fatalf("DeleteSilence() error = %v", err)
	}

	if deletedID != "/api/v2/silence/some-id" {
		t.Errorf("DeleteSilence() called path = %q, want %q", deletedID, "/api/v2/silence/some-id")
	}
}

func TestGetSilences_PassesFilter(t *testing.T) {
	var gotFilter []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotFilter = r.URL.Query()["filter"]
		writeJSON(t, w, []map[string]any{})
	}))
	defer server.Close()

	am := newTestAlertManager(t, server.URL)

	filter := []string{"alertname=TestAlert"}

	if _, err := am.GetSilences(filter); err != nil {
		t.Fatalf("GetSilences() error = %v", err)
	}

	if len(gotFilter) != 1 || gotFilter[0] != filter[0] {
		t.Errorf("GetSilences() filter query = %v, want %v", gotFilter, filter)
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("failed to write json response: %v", err)
	}
}

func decodeJSON(t *testing.T, r *http.Request, v any) {
	t.Helper()

	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		t.Fatalf("failed to decode json request: %v", err)
	}
}
