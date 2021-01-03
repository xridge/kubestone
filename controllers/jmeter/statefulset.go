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
	"errors"

	"github.com/firepear/qsplit"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatefulSet(cr *perfv1alpha1.JMeter) (*v1.StatefulSet, error) {
	if cr.Spec.Workers == nil {
		return nil, errors.New("Error creating StatefulSet, spec.workers isn't specified")
	}

	labels := map[string]string{
		"kubestone.xridge.io/app":       "jmeter",
		"kubestone.xridge.io/cr-name":   cr.Name,
		"perf.kubestone.xridge.io/role": "worker-node",
	}

	for key, value := range cr.Spec.Workers.Configuration.PodLabels {
		labels[key] = value
	}

	objectMeta := metav1.ObjectMeta{
		Name:      cr.Name,
		Labels:    labels,
		Namespace: cr.Namespace,
	}

	if cr.Spec.Workers.Configuration.Annotations != nil {
		if objectMeta.Annotations != nil {
			// If its already set add the values onto the existing map
			for k, v := range cr.Spec.Workers.Configuration.Annotations {
				objectMeta.Annotations[k] = v
			}
		} else {
			objectMeta.Annotations = cr.Spec.Workers.Configuration.Annotations
		}
	}

	if cr.Spec.Workers.Configuration.PodScheduling.Affinity == nil {
		cr.Spec.Workers.Configuration.PodScheduling.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  nil,
				PreferredDuringSchedulingIgnoredDuringExecution: nil,
			},
			PodAffinity: &corev1.PodAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  nil,
				PreferredDuringSchedulingIgnoredDuringExecution: nil,
			},
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution:  nil,
				PreferredDuringSchedulingIgnoredDuringExecution: nil,
			},
		}
	}

	args := qsplit.ToStrings([]byte(cr.Spec.Controller.Args))
	args = append(args, "-s", "-J", "server.rmi.ssl.disable=true")
	command := qsplit.ToStrings([]byte(cr.Spec.Controller.Command))

	statefulset := &v1.StatefulSet{
		ObjectMeta: objectMeta,
		Spec: v1.StatefulSetSpec{
			Replicas:            cr.Spec.Workers.Replicas,
			Selector:            metav1.SetAsLabelSelector(labels),
			ServiceName:         cr.Name,
			PodManagementPolicy: "Parallel",
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      labels,
					Annotations: objectMeta.Annotations,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						corev1.Container{
							Name:      "worker",
							Image:     cr.Spec.Workers.Image.Name,
							Args:      args,
							Command:   command,
							Resources: cr.Spec.Workers.Configuration.Resources,
							Ports: []corev1.ContainerPort{
								corev1.ContainerPort{
									Name:          "rmi",
									ContainerPort: 1099,
								},
							},
						},
					},
					ImagePullSecrets: []corev1.LocalObjectReference{
						{Name: cr.Spec.Workers.Image.PullSecret},
					},
					Affinity:     cr.Spec.Workers.Configuration.PodScheduling.Affinity,
					Tolerations:  cr.Spec.Workers.Configuration.PodScheduling.Tolerations,
					NodeSelector: cr.Spec.Workers.Configuration.PodScheduling.NodeSelector,
					NodeName:     cr.Spec.Workers.Configuration.PodScheduling.NodeName,
				},
			},
		},
	}

	return statefulset, nil
}
