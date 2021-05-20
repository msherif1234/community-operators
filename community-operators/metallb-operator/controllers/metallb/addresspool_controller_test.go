/*
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

package controllers

import (
	"context"
	"time"

	metallbv1beta1 "community-operators/community-operators/metallb-operator/apis/metallb/v1beta1"
	. "github.com/onsi/ginkgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("AddressPool Controller", func() {

	const timeout = time.Second * 30
	const interval = time.Second * 1

	BeforeEach(func() {
		// Add any setup steps that needs to be executed before each test
	})

	AfterEach(func() {
		// Add any teardown steps that needs to be executed after each test
	})

	// Add Tests for OpenAPI validation (or additonal CRD features) specified in
	// your API definition.
	// Avoid adding tests for vanilla CRUD operations because they would
	// test Kubernetes API server, which isn't the goal here.
	Context("Creating AddressPool", func() {
		It("Should create successfully", func() {
			By("By creating a new Job")
			ctx, addrpool := context.Background(), &metallbv1beta1.AddressPool{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "metallb.metallb.io/v1beta1",
					Kind:       "AddressPool",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-addresspool",
					Namespace: "test-addresspool-namesapce",
				},
				Spec: metallbv1beta1.AddressPoolSpec{
					AddressPool: metallbv1beta1.AddressPoolType{
						Name:     "test1",
						Protocol: "layer2",
						Addresses: []string{
							"1.1.1.1",
							"1.1.1.100",
						},
					},
				},
			}

			// Create
			k8sClient.Create(ctx, addrpool)
		})
	})
})
