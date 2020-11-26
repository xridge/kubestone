/*
Copyright 2019 The xridge kubestone contributors.

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

package pgbench

import (
	"context"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/controllers/common"
	"github.com/xridge/kubestone/pkg/k8s"
	batchv1 "k8s.io/api/batch/v1"
)

// Reconciler reconciles a Pgbench object
type Reconciler struct {
	K8S k8s.Access
	Log logr.Logger
}

// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=pgbenches,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=pgbenches/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=pgbenches/finalizers,verbs=update

// Reconcile creates pgbench job(s) based on the custom resource(s)
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var cr perfv1alpha1.Pgbench
	ctx := context.Background()
	commonCr := common.CustomResource{Rto: &cr, Status: cr.Status, NewJobFn: func() *batchv1.Job { return NewJob(&cr) }}
	commonReconciler := common.Reconciler{K8S: r.K8S, Log: r.Log}

	if reconciled, err := commonReconciler.IsCrReconciled(ctx, commonCr, req); reconciled || err != nil {
		return ctrl.Result{}, err
	}

	if err := commonReconciler.MarkCrAsRunning(ctx, commonCr); err != nil {
		return ctrl.Result{}, err
	}

	jobFinished, err := commonReconciler.HandleJob(ctx, commonCr, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !jobFinished {
		// Wait for the job to be completed
		return ctrl.Result{Requeue: true}, nil
	}

	// The cr could have been modified since the last time we got it
	if err := r.K8S.Client.Get(ctx, req.NamespacedName, &cr); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	if err := commonReconciler.MarkCrAsCompleted(ctx, commonCr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the Reconciler with the provided manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&perfv1alpha1.Pgbench{}).
		Complete(r)
}
