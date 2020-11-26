package common

import (
	"context"

	"github.com/go-logr/logr"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/pkg/k8s"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Reconciler reconciles a k8s runtime object
type Reconciler struct {
	K8S k8s.Access
	Log logr.Logger
}

// CustomResource provides fields from concrete reconciler cr
type CustomResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Rto runtime.Object

	Status   perfv1alpha1.BenchmarkStatus `json:"status,omitempty"`
	NewJobFn func() *batchv1.Job
}

// IsCrReconciled checks if the cr is already reconciled with K8S
func (r *Reconciler) IsCrReconciled(ctx context.Context, cr CustomResource, req ctrl.Request) (bool, error) {
	if err := r.K8S.Client.Get(ctx, req.NamespacedName, cr.Rto); err != nil {
		return true, k8s.IgnoreNotFound(err)
	}

	if cr.Status.Completed {
		return true, nil
	}

	return false, nil
}

// MarkCrAsRunning marks the cr as running and ensures K8S knows about this
func (r *Reconciler) MarkCrAsRunning(ctx context.Context, cr CustomResource) error {
	cr.Status.Running = true
	if err := r.K8S.Client.Status().Update(ctx, cr.Rto); err != nil {
		return err
	}
	return nil
}

// MarkCrAsCompleted marks the cr as completed and ensures K8S knows about this
func (r *Reconciler) MarkCrAsCompleted(ctx context.Context, cr CustomResource) error {
	cr.Status.Running = false
	cr.Status.Completed = true
	if err := r.K8S.Client.Status().Update(ctx, cr.Rto); err != nil {
		return err
	}
	return nil
}

// HandleJob creates a job specified by the cr passes it to K8S for execution
// and checks if the job is finished
func (r *Reconciler) HandleJob(ctx context.Context, cr CustomResource, req ctrl.Request) (bool, error) {

	job := cr.NewJobFn()
	if err := r.K8S.CreateWithReference(ctx, job, &cr); err != nil {
		return false, err
	}

	// Check if finished
	jobFinished, err := r.K8S.IsJobFinished(types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Name,
	})
	if err != nil {
		return false, err
	}
	if !jobFinished {
		return false, nil
	}

	return true, nil
}
