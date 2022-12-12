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
	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/internal"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ListMachines", func() {
	ctx := testutils.SetupContext()
	ns, providerSecret, drv := SetupTest(ctx)

	It("should create a machine", func() {
		By("creating machine")
		Expect((*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(internal.ProviderSpec),
			Secret:       providerSecret,
		})).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0),
			NodeName:   "machine-0",
		}))

		By("ensuring the machine status is set")
		Expect((*drv).ListMachines(ctx, &driver.ListMachinesRequest{
			MachineClass: newMachineClass(internal.ProviderSpec),
			Secret:       providerSecret,
		})).To(Equal(&driver.ListMachinesResponse{
			MachineList: map[string]string{fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0): "machine-0"},
		}))
	})
})
