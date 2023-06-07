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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListMachines", func() {
	ns, providerSecret, drv := SetupTest()

	It("should fail if no provider has been set", func(ctx SpecContext) {
		By("ensuring an error if no provider has been set")
		_, err := (*drv).ListMachines(ctx, &driver.ListMachinesRequest{
			MachineClass: newMachineClass("", testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).To(HaveOccurred())
	})

	It("should list no machines if none have been created", func(ctx SpecContext) {
		By("ensuring the list response contains no machines")
		listMachineResponse, err := (*drv).ListMachines(ctx, &driver.ListMachinesRequest{
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(listMachineResponse.MachineList).To(Equal(map[string]string{}))
	})

	It("should list a single machine if one has been created", func(ctx SpecContext) {
		By("creating a machine")
		craeteMachineResponse, err := (*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(craeteMachineResponse).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0),
			NodeName:   "machine-0",
		}))

		By("ensuring the list response contains the correct machine")
		listMachineResponse, err := (*drv).ListMachines(ctx, &driver.ListMachinesRequest{
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(listMachineResponse.MachineList).To(Equal(
			map[string]string{fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0): "machine-0"},
		))

		By("ensuring the cleanup of the machine")
		DeferCleanup((*drv).DeleteMachine, &driver.DeleteMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
	})

	It("should list two machines if two have been created", func(ctx SpecContext) {
		By("creating the first machine")
		craeteMachineResponse, err := (*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", 0, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(craeteMachineResponse).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0),
			NodeName:   "machine-0",
		}))

		By("creating the second machine")
		craeteMachineResponse, err = (*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", 1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(craeteMachineResponse).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 1),
			NodeName:   "machine-1",
		}))

		By("ensuring the machine status contains 2 machines")
		listMachinesResponse, err := (*drv).ListMachines(ctx, &driver.ListMachinesRequest{
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(listMachinesResponse.MachineList).To(Equal(map[string]string{
			fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0): "machine-0",
			fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 1): "machine-1",
		}))

		By("ensuring the cleanup of the first machine")
		DeferCleanup((*drv).DeleteMachine, &driver.DeleteMachineRequest{
			Machine:      newMachine(ns, "machine", 0, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})

		By("ensuring the cleanup of the second machine")
		DeferCleanup((*drv).DeleteMachine, &driver.DeleteMachineRequest{
			Machine:      newMachine(ns, "machine", 1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})
	})
})
