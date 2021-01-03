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

package jmeter

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/pkg/k8s"
)

// Reconciler reconciles a JMeter object
type Reconciler struct {
	K8S k8s.Access
	Log logr.Logger
}

// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=jmeters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=jmeters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=perf.kubestone.xridge.io,resources=jmeters/finalizers,verbs=update

// Reconcile creates jmeter job(s) based on the custom resource(s)
func (r *Reconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	var cr perfv1alpha1.JMeter
	if err := r.K8S.Client.Get(ctx, req.NamespacedName, &cr); err != nil {
		return ctrl.Result{}, k8s.IgnoreNotFound(err)
	}

	if cr.Status.Completed {
		return ctrl.Result{}, nil
	}

	// Validate on first entry
	if !cr.Status.Completed && !cr.Status.Running {
		if valid, err := IsCrValid(&cr); !valid {
			_ = r.K8S.RecordEventf(&cr, corev1.EventTypeWarning, k8s.CreateFailed,
				"CR validation failed: %v", err)
			cr.Status.Valid = false
			if err := r.K8S.Client.Update(ctx, &cr); err != nil {
				return ctrl.Result{}, err
			}

			// Do not requeue invalid CRs
			return ctrl.Result{}, nil
		}
	}

	cr.Status.Valid = true
	cr.Status.Running = true
	if err := r.K8S.Client.Update(ctx, &cr); err != nil {
		return ctrl.Result{}, err
	}

	if cr.Spec.Controller.Volume.PersistentVolumeClaimSpec != nil {
		pvc := k8s.NewPersistentVolumeClaim(*cr.Spec.Controller.Volume.PersistentVolumeClaimSpec,
			cr.Name, cr.Namespace)
		if err := r.K8S.CreateWithReference(ctx, pvc, &cr); err != nil {
			return ctrl.Result{}, err
		}
		// Change ClaimName (from GENERATED) to the PVC was created
		cr.Spec.Controller.Volume.VolumeSource.PersistentVolumeClaim.ClaimName = cr.Name
	}

	// Create the planTestConfigMap
	planTestConfigMap, err := NewPlanTestConfigMap(&cr)

	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.K8S.CreateWithReference(ctx, planTestConfigMap, &cr); err != nil {
		return ctrl.Result{}, err
	}

	// Create the propertiesConfigMap
	var propertiesConfigMap *corev1.ConfigMap
	if cr.Spec.Controller.Props != nil {
		propertiesConfigMap, err = NewPropertiesConfigMap(&cr)

		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.K8S.CreateWithReference(ctx, propertiesConfigMap, &cr); err != nil {
			return ctrl.Result{}, err
		}
	}

	if cr.Spec.Workers != nil {
		statefulset, err := NewStatefulSet(&cr)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := r.K8S.CreateWithReference(ctx, statefulset, &cr); err != nil {
			return ctrl.Result{}, err
		}

		service := NewService(&cr, statefulset.Labels)
		if err := r.K8S.CreateWithReference(ctx, service, &cr); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Create the job
	job := NewJob(&cr, planTestConfigMap, propertiesConfigMap)
	if err := r.K8S.CreateWithReference(ctx, job, &cr); err != nil {
		return ctrl.Result{}, err
	}

	jobFinished, err := r.K8S.IsJobFinished(types.NamespacedName{
		Namespace: cr.Namespace,
		Name:      cr.Name,
	})
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
	cr.Status.Running = false
	cr.Status.Completed = true
	if err := r.K8S.Client.Update(ctx, &cr); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the Reconciler with the provided manager
func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&perfv1alpha1.JMeter{}).
		Complete(r)
}

// IsCrValid validates the given CR and raises error if semantic errors detected
// For jmeter it checks that the plan test is valid
func IsCrValid(cr *perfv1alpha1.JMeter) (valid bool, err error) {
	if len(cr.Spec.Controller.TestName) == 0 {
		return false, errors.New("You need to specify the TestName")
	}

	if strings.Contains(cr.Spec.Controller.Args, "-t") {
		return false, fmt.Errorf("You can't specify the flag '-t'")
	}

	if strings.Contains(cr.Spec.Controller.Args, "-o") {
		return false, fmt.Errorf("You can't specify the flag '-o'")
	}

	if strings.Contains(cr.Spec.Controller.Args, "-s") {
		return false, fmt.Errorf("You can't specify the flag '-s' on the controller spec")
	}

	if cr.Spec.Workers != nil && strings.Contains(cr.Spec.Workers.Args, "-s") {
		return false, fmt.Errorf("You can't specify the flag '-s' on the workers spec")
	}

	testName := cr.Spec.Controller.TestName
	planTest, ok := cr.Spec.Controller.PlanTest[testName]

	if !ok {
		return false, fmt.Errorf("The key '%s' is missing at spec.controller.planTest", testName)
	}

	if planTest == "" {
		return false, fmt.Errorf("The key '%s' is empty at spec.controller.planTest", testName)
	}

	if ok, err := cr.Spec.Controller.Volume.Validate(); !ok || err != nil {
		return false, fmt.Errorf("The volume spec is invalid: %s", err)
	}

	return true, nil
}
