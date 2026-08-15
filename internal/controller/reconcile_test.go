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

package controller

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	monitoringv1alpha1 "github.com/silence-operator/silence-operator/api/v1alpha1"
)

const (
	testAlert         = "TestAlert"
	alertNameLabel    = "alertname"
	testComment       = "test silence"
	existingID        = "existing-id"
	reconcileInterval = time.Minute
)

// --- Silence / client fixtures ---

// newSilence builds a Silence object for tests, optionally customized by mutate.
func newSilence(name string, mutate func(*monitoringv1alpha1.Silence)) *monitoringv1alpha1.Silence {
	s := &monitoringv1alpha1.Silence{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: monitoringv1alpha1.SilenceSpec{
			Comment:  testComment,
			Matchers: monitoringv1alpha1.Matchers{{Name: alertNameLabel, Value: testAlert, IsEqual: true}},
		},
	}
	if mutate != nil {
		mutate(s)
	}
	return s
}

// withFinalizer sets the finalizer Reconcile adds on a fresh object, so tests can jump
// straight to the alertmanager-facing branches instead of the finalizer-add branch.
func withFinalizer(s *monitoringv1alpha1.Silence) {
	s.Finalizers = []string{monitoringv1alpha1.SilenceFinalizer}
}

// newFakeClient returns a Silence-aware fake client seeded with objs, with the status
// subresource enabled so r.Status().Update calls behave like the real API server.
func newFakeClient(t *testing.T, opts ...func(*fake.ClientBuilder)) client.Client {
	t.Helper()

	scheme := runtime.NewScheme()
	if err := monitoringv1alpha1.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme() error = %v", err)
	}

	b := fake.NewClientBuilder().WithScheme(scheme).WithStatusSubresource(&monitoringv1alpha1.Silence{})
	for _, opt := range opts {
		opt(b)
	}

	return b.Build()
}

func withObjects(objs ...client.Object) func(*fake.ClientBuilder) {
	return func(b *fake.ClientBuilder) { *b = *b.WithObjects(objs...) }
}

func withInterceptor(funcs interceptor.Funcs) func(*fake.ClientBuilder) {
	return func(b *fake.ClientBuilder) { *b = *b.WithInterceptorFuncs(funcs) }
}

// getSilence fetches name from c, failing the test if it's missing.
func getSilence(t *testing.T, c client.Client, name string) *monitoringv1alpha1.Silence {
	t.Helper()

	got := &monitoringv1alpha1.Silence{}
	if err := c.Get(context.Background(), types.NamespacedName{Name: name}, got); err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	return got
}

// --- Alertmanager fixture ---

// callLog counts alertmanager calls by name (DeleteSilence keys on its id too), so a test can
// assert a call was (or wasn't) made without a bespoke bool flag per case.
type callLog map[string]int

// fakeAlertManager is a configurable AlertManagerClient double. A method with no matching
// *Func field fails the test, so "must not be called" just means leaving it unset.
type fakeAlertManager struct {
	t     *testing.T
	calls callLog

	getSilenceFunc    func(id string) (*silence.GetSilenceOK, error)
	upsertSilenceFunc func(ctx context.Context, s *monitoringv1alpha1.Silence, startsAt *strfmt.DateTime) (string, error)
	deleteSilenceFunc func(id string) error
}

func newFakeAlertManager(t *testing.T) *fakeAlertManager {
	t.Helper()
	return &fakeAlertManager{t: t, calls: callLog{}}
}

func (f *fakeAlertManager) GetSilence(id string) (*silence.GetSilenceOK, error) {
	f.calls["GetSilence"]++
	if f.getSilenceFunc == nil {
		f.t.Fatalf("unexpected GetSilence(%q)", id)
	}
	return f.getSilenceFunc(id)
}

func (f *fakeAlertManager) UpsertSilence(ctx context.Context, s *monitoringv1alpha1.Silence, startsAt *strfmt.DateTime) (string, error) {
	f.calls["UpsertSilence"]++
	if f.upsertSilenceFunc == nil {
		f.t.Fatal("unexpected UpsertSilence call")
	}
	return f.upsertSilenceFunc(ctx, s, startsAt)
}

