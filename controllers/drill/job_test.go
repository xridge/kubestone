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

package drill

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ksapi "github.com/xridge/kubestone/api/v1alpha1"
	perfv1alpha1 "github.com/xridge/kubestone/api/v1alpha1"
)

var _ = Describe("drill job", func() {
	Describe("cr with cmd args", func() {
		var cr perfv1alpha1.Drill
		var job *batchv1.Job

		BeforeEach(func() {
			cr = perfv1alpha1.Drill{
				Spec: perfv1alpha1.DrillSpec{
					Image: perfv1alpha1.ImageSpec{
						Name:       "xridge/drill:test",
						PullPolicy: "Always",
						PullSecret: "the-pull-secret",
					},
					BenchmarksVolume: map[string]string{
						"the-benchmark.yml": "benchmark content",
						"included-file.yml": "included content",
					},
					BenchmarkFile: "/benchmarks/the-benchmark.yml",
					Options:       "--no-check-certificate --stats",
					PodConfig: ksapi.PodConfigurationSpec{
						PodLabels: map[string]string{"labels": "are", "still": "useful"},
						PodScheduling: ksapi.PodSchedulingSpec{
							Affinity: corev1.Affinity{
								NodeAffinity: &corev1.NodeAffinity{
									RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
										NodeSelectorTerms: []corev1.NodeSelectorTerm{
											{
												MatchExpressions: []corev1.NodeSelectorRequirement{
													{
														Key:      "mutated",
														Operator: corev1.NodeSelectorOperator(corev1.NodeSelectorOpIn),
														Values:   []string{"nano-virus"},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			}
			configMap := corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: "cm"},
			}
			job = NewJob(&cr, &configMap)
		})

		Context("with Image details specified", func() {
			It("should match on Image.Name", func() {
				Expect(job.Spec.Template.Spec.Containers[0].Image).To(
					Equal(cr.Spec.Image.Name))
			})
			It("should match on Image.PullPolicy", func() {
				Expect(job.Spec.Template.Spec.Containers[0].ImagePullPolicy).To(
					Equal(corev1.PullPolicy(cr.Spec.Image.PullPolicy)))
			})
			It("should match on Image.PullSecret", func() {
				Expect(job.Spec.Template.Spec.ImagePullSecrets[0].Name).To(
					Equal(cr.Spec.Image.PullSecret))
			})
		})

		Context("with command line args specified", func() {
			It("should have the same args", func() {
				Expect(job.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement("--no-check-certificate"))
				Expect(job.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement("--stats"))
				Expect(job.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement("--benchmark"))
				Expect(job.Spec.Template.Spec.Containers[0].Args).To(
					ContainElement(cr.Spec.BenchmarkFile))
			})
		})

		Context("with podAffinity specified", func() {
			It("should match with Affinity", func() {
				Expect(job.Spec.Template.Spec.Affinity).To(
					Equal(&cr.Spec.PodConfig.PodScheduling.Affinity))
			})
			It("should match with Tolerations", func() {
				Expect(job.Spec.Template.Spec.Tolerations).To(
					Equal(cr.Spec.PodConfig.PodScheduling.Tolerations))
			})
			It("should match with NodeSelector", func() {
				Expect(job.Spec.Template.Spec.NodeSelector).To(
					Equal(cr.Spec.PodConfig.PodScheduling.NodeSelector))
			})
			It("should match with NodeName", func() {
				Expect(job.Spec.Template.Spec.NodeName).To(
					Equal(cr.Spec.PodConfig.PodScheduling.NodeName))
			})
		})
	})
})