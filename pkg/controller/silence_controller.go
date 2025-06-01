/*
Copyright 2024.

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
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	monitoringv1alpha1 "github.com/silence-operator/silence-operator/api/v1alpha1"
	"github.com/silence-operator/silence-operator/internal/alertmanager"
)

// SilenceReconciler reconciles a Silence object
type SilenceReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	AlertManager *alertmanager.AlertManager
	Interval     time.Duration
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
	if err := r.Get(ctx, req.NamespacedName, obj); err != nil {
		reconciliationCompleted = false

		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Handle object deletion
	if !obj.DeletionTimestamp.IsZero() {
		if obj.Status.AlertManagerID != "" {
			log.Info("deleting alertmanager silence")

			err := r.AlertManager.DeleteSilence(obj.Status.AlertManagerID)
			if err != nil {
				reconciliationCompleted = false
				log.Error(err, "unable to delete silence in alertmanager")
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

			return ctrl.Result{Requeue: true}, err
		}

		log.Info("successfully added finalizer to silence")

		return ctrl.Result{Requeue: true}, nil
	}

	if obj.Spec.Suspend {
		log.Info("reconciliation is suspended")

		return ctrl.Result{}, nil
	}

	var startsAt *strfmt.DateTime

	if obj.Status.AlertManagerID == "" {
		log.Info("silence is not created yet, creating")
	} else {
		response, err := r.AlertManager.GetSilence(obj.Status.AlertManagerID)
		if err != nil {
			log.Info("unable to get alertmanager silence")

			obj.Status.AlertManagerID = ""
		} else {
			s := response.GetPayload()
			startsAt = s.StartsAt

			if *s.Status.State == models.SilenceStatusStateExpired {
				log.Info("silence expired, updating expireAt")
			} else {
				if obj.ObjectMeta.Generation != obj.Status.LastAppliedGeneration {
					log.Info("updating alertmanager silence")
				} else {
					// Extend silence if three or less reconciliations left
					deadline := time.Now().Add(r.Interval * 3)

					if deadline.Before(time.Time(*s.EndsAt)) {
						log.Info("no need for reconciliation")
						reconciliationCompleted = false

						return ctrl.Result{RequeueAfter: r.Interval}, nil
					}
				}
			}
		}
	}

	id, err := r.AlertManager.UpsertSilence(ctx, obj, startsAt)
	if err != nil {
		reconciliationCompleted = false

		log.Error(err, "unable to upsert silence")

		return ctrl.Result{RequeueAfter: r.Interval}, err
	}

	if obj.Status.AlertManagerID == id {
		return ctrl.Result{RequeueAfter: r.Interval}, err
	}

	log.Info("updating status of the silence object")

	obj.Status.AlertManagerID = id
	obj.Status.LastAppliedGeneration = obj.ObjectMeta.Generation

	err = r.Status().Update(ctx, obj)
	if err != nil {
		reconciliationCompleted = false

		log.Error(err, "unable to update status")

		log.Info("cleaning up alertmanager silence")

		err2 := r.AlertManager.DeleteSilence(id)
		if err2 != nil {
			log.Error(err2, "unable to delete alertmanager silence")
		}

		return ctrl.Result{Requeue: true}, err
	}

	return ctrl.Result{}, err
}

// SetupWithManager sets up the controller with the Manager.
func (r *SilenceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.Silence{}).
		Named("silence").
		Owns(&monitoringv1alpha1.Silence{}).
		Complete(r)
}
