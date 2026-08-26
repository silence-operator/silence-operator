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
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/silence-operator/silence-operator/api/v1alpha1"
)

const (
	silencesPath             = "/api/v2/silences"
	existingSilenceID        = "existing-id"
	alertNameLabel           = "alertname"
	testAlertName            = "TestAlert"
	testAlertmanagerHostPort = "alertmanager.default:9093"
	httpScheme               = "http"
)

func newTestSilence(amID string) *v1alpha1.Silence {
	s := &v1alpha1.Silence{
		Spec: v1alpha1.SilenceSpec{
			Comment: "test silence",
			Matchers: v1alpha1.Matchers{
				{Name: alertNameLabel, Value: testAlertName, IsEqual: true, IsRegex: false},
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

var errRoundTripStubbed = errors.New("roundTripFunc: no real request sent")

// captureRequest swaps in a fake transport that records the outgoing request instead
// of sending it, issues a GetSilences call to trigger one, and returns it.
func captureRequest(t *testing.T, mgr *AlertManager) *http.Request {
	t.Helper()

	rt, ok := mgr.am.Transport.(*httptransport.Runtime)
	if !ok {
		t.Fatalf("Transport = %T, want *httptransport.Runtime", mgr.am.Transport)
	}

	var captured *http.Request

	rt.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		captured = req
		return nil, errRoundTripStubbed
	})

	_, _ = mgr.GetSilences(context.Background(), nil)

	if captured == nil {
		t.Fatal("no request reached the fake transport")
	}

	return captured
}

func TestNew_RejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "unparseable url", url: "http://[::1]:namedport"},
		{name: "empty url", url: ""},
		{name: "scheme with no host", url: "http://"},
		{name: "misspelled scheme", url: "htttps://" + testAlertmanagerHostPort},
		{name: "unsupported scheme", url: "ftp://" + testAlertmanagerHostPort},
		{name: "basic auth in url", url: "http://user:pass@" + testAlertmanagerHostPort},
		{name: "username only in url", url: "http://user@" + testAlertmanagerHostPort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(&Config{URL: tt.url})
			if err == nil {
				t.Fatalf("New() error = nil, want error for url %q", tt.url)
			}
		})
	}
}

// Proves New()'s client actually reaches the configured host end to end.
func TestNew_ResolvesBareHostPort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, []map[string]any{})
	}))
	t.Cleanup(server.Close)

	bareHostPort := strings.TrimPrefix(server.URL, "http://")

	mgr := newTestAlertManager(t, bareHostPort)

	_, err := mgr.GetSilences(context.Background(), nil)
	if err != nil {
		t.Fatalf("GetSilences() error = %v, want nil", err)
	}
}

// Asserts the scheme/host on the outgoing request directly, instead of inferring it
// from whether a live connection happens to succeed or fail.
func TestNew_SchemeAndHostReachTheRequest(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantScheme string
		wantHost   string
	}{
		{name: "explicit http scheme", url: "http://" + testAlertmanagerHostPort, wantScheme: httpScheme, wantHost: testAlertmanagerHostPort},
		{name: "explicit https scheme", url: "https://" + testAlertmanagerHostPort, wantScheme: "https", wantHost: testAlertmanagerHostPort},
		{name: "bare host:port defaults to http", url: testAlertmanagerHostPort, wantScheme: httpScheme, wantHost: testAlertmanagerHostPort},
		{name: "bare ip:port defaults to http", url: "127.0.0.1:9093", wantScheme: httpScheme, wantHost: "127.0.0.1:9093"},
		{name: "uppercase scheme is recognized as already having one", url: "HTTPS://" + testAlertmanagerHostPort, wantScheme: "https", wantHost: testAlertmanagerHostPort},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newTestAlertManager(t, tt.url)

			req := captureRequest(t, mgr)

			if req.URL.Scheme != tt.wantScheme {
				t.Errorf("request scheme = %q, want %q", req.URL.Scheme, tt.wantScheme)
			}

			if req.URL.Host != tt.wantHost {
				t.Errorf("request host = %q, want %q", req.URL.Host, tt.wantHost)
			}
		})
	}
}

