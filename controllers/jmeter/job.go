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
	"fmt"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/firepear/qsplit"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
	"github.com/xridge/kubestone/pkg/k8s"
)

const (
	testsDir      = "/jmeter-plan-tests"
	propertiesDir = "/jmeter-properties"
	reportsDir    = "/jmeter-reports"
)

// NewJob creates a new jmeter job
func NewJob(cr *perfv1alpha1.JMeter, planTestConfigMap, propertiesConfigMap *corev1.ConfigMap) *batchv1.Job {
	var initContainers []corev1.Container
	objectMeta := metav1.ObjectMeta{
		Name:      cr.Name,
		Namespace: cr.Namespace,
	}

	volumes := generateVolumes(cr, planTestConfigMap, propertiesConfigMap)
	volumeMounts := generateVolumeMounts(cr, planTestConfigMap, propertiesConfigMap)

	args := qsplit.ToStrings([]byte(cr.Spec.Controller.Args))
	args = append(args,
		"-t",
		fmt.Sprintf("%s/%s", testsDir, cr.Spec.Controller.TestName),
		"-e",
		"-j",
		fmt.Sprintf("%s/jmeter.log", reportsDir),
		"-l",
		fmt.Sprintf("%s/test-plan.jtl", reportsDir),
		"-o",
		fmt.Sprintf("%s/report", reportsDir),
	)

	if cr.Spec.Workers != nil {
		clusterDomain := "cluster.local"
		if cr.Spec.Controller.ClusterDomain != "" {
			clusterDomain = cr.Spec.Controller.ClusterDomain
		}

		servers := []string{}
		for i := 0; i < int(*cr.Spec.Workers.Replicas); i++ {
			servers = append(servers,
				fmt.Sprintf("%s-%d.%s.%s.svc.%s",
					cr.Name,
					i,
					cr.Name,
					cr.Namespace,
					clusterDomain,
				),
			)
		}

		if len(servers) != 0 {
			args = append(args,
				"-R",
				strings.Join(servers, ","),
				"-J",
				"server.rmi.ssl.disable=true",
			)
		}

		initContainers = []corev1.Container{
			corev1.Container{
				Name:    "check-workers",
				Image:   "alpine:3",
				Command: []string{"/bin/sh"},
				Args: []string{
					"-c",
					fmt.Sprintf("for worker in %s; do until nc -w 3 -z $worker 1099; do echo Waiting for $worker; sleep 2; done; done; echo All up!", strings.Join(servers, " ")),
				},
			},
		}
	}

	if propertiesConfigMap != nil {
		args = append(args, "-p", fmt.Sprintf("%s/%s", propertiesDir, cr.Spec.Controller.PropsName))
	}

	job := k8s.NewPerfJob(objectMeta, "jmeter", cr.Spec.Controller.Image, cr.Spec.Controller.Configuration)
	job.Spec.Template.Spec.Volumes = volumes
	job.Spec.Template.Spec.Containers[0].Args = args
	job.Spec.Template.Spec.Containers[0].Command = qsplit.ToStrings([]byte(cr.Spec.Controller.Command))
	job.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts
	job.Spec.Template.Spec.InitContainers = initContainers
	return job
}

func generateVolumes(cr *perfv1alpha1.JMeter, planTestConfigMap, propertiesConfigMap *corev1.ConfigMap) []corev1.Volume {
	volumes := []corev1.Volume{
		corev1.Volume{
			Name: "plans",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: planTestConfigMap.Name,
					},
				},
			},
		},
		corev1.Volume{
			Name:         "reports",
			VolumeSource: cr.Spec.Controller.Volume.VolumeSource,
		},
	}

	if propertiesConfigMap != nil {
		propertiesVolume := corev1.Volume{
			Name: "properties",
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: propertiesConfigMap.Name,
					},
				},
			},
		}

		volumes = append(volumes, propertiesVolume)
	}

	return volumes
}

func generateVolumeMounts(cr *perfv1alpha1.JMeter, planTestConfigMap, propertiesConfigMap *corev1.ConfigMap) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		corev1.VolumeMount{
			Name:      "plans",
			MountPath: testsDir,
		},
		corev1.VolumeMount{
			Name:      "reports",
			MountPath: reportsDir,
		},
	}

	if propertiesConfigMap != nil {
		propertiesVolumeMount := corev1.VolumeMount{
			Name:      "properties",
			MountPath: propertiesDir,
		}
		volumeMounts = append(volumeMounts, propertiesVolumeMount)
	}

	return volumeMounts
}
