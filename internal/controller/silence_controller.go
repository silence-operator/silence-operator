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

package controller

import (
	"context"
	"errors"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitoringv1alpha1 "github.com/silence-operator/silence-operator/api/v1alpha1"
)

// extendSilenceThresholdReconciles is how many reconciliations of headroom a silence must have
// left on its EndsAt before we bother extending it early, to avoid re-extending on every loop.
const extendSilenceThresholdReconciles = 3

// AlertManagerClient is the subset of alertmanager.AlertManager's API the reconciler depends on;
// depending on this instead of the concrete type lets tests substitute a fake.
type AlertManagerClient interface {
	GetSilence(ctx context.Context, id string) (*silence.GetSilenceOK, error)
	UpsertSilence(ctx context.Context, s *monitoringv1alpha1.Silence, startsAt *strfmt.DateTime) (string, error)
	DeleteSilence(ctx context.Context, id string) error
}

// SilenceReconciler reconciles a Silence object
type SilenceReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	AlertManager AlertManagerClient
	Interval     time.Duration

	GetSilenceAttempts int
	GetSilenceInterval time.Duration
}

// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=silences,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=silences/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.coreos.com,resources=silences/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *SilenceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	start := time.Now()

	reconciliationCompleted := true

	log := ctrl.LoggerFrom(ctx)

	defer func() {
		if reconciliationCompleted {
			end := time.Now()
			log.Info("reconciliation completed", "duration", end.Sub(start))
		}
	}()

	obj := &monitoringv1alpha1.Silence{}

	err := r.Get(ctx, req.NamespacedName, obj)
	if err != nil {
		reconciliationCompleted = false

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle object deletion
	if !obj.DeletionTimestamp.IsZero() {
		if obj.Status.AlertManagerID != "" {
			log.Info("deleting alertmanager silence", "am_id", obj.Status.AlertManagerID)

			err := r.AlertManager.DeleteSilence(ctx, obj.Status.AlertManagerID)
			if err != nil {
				reconciliationCompleted = false

				log.Error(err, "unable to delete silence in alertmanager", "am_id", obj.Status.AlertManagerID)
			}
		}

		if removed := controllerutil.RemoveFinalizer(obj, monitoringv1alpha1.SilenceFinalizer); removed {
			err := r.Update(ctx, obj)
			if err != nil {
				reconciliationCompleted = false

				log.Error(err, "unable to remove finalizer from silence")
			}

			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	if finalizerAdded := controllerutil.AddFinalizer(obj, monitoringv1alpha1.SilenceFinalizer); finalizerAdded {
		err := r.Update(ctx, obj)
		if err != nil {
			reconciliationCompleted = false

			log.Error(err, "unable to add finalizer to silence")

			return ctrl.Result{RequeueAfter: r.Interval}, err
		}

		log.Info("successfully added finalizer to silence")

		return ctrl.Result{RequeueAfter: r.Interval}, nil
	}

	if obj.Spec.Suspend {
		log.Info("reconciliation is suspended")

		return ctrl.Result{}, nil
	}

	state := r.fetchSilenceState(ctx, obj)

	d := decideSilence(obj, state, r.Interval, time.Now())
	if d.logMsg != "" {
		log.Info(d.logMsg, "am_id", obj.Status.AlertManagerID)
	}

	if d.skip {
		log.Info("no need for reconciliation")

		reconciliationCompleted = false

		return ctrl.Result{RequeueAfter: r.Interval}, nil
	}

	res, err := r.applyUpsert(ctx, obj, d.startsAt)
	if err != nil {
		reconciliationCompleted = false
	}

	return res, err
}

// fetchSilenceState resolves what alertmanager reports about obj's silence, retrying up to
// GetSilenceAttempts times. nil means unknown: no silence yet, or a lookup that gave up.
func (r *SilenceReconciler) fetchSilenceState(ctx context.Context, obj *monitoringv1alpha1.Silence) *models.GettableSilence {
	log := ctrl.LoggerFrom(ctx)

	if obj.Status.AlertManagerID == "" {
		log.Info("silence is not created yet, creating")
		return nil
	}

	lastErr := errors.New("no attempts configured to get alertmanager silence")

	for attempt := 1; attempt <= r.GetSilenceAttempts; attempt++ {
		log.Info("getting silence", "attempt", attempt, "am_id", obj.Status.AlertManagerID)

		response, err := r.AlertManager.GetSilence(ctx, obj.Status.AlertManagerID)
		if err == nil {
			return response.GetPayload()
		}

		lastErr = err

		time.Sleep(r.GetSilenceInterval)
	}

	// A lookup that never resolves is treated as "gone": reset the id so the caller creates one.
	log.Info("unable to get alertmanager silence", "am_id", obj.Status.AlertManagerID, "err", lastErr.Error())
	obj.Status.AlertManagerID = ""

	return nil
}

// silenceDecision is what Reconcile should do about obj's silence; decideSilence computes it
// from plain values, with no Kubernetes or AlertManager access.
type silenceDecision struct {
	skip     bool
	startsAt *strfmt.DateTime
	logMsg   string // non-empty: the caller logs this (with "am_id") before upserting
}

// decideSilence decides whether to skip reconciliation or upsert the silence for obj, given
// what fetchSilenceState found (state == nil: no silence yet, or its lookup gave up).
func decideSilence(obj *monitoringv1alpha1.Silence, state *models.GettableSilence, interval time.Duration, now time.Time) silenceDecision {
	if state == nil {
		return silenceDecision{}
	}

	if *state.Status.State == models.SilenceStatusStateExpired {
		return silenceDecision{startsAt: state.StartsAt, logMsg: "silence expired, updating expireAt"}
	}

	if obj.Generation != obj.Status.LastAppliedGeneration {
		return silenceDecision{startsAt: state.StartsAt, logMsg: "updating alertmanager silence"}
	}

	// Extend silence if extendSilenceThresholdReconciles or fewer reconciliations are left
	deadline := now.Add(interval * extendSilenceThresholdReconciles)
	if deadline.Before(time.Time(*state.EndsAt)) {
		return silenceDecision{skip: true}
	}

	return silenceDecision{startsAt: state.StartsAt}
}

// applyUpsert creates or updates obj's silence with startsAt and persists the id to Status.
// A failed status update deletes the silence it just wrote, keeping both systems in sync.
func (r *SilenceReconciler) applyUpsert(ctx context.Context, obj *monitoringv1alpha1.Silence, startsAt *strfmt.DateTime) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)

	id, err := r.AlertManager.UpsertSilence(ctx, obj, startsAt)
	if err != nil {
		log.Error(err, "unable to upsert silence", "am_id", obj.Status.AlertManagerID)

		return ctrl.Result{RequeueAfter: r.Interval}, err
	}

	if obj.Status.AlertManagerID == id {
		return ctrl.Result{RequeueAfter: r.Interval}, nil
	}

	log.Info("updating status of the silence object")

	obj.Status.AlertManagerID = id
	obj.Status.LastAppliedGeneration = obj.Generation

	err = r.Status().Update(ctx, obj)
	if err != nil {
		log.Error(err, "unable to update status")
		log.Info("cleaning up alertmanager silence")

		// A canceled ctx must not block this compensating delete: it would otherwise leak
		// id in alertmanager whenever the status update fails because ctx itself is done.
		err2 := r.AlertManager.DeleteSilence(context.WithoutCancel(ctx), id)
		if err2 != nil {
			log.Error(err2, "unable to delete alertmanager silence")
		}

		return ctrl.Result{RequeueAfter: r.Interval}, err
	}

	return ctrl.Result{RequeueAfter: r.Interval}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SilenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Silence{}).
		Named("silence").
		Owns(&monitoringv1alpha1.Silence{}).
		Complete(r)
}