// postedSilenceBody is what the wire body of a PostSilences call decodes to.
type postedSilenceBody struct {
	ID        string          `json:"id"`
	Comment   string          `json:"comment"`
	CreatedBy string          `json:"createdBy"`
	StartsAt  time.Time       `json:"startsAt"`
	EndsAt    time.Time       `json:"endsAt"`
	Matchers  []postedMatcher `json:"matchers"`
}

type postedMatcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsEqual bool   `json:"isEqual"`
	IsRegex bool   `json:"isRegex"`
}

// silencesServerCapture records what a newSilencesServer handler observed.
type silencesServerCapture struct {
	getCalled  bool
	getFilter  []string
	postedBody []byte
}

func (c *silencesServerCapture) postedSilence(t *testing.T) postedSilenceBody {
	t.Helper()

	var body postedSilenceBody

	err := json.Unmarshal(c.postedBody, &body)
	if err != nil {
		t.Fatalf("failed to decode posted silence: %v", err)
	}

	return body
}

// activeSilence is a newSilencesServer getPayload entry matching newTestSilence's matcher.
func activeSilence(id string) map[string]any {
	return silencePayload(id, "active", time.Now(), time.Now().Add(time.Hour))
}

// expiredSilence is a newSilencesServer getPayload entry matching newTestSilence's matcher.
func expiredSilence(id string) map[string]any {
	return silencePayload(id, "expired", time.Now().Add(-2*time.Hour), time.Now().Add(-time.Hour))
}

func silencePayload(id, state string, startsAt, endsAt time.Time) map[string]any {
	return map[string]any{
		"id":     id,
		"status": map[string]any{"state": state},
		"matchers": []map[string]any{
			{"name": alertNameLabel, "value": testAlertName, "isEqual": true, "isRegex": false},
		},
		"comment":   "old",
		"createdBy": "old-author",
		"startsAt":  startsAt.Format(time.RFC3339),
		"endsAt":    endsAt.Format(time.RFC3339),
	}
}

// newSilencesServer stubs GET/POST /api/v2/silences: GET returns getPayload verbatim,
// POST always returns postSilenceID.
func newSilencesServer(t *testing.T, getPayload []map[string]any, postSilenceID string) (*httptest.Server, *silencesServerCapture) {
	t.Helper()

	return newSilencesServerHandler(t, getPayload, postSilenceID, false)
}

// newFailingPostSilencesServer is newSilencesServer, but POST /api/v2/silences fails with 500.
func newFailingPostSilencesServer(t *testing.T, getPayload []map[string]any) (*httptest.Server, *silencesServerCapture) {
	t.Helper()

	return newSilencesServerHandler(t, getPayload, "", true)
}

func newSilencesServerHandler(t *testing.T, getPayload []map[string]any, postSilenceID string, postFails bool) (*httptest.Server, *silencesServerCapture) {
	t.Helper()

	capture := &silencesServerCapture{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == silencesPath:
			capture.getCalled = true
			capture.getFilter = r.URL.Query()["filter"]

			writeJSON(t, w, getPayload)
		case r.Method == http.MethodPost && r.URL.Path == silencesPath:
			if postFails {
				http.Error(w, "boom", http.StatusInternalServerError)
				return
			}

			capture.postedBody = readAll(t, r.Body)
			writeJSON(t, w, map[string]any{"silenceID": postSilenceID})
		default:
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	}))
	t.Cleanup(server.Close)

	return server, capture
}

