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

package e2e

import (
	"io/ioutil"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/xridge/kubestone/api/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	fioCrBaseDir = samplesDir + "/fio"
)

var _ = Describe("end to end test", func() {
	fioCrDirs := []string{fioCrBaseDir + "/base/"}
	fioOverlayContents, err := ioutil.ReadDir(fioCrBaseDir + "/overlays")
	if err != nil {
		Fail("Didn't find any fio CRs under " + fioCrBaseDir)
	}
	for _, fioOverlayContent := range fioOverlayContents {
		if fioOverlayContent.IsDir() {
			fioCrDirs = append(fioCrDirs, fioCrBaseDir+"/overlays/"+fioOverlayContent.Name())
		}
	}

	Describe("creating fio job from multiple CRs", func() {
		for _, fioCrDir := range fioCrDirs {
			splits := strings.Split(fioCrDir, "/")
			dirName := splits[len(splits)-1]
			crName := "fio-" + strings.ReplaceAll(dirName, "_", "-")

			Context("when creating from cr", func() {
				It("should create fio-sample cr", func() {
					_, _, err := run(`bash -c "` +
						"kustomize build " + fioCrDir + " | " +
						"sed 's/name: fio-sample/name: " + crName + "/' | " +
						"kubectl create -n " + e2eNamespaceFio + ` -f -"`)
					Expect(err).To(BeNil())
				})
			})

			Context("the created job", func() {
				It("should finish in a pre-defined time", func() {
					timeout := 90
					cr := &v1alpha1.Fio{}
					// TODO: find the respective objects via the CR owner reference
					namespacedName := types.NamespacedName{
						Namespace: e2eNamespaceFio,
						Name:      crName,
					}
					Eventually(func() bool {
						if err := client.Get(ctx, namespacedName, cr); err != nil {
							Fail("Unable to get fio CR: " + err.Error())
						}
						return !cr.Status.Running && cr.Status.Completed
					}, timeout).Should(BeTrue())
				})
				It("Should leave a successful job", func() {
					job := &batchv1.Job{}
					namespacedName := types.NamespacedName{
						Namespace: e2eNamespaceFio,
						Name:      crName,
					}
					Expect(client.Get(ctx, namespacedName, job)).To(Succeed())
					Expect(job.Status.Succeeded).To(Equal(int32(1)))
				})
			})
		}
	})
})
