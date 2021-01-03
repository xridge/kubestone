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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

func indexOf(element string, data []string) int {
	for k, v := range data {
		if element == v {
			return k
		}
	}
	return -1
}

var _ = Describe("jmeter job", func() {
	Describe("cr minimum parameter set", func() {
		var cr perfv1alpha1.JMeter
		var configMap *corev1.ConfigMap
		var job *batchv1.Job

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
			var err error

			configMap, err = NewPlanTestConfigMap(&cr)
			if err != nil {
				panic(err)
			}

			job = NewJob(&cr, configMap, nil)
		})

		Context("with image, volume and plan test values", func() {
			It("should have the flag -t with path to the test", func() {
				Expect(len(job.Spec.Template.Spec.Containers[0].Args)).To(Equal(9))
				args := job.Spec.Template.Spec.Containers[0].Args
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement(fmt.Sprintf("/jmeter-plan-tests/%s", cr.Spec.Controller.TestName)))
			})
			It("should have the configMap with the data specified at CR", func() {
				Expect(configMap.Data).To(Equal(cr.Spec.Controller.PlanTest))
			})
			It("shouldn't have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).To(BeEmpty())
			})
			It("CR Validation should succeed", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("cr with invalid values", func() {
		Context("when non PlanTest is defined", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						TestName: "test.jmx",
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})

		Context("when an empty PlanTest defined", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						PlanTest: map[string]string{
							"test.jmx": "",
						},
						TestName: "test.jmx",
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with non TestName defined", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
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

		Context("with -t flag", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						TestName: "test.jmx",
						Args:     "-t /path/to/test.jmx",
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with -o flag", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						TestName: "test.jmx",
						Args:     "-o /path/to/report",
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with -s flag", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						TestName: "test.jmx",
						Args:     "-s",
					},
				},
			}

			It("CR Validation should fail", func() {
				valid, err := IsCrValid(&invalidCr)
				Expect(valid).To(BeFalse())
				Expect(err).NotTo(BeNil())
			})
		})

		Context("with an invalid volume", func() {
			invalidCr := perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name: "justb4/jmeter:5.3",
						},
						Volume: perfv1alpha1.VolumeSpec{
							PersistentVolumeClaimSpec: &corev1.PersistentVolumeClaimSpec{
								Resources: corev1.ResourceRequirements{
									Requests: corev1.ResourceList{
										"storage": resource.Quantity{
											Format: "1Gi",
										},
									},
								},
							},
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "Claim-Name",
								},
							},
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

	Describe("cr extra args", func() {
		var cr perfv1alpha1.JMeter
		var job *batchv1.Job

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name:       "justb4/jmeter:5.3",
							PullPolicy: "Always",
						},
						Args:     "-L jmeter.util=DEBUG",
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							// Declared at suit_test.go
							"jmeter-sample-test.jmx": JMeterPlan,
						},
					},
				},
			}
			configMap, err := NewPlanTestConfigMap(&cr)
			if err != nil {
				panic(err)
			}

			job = NewJob(&cr, configMap, nil)
		})

		Context("with extra command line args specified", func() {
			It("should have the flags -t with a path to the test, and a -L flag with the log level", func() {
				args := job.Spec.Template.Spec.Containers[0].Args
				Expect(len(args)).To(Equal(11))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement(fmt.Sprintf("/jmeter-plan-tests/%s", cr.Spec.Controller.TestName)))
				Expect(args).To(ContainElement("-L"))
				Expect(args).To(ContainElement("jmeter.util=DEBUG"))
			})
			It("shouldn't have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).To(BeEmpty())
			})
			It("shouldn't have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).To(BeEmpty())
			})
			It("CR Validation should succeed", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("cr with properties specified", func() {
		var cr perfv1alpha1.JMeter
		var configMapPlanTest, configMapProperties *corev1.ConfigMap
		var job *batchv1.Job

		BeforeEach(func() {
			cr = perfv1alpha1.JMeter{
				Spec: perfv1alpha1.JMeterSpec{
					Controller: &perfv1alpha1.JMeterController{
						Image: perfv1alpha1.ImageSpec{
							Name:       "justb4/jmeter:5.3",
							PullPolicy: "Always",
						},
						TestName: "jmeter-sample-test.jmx",
						PlanTest: map[string]string{
							// Declared at suit_test.go
							"jmeter-sample-test.jmx": JMeterPlan,
						},
						Volume: perfv1alpha1.VolumeSpec{
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						PropsName: "test.properties",
						Props: map[string]string{
							// Declared at suit_test.go
							"test.properties": JMeterProperties,
						},
					},
				},
			}
			var err error

			configMapPlanTest, err = NewPlanTestConfigMap(&cr)
			if err != nil {
				panic(err)
			}

			configMapProperties, err = NewPropertiesConfigMap(&cr)
			if err != nil {
				panic(err)
			}

			job = NewJob(&cr, configMapPlanTest, configMapProperties)
		})
		Context("when props and propsName are specified", func() {
			It("should have the flag -t with path to the test", func() {
				args := job.Spec.Template.Spec.Containers[0].Args
				Expect(len(args)).To(Equal(11))
				Expect(args).To(ContainElement("-t"))
				Expect(args).To(ContainElement(fmt.Sprintf("/jmeter-plan-tests/%s", cr.Spec.Controller.TestName)))
			})
			It("should have the flag -p with path to the properties", func() {
				args := job.Spec.Template.Spec.Containers[0].Args
				Expect(len(args)).To(Equal(11))
				Expect(args).To(ContainElement("-p"))
				Expect(args).To(ContainElement(fmt.Sprintf("/jmeter-properties/%s", cr.Spec.Controller.PropsName)))
			})
			It("shouldn't have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).To(BeEmpty())
			})
			It("CR Validation shluld pass", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("cr with workers", func() {
		var cr perfv1alpha1.JMeter
		var configMapPlanTest, configMapProperties *corev1.ConfigMap
		var job *batchv1.Job
		replicas := int32(5)

		Context("when workers and workers.replicas are specified", func() {
			BeforeEach(func() {
				cr = perfv1alpha1.JMeter{
					ObjectMeta: v1.ObjectMeta{
						Name:      "jmeter-test",
						Namespace: "default",
					},
					Spec: perfv1alpha1.JMeterSpec{
						Workers: &perfv1alpha1.JMeterWorkers{
							Replicas: &replicas,
							Image: perfv1alpha1.ImageSpec{
								Name:       "justb4/jmeter:5.3",
								PullPolicy: "Always",
							},
						},
						Controller: &perfv1alpha1.JMeterController{
							Image: perfv1alpha1.ImageSpec{
								Name:       "justb4/jmeter:5.3",
								PullPolicy: "Always",
							},
							TestName: "jmeter-sample-test.jmx",
							PlanTest: map[string]string{
								// Declared at suit_test.go
								"jmeter-sample-test.jmx": JMeterPlan,
							},
							Volume: perfv1alpha1.VolumeSpec{
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				}
				var err error

				configMapPlanTest, err = NewPlanTestConfigMap(&cr)
				if err != nil {
					panic(err)
				}

				job = NewJob(&cr, configMapPlanTest, configMapProperties)
			})
			It("shoud have the flag -R and the svc names", func() {
				args := job.Spec.Template.Spec.Containers[0].Args
				index := indexOf("-R", args)
				serverStr := args[index+1]

				Expect(len(args)).To(Equal(13))
				Expect(args).To(ContainElement("-R"))
				Expect(serverStr).To(MatchRegexp("\\.cluster\\.local"))
			})
			It("shoud have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).ToNot(BeEmpty())
			})
			It("CR Validation shluld pass", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
		Context("when spec.controller.clusterDomain is specified", func() {
			BeforeEach(func() {
				cr = perfv1alpha1.JMeter{
					ObjectMeta: v1.ObjectMeta{
						Name:      "jmeter-test",
						Namespace: "default",
					},
					Spec: perfv1alpha1.JMeterSpec{
						Workers: &perfv1alpha1.JMeterWorkers{
							Replicas: &replicas,
							Image: perfv1alpha1.ImageSpec{
								Name:       "justb4/jmeter:5.3",
								PullPolicy: "Always",
							},
						},
						Controller: &perfv1alpha1.JMeterController{
							Image: perfv1alpha1.ImageSpec{
								Name:       "justb4/jmeter:5.3",
								PullPolicy: "Always",
							},
							ClusterDomain: "my.domain",
							TestName:      "jmeter-sample-test.jmx",
							PlanTest: map[string]string{
								// Declared at suit_test.go
								"jmeter-sample-test.jmx": JMeterPlan,
							},
							Volume: perfv1alpha1.VolumeSpec{
								VolumeSource: corev1.VolumeSource{
									EmptyDir: &corev1.EmptyDirVolumeSource{},
								},
							},
						},
					},
				}
				var err error

				configMapPlanTest, err = NewPlanTestConfigMap(&cr)
				if err != nil {
					panic(err)
				}

				job = NewJob(&cr, configMapPlanTest, configMapProperties)
			})
			It("shoud have the flag -R and the svc names", func() {
				replicas := int(*cr.Spec.Workers.Replicas)
				args := job.Spec.Template.Spec.Containers[0].Args
				index := indexOf("-R", args)
				serverStr := args[index+1]
				servers := strings.Split(serverStr, ",")

				Expect(len(args)).To(Equal(13))
				Expect(args).To(ContainElement("-R"))
				Expect(index).NotTo(Equal(-1))
				Expect(len(servers)).To(Equal(replicas))
			})
			It("shoud have servers with spec.controller.clusterDomain", func() {
				args := job.Spec.Template.Spec.Containers[0].Args
				index := indexOf("-R", args)
				serverStr := args[index+1]

				Expect(serverStr).To(MatchRegexp("\\.my\\.domain"))
			})
			It("shoud have an initContainer", func() {
				initContainers := job.Spec.Template.Spec.InitContainers
				Expect(initContainers).ToNot(BeEmpty())
			})
			It("CR Validation shluld pass", func() {
				valid, err := IsCrValid(&cr)
				Expect(valid).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
	})
})