// No AlertManagerID and no matching existing silence: a plain POST with an empty id.
func TestUpsertSilence_CreatesNewSilence(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "new-silence-id")

	mgr := newTestAlertManager(t, server.URL)

	before := time.Now()

	id, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "new-silence-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "new-silence-id")
	}

	posted := capture.postedSilence(t)

	if posted.ID != "" {
		t.Errorf("posted id = %q, want empty (new silence)", posted.ID)
	}

	if posted.CreatedBy != "test-author" {
		t.Errorf("posted createdBy = %q, want %q", posted.CreatedBy, "test-author")
	}

	if !strings.HasSuffix(posted.Comment, "Instance: test-instance") {
		t.Errorf("posted comment = %q, want suffix %q", posted.Comment, "Instance: test-instance")
	}

	if posted.StartsAt.Before(before.Add(-time.Second)) {
		t.Errorf("posted startsAt = %v, want ~now (nil startsAt defaults to now)", posted.StartsAt)
	}

	wantEndsAt := before.Add(time.Hour) // client.go anchors endsAt to time.Now(), not startsAt
	if posted.EndsAt.Sub(wantEndsAt).Abs() > time.Second {
		t.Errorf("posted endsAt = %v, want ~%v (now + SilenceDuration)", posted.EndsAt, wantEndsAt)
	}
}

// A non-nil startsAt (what the controller passes when extending a known silence) must
// be threaded through verbatim, not overridden by the nil-defaults-to-now path.
func TestUpsertSilence_ThreadsExplicitStartsAt(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "known-id")

	mgr := newTestAlertManager(t, server.URL)

	before := time.Now()
	explicitStartsAt := strfmt.DateTime(before.Add(-time.Hour))

	_, err := mgr.UpsertSilence(context.Background(), newTestSilence("known-id"), &explicitStartsAt)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	posted := capture.postedSilence(t)

	if diff := posted.StartsAt.Sub(time.Time(explicitStartsAt)); diff.Abs() > time.Second {
		t.Errorf("posted startsAt = %v, want %v", posted.StartsAt, time.Time(explicitStartsAt))
	}

	wantEndsAt := before.Add(time.Hour) // endsAt is anchored to now, not the explicit startsAt
	if posted.EndsAt.Sub(wantEndsAt).Abs() > time.Second {
		t.Errorf("posted endsAt = %v, want ~%v (now + SilenceDuration)", posted.EndsAt, wantEndsAt)
	}
}

// The existing-silence lookup filters by Matchers.String(), not some other encoding.
func TestUpsertSilence_FiltersExistingByMatcherString(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "id")

	mgr := newTestAlertManager(t, server.URL)

	_, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	want := []string{"alertname=TestAlert"}
	if !reflect.DeepEqual(capture.getFilter, want) {
		t.Errorf("GetSilences filter = %v, want %v", capture.getFilter, want)
	}
}

// A matching, non-expired existing silence is reused: its id is filled in and posted back.
func TestUpsertSilence_ReusesMatchingExistingSilence(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{
		activeSilence(existingSilenceID),
	}, existingSilenceID)

	mgr := newTestAlertManager(t, server.URL)

	s := newTestSilence("")

	id, err := mgr.UpsertSilence(context.Background(), s, nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != existingSilenceID {
		t.Errorf("UpsertSilence() id = %q, want %q", id, existingSilenceID)
	}

	if got := capture.postedSilence(t).ID; got != existingSilenceID {
		t.Errorf("posted id = %q, want %q", got, existingSilenceID)
	}

	if s.Status.AlertManagerID != existingSilenceID {
		t.Errorf("Silence.Status.AlertManagerID = %q, want %q", s.Status.AlertManagerID, existingSilenceID)
	}
}

// An expired existing silence with matching matchers must not be reused.
func TestUpsertSilence_SkipsExpiredSilence(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{
		expiredSilence("expired-id"),
	}, "brand-new-id")

	mgr := newTestAlertManager(t, server.URL)

	id, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "brand-new-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "brand-new-id")
	}

	if got := capture.postedSilence(t).ID; got != "" {
		t.Errorf("posted id = %q, want empty (expired silence must not be reused)", got)
	}
}

