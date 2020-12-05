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

package v1alpha1

import (
	"errors"

	corev1 "k8s.io/api/core/v1"
)

// PullPolicy controls how the docker images are downloaded
// Defaults to Always if :latest tag is specified, or IfNotPresent otherwise.
// +kubebuilder:validation:Enum=Always;Never;IfNotPresent
type PullPolicy string

// ImageSpec defines parameters for docker image executed on Kubernetes
type ImageSpec struct {
	// Name is the Docker Image location including the tag
	Name string `json:"name"`

	// +optional
	PullPolicy PullPolicy `json:"pullPolicy,omitempty"`

	// PullSecret is an optional list of references to secrets
	// in the same namespace to use for pulling any of the images
	// +optional
	PullSecret string `json:"pullSecret,omitempty"`
}

// PodSchedulingSpec encapsulates the scheduling related
// fields of a Kubernetes Pod
type PodSchedulingSpec struct {
	// Affinity is a group of affinity scheduling rules.
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`

	// A node selector represents the union of the results of
	// one or more label queries over a set of nodes; that is,
	// it represents the OR of the selectors represented by the
	// node selector terms.
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	// +optional
	NodeName string `json:"nodeName,omitempty"`
}

// VolumeSpec contains the Volume Definition used for the benchmarks.
// It can point to an EmptyDir, HostPath, already existing PVC or PVC
// to be created benchmark time.
type VolumeSpec struct {
	// VolumeSource represents the source of the volume, e.g. EmptyDir,
	// HostPath, Ceph, PersistentVolumeClaim, etc.
	// PersistentVolumeClaim.claimName can be set to point to an already
	// existing PVC or could be set to 'GENERATED'. When set to 'GENERATED'
	// The PVC will be created based on the PersistentVolumeClaimSpec provided
	// to the VolumeSpec.
	VolumeSource corev1.VolumeSource `json:"volumeSource"`
	VolumeDevice corev1.VolumeDevice `json:"volumeDevice"`
	// PersistentVolumeClaimSpec describes the persistent volume claim that will be
	// created and used by the pod. If specified, the VolumeSource.PersistentVolumeClaim's
	// claimName must be set to 'GENERATED'
	// +optional
	PersistentVolumeClaimSpec *corev1.PersistentVolumeClaimSpec `json:"persistentVolumeClaimSpec,omitempty"`
}

// GeneratedPVC is the pre-defined name to be used as ClaimName
// when the PVC is created on the fly for the benchmark.
const GeneratedPVC = "GENERATED"

// Validate method validates that the provided VolumeSpec meets the
// requirements:
// If PersistentVolumeClaimSpec is provided, then the VolumeSource's
// PersistentVolumClaim's ClaimName should be set to GeneratedPVC
func (v *VolumeSpec) Validate() (ok bool, err error) {
	if v.PersistentVolumeClaimSpec != nil {
		if v.VolumeSource.PersistentVolumeClaim != nil &&
			v.VolumeSource.PersistentVolumeClaim.ClaimName != GeneratedPVC {
			return false, errors.New("If PersistentVolumeClaimSpec is defined, " +
				"VolumeSource.PersistentVolumeClaim.ClaimName must be set to " + GeneratedPVC)
		}
	}
	return true, nil
}

// PodConfigurationSpec contains the configuration for the benchmark pods
type PodConfigurationSpec struct {

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// PodLabels are added to the pod as labels.
	// +optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// PodScheduling contains options to determine which
	// node the pod should be scheduled on
	// +optional
	PodScheduling PodSchedulingSpec `json:"podScheduling,omitempty"`

	// Resources required by the benchmark pod container
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}