func (f *fakeAlertManager) DeleteSilence(id string) error {
	f.calls["DeleteSilence:"+id]++
	if f.deleteSilenceFunc == nil {
		f.t.Fatalf("unexpected DeleteSilence(%q)", id)
	}
	return f.deleteSilenceFunc(id)
}

// gettableSilence builds the payload GetSilence returns.
func gettableSilence(state string, startsAt, endsAt time.Time) *models.GettableSilence {
	sa, ea := strfmt.DateTime(startsAt), strfmt.DateTime(endsAt)
	return &models.GettableSilence{
		Status:  &models.SilenceStatus{State: &state},
		Silence: models.Silence{StartsAt: &sa, EndsAt: &ea},
	}
}

// --- decideSilence: pure, no fixtures at all ---

func TestDecideSilence(t *testing.T) {
	now := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	startsAt := strfmt.DateTime(now.Add(-time.Hour))

	sameGen := newSilence("s", withFinalizer)
	sameGen.Generation, sameGen.Status.LastAppliedGeneration = 1, 1

	changedGen := newSilence("s", withFinalizer)
	changedGen.Generation, changedGen.Status.LastAppliedGeneration = 2, 1

	tests := []struct {
		name  string
		obj   *monitoringv1alpha1.Silence
		state *models.GettableSilence
		want  silenceDecision
	}{
		{
			name: "no state yet: create fresh",
			obj:  sameGen,
			want: silenceDecision{},
		},
		{
			name:  "expired: recreate with the reported startsAt",
			obj:   sameGen,
			state: gettableSilence(models.SilenceStatusStateExpired, time.Time(startsAt), now.Add(time.Hour)),
			want:  silenceDecision{startsAt: &startsAt, logMsg: "silence expired, updating expireAt"},
		},
		{
			name:  "generation changed: update regardless of expiry",
			obj:   changedGen,
			state: gettableSilence(models.SilenceStatusStateActive, time.Time(startsAt), now.Add(10*reconcileInterval)),
			want:  silenceDecision{startsAt: &startsAt, logMsg: "updating alertmanager silence"},
		},
		{
			name:  "same generation, comfortably before expiry: skip",
			obj:   sameGen,
			state: gettableSilence(models.SilenceStatusStateActive, time.Time(startsAt), now.Add(10*reconcileInterval)),
			want:  silenceDecision{skip: true},
		},
		{
			name:  "same generation, inside the extend window: upsert",
			obj:   sameGen,
			state: gettableSilence(models.SilenceStatusStateActive, time.Time(startsAt), now.Add(reconcileInterval)),
			want:  silenceDecision{startsAt: &startsAt},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := decideSilence(tt.obj, tt.state, reconcileInterval, now); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decideSilence() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// --- fetchSilenceState: only needs the 3-method fake, no k8s client ---

func TestFetchSilenceState(t *testing.T) {
	t.Run("no id yet: returns nil without calling GetSilence", func(t *testing.T) {
		obj := newSilence("s", withFinalizer)
		r := &SilenceReconciler{AlertManager: newFakeAlertManager(t)}

		if got := r.fetchSilenceState(context.Background(), obj); got != nil {
			t.Errorf("fetchSilenceState() = %v, want nil", got)
		}
	})

	t.Run("succeeds on the first attempt", func(t *testing.T) {
		am := newFakeAlertManager(t)
		payload := gettableSilence(models.SilenceStatusStateActive, time.Now(), time.Now().Add(time.Hour))
		am.getSilenceFunc = func(string) (*silence.GetSilenceOK, error) {
			return &silence.GetSilenceOK{Payload: payload}, nil
		}

		obj := newSilence("s", func(s *monitoringv1alpha1.Silence) {
			withFinalizer(s)
			s.Status.AlertManagerID = existingID
		})
		r := &SilenceReconciler{AlertManager: am, GetSilenceAttempts: 1}

		if got := r.fetchSilenceState(context.Background(), obj); got != payload {
			t.Errorf("fetchSilenceState() = %v, want %v", got, payload)
		}
	})

	t.Run("resets the id once attempts are exhausted", func(t *testing.T) {
		am := newFakeAlertManager(t)
		am.getSilenceFunc = func(string) (*silence.GetSilenceOK, error) { return nil, errors.New("boom") }

		obj := newSilence("s", func(s *monitoringv1alpha1.Silence) {
			withFinalizer(s)
			s.Status.AlertManagerID = "stale-id"
		})
		r := &SilenceReconciler{AlertManager: am, GetSilenceAttempts: 2, GetSilenceInterval: time.Millisecond}

		if got := r.fetchSilenceState(context.Background(), obj); got != nil {
			t.Errorf("fetchSilenceState() = %v, want nil", got)
		}
		if am.calls["GetSilence"] != 2 {
			t.Errorf("GetSilence called %d times, want 2", am.calls["GetSilence"])
		}
		if obj.Status.AlertManagerID != "" {
			t.Errorf("AlertManagerID = %q, want reset to empty", obj.Status.AlertManagerID)
		}
	})

	t.Run("zero attempts resets the id without calling GetSilence", func(t *testing.T) {
		obj := newSilence("s", func(s *monitoringv1alpha1.Silence) {
			withFinalizer(s)
			s.Status.AlertManagerID = "stale-id"
		})
		r := &SilenceReconciler{AlertManager: newFakeAlertManager(t)} // any call fails the test

		if got := r.fetchSilenceState(context.Background(), obj); got != nil {
			t.Errorf("fetchSilenceState() = %v, want nil", got)
		}
		if obj.Status.AlertManagerID != "" {
			t.Error("AlertManagerID was not reset")
		}
	})
}

func TestReconcile_MissingObjectIsIgnored(t *testing.T) {
	r := &SilenceReconciler{Client: newFakeClient(t), AlertManager: newFakeAlertManager(t)}

	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: "missing"}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res != (reconcile.Result{}) {
		t.Errorf("Reconcile() result = %v, want zero value", res)
	}
}

func TestReconcile_AddsFinalizerWithoutTouchingAlertManager(t *testing.T) {
	obj := newSilence("no-finalizer", nil)
	c := newFakeClient(t, withObjects(obj))
	r := &SilenceReconciler{Client: c, AlertManager: newFakeAlertManager(t), Interval: reconcileInterval}

	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res.RequeueAfter != reconcileInterval {
		t.Errorf("RequeueAfter = %v, want %v", res.RequeueAfter, reconcileInterval)
	}
	if got := getSilence(t, c, obj.Name); !controllerutil.ContainsFinalizer(got, monitoringv1alpha1.SilenceFinalizer) {
		t.Error("finalizer was not added")
	}
}

func TestReconcile_SuspendedSkipsAlertManager(t *testing.T) {
	obj := newSilence("suspended", func(s *monitoringv1alpha1.Silence) {
		withFinalizer(s)
		s.Spec.Suspend = true
	})
	c := newFakeClient(t, withObjects(obj))
	r := &SilenceReconciler{Client: c, AlertManager: newFakeAlertManager(t), Interval: reconcileInterval}

	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res != (reconcile.Result{}) {
		t.Errorf("Reconcile() result = %v, want zero value", res)
	}
}

func TestReconcile_CreatesNewSilenceAndUpdatesStatus(t *testing.T) {
	obj := newSilence("new-silence", withFinalizer)
	c := newFakeClient(t, withObjects(obj))

	am := newFakeAlertManager(t)
	am.upsertSilenceFunc = func(context.Context, *monitoringv1alpha1.Silence, *strfmt.DateTime) (string, error) {
		return "new-id", nil
	}
	r := &SilenceReconciler{Client: c, AlertManager: am, Interval: reconcileInterval}

	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res.RequeueAfter != reconcileInterval {
		t.Errorf("RequeueAfter = %v, want %v", res.RequeueAfter, reconcileInterval)
	}

	got := getSilence(t, c, obj.Name)
	if got.Status.AlertManagerID != "new-id" {
		t.Errorf("Status.AlertManagerID = %q, want %q", got.Status.AlertManagerID, "new-id")
	}
	if got.Status.LastAppliedGeneration != got.Generation {
		t.Errorf("LastAppliedGeneration = %d, want %d (=Generation)", got.Status.LastAppliedGeneration, got.Generation)
	}
}

func TestReconcile_ExtendsSilenceWhoseGenerationChanged(t *testing.T) {
	obj := newSilence("changed-gen", func(s *monitoringv1alpha1.Silence) {
		withFinalizer(s)
		s.Status.AlertManagerID = existingID
	})
	c := newFakeClient(t, withObjects(obj))

	am := newFakeAlertManager(t)
	am.getSilenceFunc = func(string) (*silence.GetSilenceOK, error) {
		payload := gettableSilence(models.SilenceStatusStateActive, time.Now(), time.Now().Add(time.Hour))
		return &silence.GetSilenceOK{Payload: payload}, nil
	}
	am.upsertSilenceFunc = func(context.Context, *monitoringv1alpha1.Silence, *strfmt.DateTime) (string, error) {
		return existingID, nil // unchanged id: Reconcile must not touch Status again
	}
	r := &SilenceReconciler{Client: c, AlertManager: am, Interval: reconcileInterval, GetSilenceAttempts: 1}

	res, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res.RequeueAfter != reconcileInterval {
		t.Errorf("RequeueAfter = %v, want %v", res.RequeueAfter, reconcileInterval)
	}
}

func TestReconcile_DeletionRemovesAlertManagerSilenceAndFinalizer(t *testing.T) {
	obj := newSilence("deleting", func(s *monitoringv1alpha1.Silence) {
		withFinalizer(s)
		s.Status.AlertManagerID = "to-delete-id"
	})
	c := newFakeClient(t, withObjects(obj))
	ctx := context.Background()
	if err := c.Delete(ctx, obj); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	am := newFakeAlertManager(t)
	am.deleteSilenceFunc = func(string) error { return nil }
	r := &SilenceReconciler{Client: c, AlertManager: am, Interval: reconcileInterval}

	res, err := r.Reconcile(ctx, reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}})
	if err != nil {
		t.Fatalf("Reconcile() error = %v", err)
	}
	if res != (reconcile.Result{}) {
		t.Errorf("Reconcile() result = %v, want zero value", res)
	}
	if am.calls["DeleteSilence:to-delete-id"] == 0 {
		t.Error("DeleteSilence was not called")
	}
	if err := c.Get(ctx, types.NamespacedName{Name: obj.Name}, &monitoringv1alpha1.Silence{}); !apierrors.IsNotFound(err) {
		t.Errorf("Get() after finalizer removal error = %v, want NotFound", err)
	}
}

