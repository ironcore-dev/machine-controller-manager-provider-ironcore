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
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"github.com/onmetal/onmetal-api/testutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("CreateMachine", func() {
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

		By("expecting that the onmetal machine is present")
		machineKey := types.NamespacedName{Namespace: ns.Name, Name: "machine-0"}
		machine := &computev1alpha1.Machine{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, machineKey, machine)
			Expect(err).ToNot(HaveOccurred())
			g.Expect(machine.Spec).To(Equal(computev1alpha1.MachineSpec{
				MachineClassRef:     corev1.LocalObjectReference{Name: "foo"},
				MachinePoolSelector: map[string]string{"foo": "bar"},
				MachinePoolRef:      &corev1.LocalObjectReference{Name: "foo"},
				Image:               "foo",
				ImagePullSecretRef:  &corev1.LocalObjectReference{Name: "foo"},
				NetworkInterfaces:   nil,
				Volumes:             nil,
				IgnitionRef: &commonv1alpha1.SecretKeySelector{
					Name: "machine-0-ignition",
					Key:  defaultIgnitionKey,
				},
			}))
		}).Should(Succeed())

		By("ensuring that the ignition secret has been created")
		ignitionKey := types.NamespacedName{Namespace: ns.Name, Name: "machine-0-ignition"}
		ignition := &corev1.Secret{}
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, ignitionKey, ignition)
			Expect(err).NotTo(HaveOccurred())
			// TODO: validate ignition content
		}).Should(Succeed())
	})
})
