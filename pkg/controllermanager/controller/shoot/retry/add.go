// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package retry

import (
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
)

// ControllerName is the name of this controller.
const ControllerName = "shoot-retry"

// AddToManager adds Reconciler to the given manager.
func (r *Reconciler) AddToManager(mgr manager.Manager) error {
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}

	return builder.
		ControllerManagedBy(mgr).
		Named(ControllerName).
		For(&gardencorev1beta1.Shoot{}, builder.WithPredicates(r.ShootPredicate())).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: ptr.Deref(r.Config.ConcurrentSyncs, 0),
		}).
		Complete(r)
}

// ShootPredicate reacts only on 'CREATE' and 'UPDATE' Shoot events.
func (r *Reconciler) ShootPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(_ event.CreateEvent) bool {
			return true
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			shoot, ok := e.ObjectNew.(*gardencorev1beta1.Shoot)
			if !ok {
				return false
			}

			oldShoot, ok := e.ObjectOld.(*gardencorev1beta1.Shoot)
			if !ok {
				return false
			}

			return shootFailedDueToRateLimits(shoot) && !isShootFailed(oldShoot)
		},
		DeleteFunc:  func(_ event.DeleteEvent) bool { return false },
		GenericFunc: func(_ event.GenericEvent) bool { return false },
	}
}

func isShootFailed(shoot *gardencorev1beta1.Shoot) bool {
	lastOperation := shoot.Status.LastOperation

	return lastOperation != nil &&
		lastOperation.State == gardencorev1beta1.LastOperationStateFailed &&
		shoot.Generation == shoot.Status.ObservedGeneration
}

func shootFailedDueToRateLimits(shoot *gardencorev1beta1.Shoot) bool {
	return isShootFailed(shoot) && v1beta1helper.HasErrorCode(shoot.Status.LastErrors, gardencorev1beta1.ErrorInfraRateLimitsExceeded)
}