func TestReconcile_CleansUpAlertManagerSilenceWhenStatusUpdateFails(t *testing.T) {
	obj := newSilence("status-update-fails", withFinalizer)
	c := newFakeClient(t, withObjects(obj), withInterceptor(interceptor.Funcs{
		SubResourceUpdate: func(context.Context, client.Client, string, client.Object, ...client.SubResourceUpdateOption) error {
			return apierrors.NewConflict(monitoringv1alpha1.GroupVersion.WithResource("silences").GroupResource(), obj.Name, nil)
		},
	}))

	am := newFakeAlertManager(t)
	am.upsertSilenceFunc = func(context.Context, *monitoringv1alpha1.Silence, *strfmt.DateTime) (string, error) {
		return "orphan-id", nil
	}
	am.deleteSilenceFunc = func(string) error { return nil }
	r := &SilenceReconciler{Client: c, AlertManager: am, Interval: reconcileInterval}

	if _, err := r.Reconcile(context.Background(), reconcile.Request{NamespacedName: types.NamespacedName{Name: obj.Name}}); err == nil {
		t.Fatal("Reconcile() error = nil, want the status update error")
	}
	if am.calls["DeleteSilence:orphan-id"] == 0 {
		t.Error("the orphaned alertmanager silence was not cleaned up")
	}
}