// An expired entry earlier in the list must be skipped (continue), not abort the
// scan (break) before a later, matching active entry is ever examined.
func TestUpsertSilence_SkipsExpiredThenReusesActive(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{
		expiredSilence("expired-id"),
		activeSilence(existingSilenceID),
	}, existingSilenceID)

	mgr := newTestAlertManager(t, server.URL)

	id, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != existingSilenceID {
		t.Errorf("UpsertSilence() id = %q, want %q", id, existingSilenceID)
	}

	if got := capture.postedSilence(t).ID; got != existingSilenceID {
		t.Errorf("posted id = %q, want %q", got, existingSilenceID)
	}
}

// PostSilences carries the full matcher array, not just a count.
func TestUpsertSilence_PostsAllMatchers(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "id")

	mgr := newTestAlertManager(t, server.URL)

	s := &v1alpha1.Silence{
		Spec: v1alpha1.SilenceSpec{
			Comment: "test silence",
			Matchers: v1alpha1.Matchers{
				{Name: alertNameLabel, Value: testAlertName, IsEqual: true, IsRegex: false},
				{Name: "severity", Value: "crit.*", IsEqual: false, IsRegex: true},
			},
		},
	}

	_, err := mgr.UpsertSilence(context.Background(), s, nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	want := []postedMatcher{
		{Name: alertNameLabel, Value: testAlertName, IsEqual: true, IsRegex: false},
		{Name: "severity", Value: "crit.*", IsEqual: false, IsRegex: true},
	}

	if got := capture.postedSilence(t).Matchers; !reflect.DeepEqual(got, want) {
		t.Errorf("posted matchers = %+v, want %+v", got, want)
	}
}

// A failure looking up existing silences must abort, not fall through to creating a new one.
func TestUpsertSilence_ReturnsErrorWhenLookupFails(t *testing.T) {
	mgr := newTestAlertManager(t, errServer(t).URL)

	_, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err == nil {
		t.Fatal("UpsertSilence() error = nil, want error when the existing-silence lookup fails")
	}
}

func TestUpsertSilence_ReturnsErrorWhenPostFails(t *testing.T) {
	server, _ := newFailingPostSilencesServer(t, []map[string]any{})

	mgr := newTestAlertManager(t, server.URL)

	_, err := mgr.UpsertSilence(context.Background(), newTestSilence(""), nil)
	if err == nil {
		t.Fatal("UpsertSilence() error = nil, want error when PostSilences fails")
	}
}

// A Silence that already carries an AlertManagerID skips the lookup and posts the known id directly.
func TestUpsertSilence_UpdatesKnownSilence(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "known-id")

	mgr := newTestAlertManager(t, server.URL)

	id, err := mgr.UpsertSilence(context.Background(), newTestSilence("known-id"), nil)
	if err != nil {
		t.Fatalf("UpsertSilence() error = %v", err)
	}

	if id != "known-id" {
		t.Errorf("UpsertSilence() id = %q, want %q", id, "known-id")
	}

	if got := capture.postedSilence(t).ID; got != "known-id" {
		t.Errorf("posted id = %q, want %q", got, "known-id")
	}

	if capture.getCalled {
		t.Errorf("GetSilences was called even though AlertManagerID was already known")
	}
}

func TestGetSilenceAndDeleteSilence_CallSilenceByIDPath(t *testing.T) {
	tests := []struct {
		name       string
		wantMethod string
		call       func(t *testing.T, mgr *AlertManager) // makes the request and checks its own return value
	}{
		{
			name:       "GetSilence",
			wantMethod: http.MethodGet,
			call: func(t *testing.T, mgr *AlertManager) {
				result, err := mgr.GetSilence(context.Background(), "some-id")
				if err != nil {
					t.Fatalf("GetSilence() error = %v", err)
				}

				if got := *result.GetPayload().ID; got != "some-id" {
					t.Errorf("GetSilence() payload id = %q, want %q", got, "some-id")
				}
			},
		},
		{
			name:       "DeleteSilence",
			wantMethod: http.MethodDelete,
			call: func(t *testing.T, mgr *AlertManager) {
				err := mgr.DeleteSilence(context.Background(), "some-id")
				if err != nil {
					t.Fatalf("DeleteSilence() error = %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
				gotPath = r.URL.Path

				writeJSON(t, w, activeSilence("some-id"))
			}))
			t.Cleanup(server.Close)

			mgr := newTestAlertManager(t, server.URL)

			tt.call(t, mgr)

			if gotMethod != tt.wantMethod {
				t.Errorf("%s() called method = %q, want %q", tt.name, gotMethod, tt.wantMethod)
			}

			if gotPath != "/api/v2/silence/some-id" {
				t.Errorf("%s() called path = %q, want %q", tt.name, gotPath, "/api/v2/silence/some-id")
			}
		})
	}
}

