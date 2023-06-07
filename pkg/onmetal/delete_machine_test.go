// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package onmetal

import (
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/onmetal/testing"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("DeleteMachine", func() {
	ns, providerSecret, drv := SetupTest()

	It("should create and delete a machine", func(ctx SpecContext) {
		By("creating an onmetal machine")
		Expect((*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0),
			NodeName:   "machine-0",
		}))

		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "machine-0",
			},
		}

		ignition := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      "machine-0-ignition",
			},
		}

		By("ensuring that the machine can be deleted")
		response, err := (*drv).DeleteMachine(ctx, &driver.DeleteMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&driver.DeleteMachineResponse{}))

		By("waiting for the machine to be gone")
		Eventually(Get(machine)).Should(Satisfy(apierrors.IsNotFound))

		By("waiting for the machine to be gone")
		Eventually(Get(ignition)).Should(Satisfy(apierrors.IsNotFound))
	})
})
