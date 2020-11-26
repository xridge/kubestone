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

package qperf

import (
	"context"

	"github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/xridge/kubestone/controllers/common"
	"github.com/xridge/kubestone/pkg/k8s"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

// Reconciler provides fields from manager to reconciler
type Reconciler struct {
	K8S k8s.Access
	Log logr.Logger
}

// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=qperves,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=qperves/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=qperves/finalizers,verbs=update

// Reconcile Qperf Benchmark Requests by creating:
//   - qperf server deployment
//   - qperf server service
//   - qperf client pod
// The creation of qperf client pod is postponed until the server
// deployment completes. Once the qperf client pod is completed,
// the server deployment and service objects are removed from k8s.
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	var cr perfv1alpha1.Qperf
	ctx := context.Background()
	commonCr := common.CustomResource{Rto: &cr, Status: cr.Status, NewJobFn: func() *batchv1.Job { return NewClientJob(&cr) }}
	commonReconciler := common.Reconciler{K8S: r.K8S, Log: r.Log}

	if reconciled, err := commonReconciler.IsCrReconciled(ctx, commonCr, req); reconciled || err != nil {
		return ctrl.Result{}, err
	}

	if err := commonReconciler.MarkCrAsCompleted(ctx, commonCr); err != nil {
		return ctrl.Result{}, err
	}

	serverDeployment := NewServerDeployment(&cr)
	if err := r.K8S.CreateWithReference(ctx, serverDeployment, &cr); err != nil {
		return ctrl.Result{}, err
	}

	serverService := NewServerService(&cr)
	if err := r.K8S.CreateWithReference(ctx, serverService, &cr); err != nil {
		return ctrl.Result{}, err
	}

	endpointReady, err := r.K8S.IsEndpointReady(
		types.NamespacedName{
			Namespace: cr.Namespace,
			Name:      cr.Name,
		},
	)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !endpointReady {
		// Wait for deployment to be connected to the service endpoint
		return ctrl.Result{Requeue: true}, nil
	}

	jobFinished, err := commonReconciler.HandleJob(ctx, commonCr, req)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !jobFinished {
		// Wait for the job to be completed
		return ctrl.Result{Requeue: true}, nil
	}

	if err := r.K8S.DeleteObject(ctx, serverService, &cr); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.K8S.DeleteObject(ctx, serverDeployment, &cr); err != nil {
		return ctrl.Result{}, err
	}

	if err := commonReconciler.MarkCrAsCompleted(ctx, commonCr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the QperfReconciler with the provided manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&perfv1alpha1.Qperf{}).
		Complete(r)
}
