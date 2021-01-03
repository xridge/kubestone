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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

// NewPlanTestConfigMap creates a new configmap containing the JMeter Plan test
// for the jmeter benchmark job
func NewPlanTestConfigMap(cr *perfv1alpha1.JMeter) (*corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-plan-tests", cr.Name),
			Namespace: cr.Namespace,
		},
		Data: cr.Spec.Controller.PlanTest,
	}

	return &configMap, nil
}

// NewPropertiesConfigMap creates a new configmap containing JMeter Properties
// for the jmeter benchmark job
func NewPropertiesConfigMap(cr *perfv1alpha1.JMeter) (*corev1.ConfigMap, error) {
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-properties", cr.Name),
			Namespace: cr.Namespace,
		},
		Data: cr.Spec.Controller.Props,
	}

	return &configMap, nil
}
