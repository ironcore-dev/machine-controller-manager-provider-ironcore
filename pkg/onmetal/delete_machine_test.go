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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/internal"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	testutils "github.com/onmetal/onmetal-api/utils/testing"
)

const MachineSecretError string = "machine codes error: code = [Internal] message = [failed to find namespace in machine secret ]"

var _ = Describe("DeleteMachine", func() {
	ctx := testutils.SetupContext()
	ns, providerSecret, drv := SetupTest(ctx)

	It("should create and delete a machine", func() {
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
		device := "oda"
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, machineKey, machine)
			Expect(err).ToNot(HaveOccurred())
			g.Expect(machine.Spec).To(Equal(computev1alpha1.MachineSpec{
				MachineClassRef:     corev1.LocalObjectReference{Name: "foo"},
				MachinePoolSelector: map[string]string{"foo": "bar"},
				MachinePoolRef:      &corev1.LocalObjectReference{Name: "foo"},
				Power:               computev1alpha1.PowerOn,
				Image:               "foo",
				ImagePullSecretRef:  &corev1.LocalObjectReference{Name: "foo"},
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "net-interface",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
							Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
								NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
									Spec: networkingv1alpha1.NetworkInterfaceSpec{
										NetworkRef: corev1.LocalObjectReference{Name: "network-ref1"},
										IPFamilies: []corev1.IPFamily{"IPv4"},
										IPs:        []networkingv1alpha1.IPSource{{Value: commonv1alpha1.MustParseNewIP("10.0.0.8")}},
										VirtualIP: &networkingv1alpha1.VirtualIPSource{
											Ephemeral: &networkingv1alpha1.EphemeralVirtualIPSource{
												VirtualIPTemplate: &networkingv1alpha1.VirtualIPTemplateSpec{
													Spec: networkingv1alpha1.VirtualIPSpec{
														Type:     networkingv1alpha1.VirtualIPTypePublic,
														IPFamily: corev1.IPv4Protocol,
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
				Volumes: []computev1alpha1.Volume{
					{
						Name:   "root-disk-1",
						Device: &device,
						VolumeSource: computev1alpha1.VolumeSource{
							VolumeRef: &corev1.LocalObjectReference{Name: "machine-0"},
						},
					},
				},
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

		By("ensuring that the machine can be deleted")
		response, err := (*drv).DeleteMachine(ctx, &driver.DeleteMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(internal.ProviderSpec),
			Secret:       providerSecret,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(response).To(Equal(&driver.DeleteMachineResponse{}))

		By("ensuring that the onmetal machine has been deleted")
		Eventually(func(g Gomega) {
			err := k8sClient.Get(ctx, machineKey, machine)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		}).Should(Succeed())

		By("ensuring that the onmetal machine ignition secret has been deleted")
		Expect(func(g Gomega) {
			err := k8sClient.Get(ctx, ignitionKey, ignition)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})

		By("failing when request is empty")
		Eventually(func(g Gomega) {
			_, err := (*drv).DeleteMachine(ctx, nil)
			g.Expect(err.Error()).To(Equal(EmptyRequestError))
		}).Should(Succeed())

		By("failing when provider is different")
		Eventually(func(g Gomega) {
			_, err := (*drv).DeleteMachine(ctx, &driver.DeleteMachineRequest{
				Machine:      newMachine(ns, "machine", -1, nil),
				MachineClass: newMachineClassWithDifferntProvider(internal.ProviderSpec),
				Secret:       providerSecret,
			})
			g.Expect(err.Error()).To(Equal(InvalidProvideNameError))
		}).Should(Succeed())

		machineSecretError := fmt.Sprintf("machine codes error: code = [Internal] message = [failed to find namespace in machine secret %s/%s]", providerSecret.Namespace, providerSecret.Name)
		By("failing when namespace not found in secret")
		delete(providerSecret.Data, "namespace")
		Eventually(func(g Gomega) {
			_, err := (*drv).DeleteMachine(ctx, &driver.DeleteMachineRequest{
				Machine:      newMachine(ns, "machine", -1, nil),
				MachineClass: newMachineClass(internal.ProviderSpec),
				Secret:       providerSecret,
			})
			g.Expect(err.Error()).To(Equal(machineSecretError))
		}).Should(Succeed())
	})
})
