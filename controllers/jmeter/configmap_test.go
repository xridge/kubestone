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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

var _ = Describe("jmeter configmaps", func() {
	Describe("cr with jmeter plan test", func() {
		var cr perfv1alpha1.JMeter
		var configMap *corev1.ConfigMap

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							"jmeter-sample-test.jmx": JMeterPlan,
						},
					},
				},
			}
			configMap, _ = NewPlanTestConfigMap(&cr)
		})

		Context("with jmx plan test specified", func() {
			It("should have them in the configmap", func() {
				Expect(configMap.Data).To(
					Equal(cr.Spec.Controller.PlanTest))
			})
		})

		Context("with jmx plan test name specified", func() {
			It("should have them in the configmap", func() {
				Expect(configMap.Data[cr.Spec.Controller.TestName]).To(
					Equal(cr.Spec.Controller.PlanTest[cr.Spec.Controller.TestName]))
			})
		})
	})

	Describe("cr with jmeter properties", func() {
		var cr perfv1alpha1.JMeter
		var configMap *corev1.ConfigMap

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						PropsName: "test.properties",
						Props: map[string]string{
							"test.properties": JMeterProperties,
						},
					},
				},
			}
			configMap, _ = NewPropertiesConfigMap(&cr)
		})

		Context("with jmx properties specified", func() {
			It("should have them in the configmap", func() {
				Expect(configMap.Data).To(
					Equal(cr.Spec.Controller.Props))
			})
		})

		Context("with jmx properties name specified", func() {
			It("should have them in the configmap", func() {
				Expect(configMap.Data[cr.Spec.Controller.PropsName]).To(
					Equal(cr.Spec.Controller.Props[cr.Spec.Controller.PropsName]))
			})
			It("should have the configMap", func() {
				Expect(configMap.Data).To(Equal(cr.Spec.Controller.Props))
			})
		})
	})
})
