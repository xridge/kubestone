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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

var _ = Describe("jmeter workers statefulset", func() {
	Describe("cr minimum parameter set", func() {
		var cr perfv1alpha1.JMeter
		var statefulSet *v1.StatefulSet
		replicas := int32(5)

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "jmeter-test",
					Namespace: "default",
				},
				Spec: perfv1alpha1.JMeterSpec{
					Workers: &perfv1alpha1.JMeterWorkers{
						Replicas: &replicas,
					},
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name:       "justb4/jmeter:5.3",
							PullPolicy: "Always",
						},
						Volume: perfv1alpha1.VolumeSpec{
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							"jmeter-sample-test.jmx": JMeterPlan,
						},
					},
				},
			}

			var err error

			statefulSet, err = NewStatefulSet(&cr)
			if err != nil {
				panic(err)
			}
		})
		Context("with replicas", func() {
			It("CR Validation should succeed", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
			It("should have the flag -s and -J", func() {
				args := statefulSet.Spec.Template.Spec.Containers[0].Args
				Expect(args).To(ContainElement("-s"))
				Expect(args).To(ContainElement("-J"))
				Expect(args).To(ContainElement("server.rmi.ssl.disable=true"))
			})
			It("statefulSet replicas should be", func() {
				Expect(*statefulSet.Spec.Replicas).To(Equal(replicas))
			})
		})
	})
	Describe("cr with invalid values", func() {
		Context("with -s flag", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Workers: &perfv1alpha1.JMeterWorkers{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						Args: "-s",
					},
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name:       "justb4/jmeter:5.3",
							PullPolicy: "Always",
						},
						Volume: perfv1alpha1.VolumeSpec{
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							"jmeter-sample-test.jmx": JMeterPlan,
						},
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})
	})
	Describe("cr without workers", func() {
		var cr perfv1alpha1.JMeter
		var statefulSet *v1.StatefulSet
		var err error

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name:       "justb4/jmeter:5.3",
							PullPolicy: "Always",
						},
						Volume: perfv1alpha1.VolumeSpec{
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							"jmeter-sample-test.jmx": JMeterPlan,
						},
					},
				},
			}
			statefulSet, err = NewStatefulSet(&cr)
		})
		Context("this context should never happen", func() {
			It("should fail with and error", func() {
				Expect(statefulSet).To(BeNil())
				Expect(err).To(MatchError(errors.New("Error creating StatefulSet, spec.workers isn't specified")))
			})
		})
	})
})