func TestGetSilences_PassesFilter(t *testing.T) {
	server, capture := newSilencesServer(t, []map[string]any{}, "")

	am := newTestAlertManager(t, server.URL)

	filter := []string{"alertname=TestAlert"}

	_, err := am.GetSilences(context.Background(), filter)
	if err != nil {
		t.Fatalf("GetSilences() error = %v", err)
	}

	if len(capture.getFilter) != 1 || capture.getFilter[0] != filter[0] {
		t.Errorf("GetSilences() filter query = %v, want %v", capture.getFilter, filter)
	}
}

// TestContextCancellationAbortsTheRequest proves each method threads ctx into the outbound
// call: a pre-canceled ctx must fail with context.Canceled before the server sees it.
func TestContextCancellationAbortsTheRequest(t *testing.T) {
	tests := []struct {
		name string
		call func(ctx context.Context, mgr *AlertManager) error
	}{
		{
			name: "GetSilence",
			call: func(ctx context.Context, mgr *AlertManager) error {
				_, err := mgr.GetSilence(ctx, "some-id")
				return err
			},
		},
		{
			name: "GetSilences",
			call: func(ctx context.Context, mgr *AlertManager) error {
				_, err := mgr.GetSilences(ctx, nil)
				return err
			},
		},
		{
			name: "DeleteSilence",
			call: func(ctx context.Context, mgr *AlertManager) error {
				return mgr.DeleteSilence(ctx, "some-id")
			},
		},
		{
			name: "UpsertSilence",
			call: func(ctx context.Context, mgr *AlertManager) error {
				_, err := mgr.UpsertSilence(ctx, newTestSilence("some-id"), nil)
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var serverHit bool

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				serverHit = true
				writeJSON(t, w, activeSilence("some-id"))
			}))
			t.Cleanup(server.Close)

			mgr := newTestAlertManager(t, server.URL)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			err := tt.call(ctx, mgr)
			if !errors.Is(err, context.Canceled) {
				t.Errorf("%s() error = %v, want context.Canceled", tt.name, err)
			}

			if serverHit {
				t.Errorf("%s() reached the server despite an already-canceled ctx", tt.name)
			}
		})
	}
}

// errServer always responds with 500, for exercising error-return paths.
func errServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	return server
}

func TestAlertManager_ReturnsErrorOnServerFailure(t *testing.T) {
	tests := []struct {
		name string
		call func(mgr *AlertManager) error
	}{
		{
			name: "GetSilence",
			call: func(mgr *AlertManager) error {
				_, err := mgr.GetSilence(context.Background(), "some-id")
				return err
			},
		},
		{
			name: "GetSilences",
			call: func(mgr *AlertManager) error {
				_, err := mgr.GetSilences(context.Background(), nil)
				return err
			},
		},
		{
			name: "DeleteSilence",
			call: func(mgr *AlertManager) error {
				return mgr.DeleteSilence(context.Background(), "some-id")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newTestAlertManager(t, errServer(t).URL)

			err := tt.call(mgr)
			if err == nil {
				t.Fatalf("%s() error = nil, want error on server failure", tt.name)
			}
		})
	}
}

func writeJSON(t *testing.T, w http.ResponseWriter, v any) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		t.Errorf("failed to write json response: %v", err)
	}
}

func readAll(t *testing.T, r io.Reader) []byte {
	t.Helper()

	body, err := io.ReadAll(r)
	if err != nil {
		t.Errorf("failed to read request body: %v", err)
	}

	return body
}
