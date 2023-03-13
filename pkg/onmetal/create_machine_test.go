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
	"encoding/json"
	"fmt"
	"net"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/onmetal/testing"
	commonv1alpha1 "github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
	testutils "github.com/onmetal/onmetal-api/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
)

var _ = Describe("CreateMachine", func() {
	ctx := testutils.SetupContext()
	ns, providerSecret, drv := SetupTest(ctx)

	It("should create a machine", func() {
		By("creating machine")
		machineName := "machine-0"
		Expect((*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
			Machine:      newMachine(ns, "machine", -1, nil),
			MachineClass: newMachineClass(v1alpha1.ProviderName, testing.SampleProviderSpec),
			Secret:       providerSecret,
		})).To(Equal(&driver.CreateMachineResponse{
			ProviderID: fmt.Sprintf("%s://%s/machine-%d", v1alpha1.ProviderName, ns.Name, 0),
			NodeName:   machineName,
		}))

		By("ensuring that the onmetal machine has been created")
		machine := &computev1alpha1.Machine{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      machineName,
			},
		}

		Eventually(Object(machine)).Should(SatisfyAll(
			HaveField("ObjectMeta.Labels", map[string]string{
				ShootNameLabelKey:      "my-shoot",
				ShootNamespaceLabelKey: "my-shoot-namespace",
			}),
			HaveField("Spec.MachineClassRef", corev1.LocalObjectReference{Name: "machine-class"}),
			HaveField("Spec.MachinePoolRef", &corev1.LocalObjectReference{Name: "az1"}),
			HaveField("Spec.Power", computev1alpha1.PowerOn),
			HaveField("Spec.NetworkInterfaces", ContainElement(computev1alpha1.NetworkInterface{
				Name: "primary",
				NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
					Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
						NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
							ObjectMeta: metav1.ObjectMeta{
								Labels: map[string]string{
									ShootNameLabelKey:      "my-shoot",
									ShootNamespaceLabelKey: "my-shoot-namespace",
								},
							},
							Spec: networkingv1alpha1.NetworkInterfaceSpec{
								NetworkRef: corev1.LocalObjectReference{
									Name: "my-network",
								},
								IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
								IPs: []networkingv1alpha1.IPSource{
									{
										Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
											PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
												Spec: ipamv1alpha1.PrefixSpec{
													IPFamily:     corev1.IPv4Protocol,
													PrefixLength: 32,
													ParentRef: &corev1.LocalObjectReference{
														Name: "my-prefix",
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
			})),
			HaveField("Spec.Volumes", ContainElement(computev1alpha1.Volume{
				Name:   "primary",
				Device: pointer.String("oda"),
				VolumeSource: computev1alpha1.VolumeSource{
					Ephemeral: &computev1alpha1.EphemeralVolumeSource{
						VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
							Spec: storagev1alpha1.VolumeSpec{
								VolumeClassRef: &corev1.LocalObjectReference{
									Name: "foo",
								},
								Resources: corev1alpha1.ResourceList{
									corev1alpha1.ResourceStorage: resource.MustParse("10Gi"),
								},
								Image: "my-image",
							},
						},
					},
				},
			})),
			HaveField("Spec.IgnitionRef", &commonv1alpha1.SecretKeySelector{
				Name: fmt.Sprintf("%s-ignition", machineName),
				Key:  defaultIgnitionKey,
			}),
		))

		By("ensuring that the ignition secret has been created")
		ignition := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns.Name,
				Name:      fmt.Sprintf("%s-ignition", machineName),
			},
		}

		ignitionData, err := json.Marshal(testing.SampleIgnition)
		Expect(err).NotTo(HaveOccurred())
		Eventually(Object(ignition)).Should(SatisfyAll(
			HaveField("Data", HaveKeyWithValue("ignition.json", MatchJSON(ignitionData))),
		))

		By("failing if the machine request is empty")
		Eventually(func(g Gomega) {
			_, err := (*drv).CreateMachine(ctx, nil)
			g.Expect(err.Error()).To(ContainSubstring("received empty request"))
		}).Should(Succeed())

		By("failing if the wrong provider is set")
		Eventually(func(g Gomega) {
			_, err := (*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
				Machine:      newMachine(ns, "machine", -1, nil),
				MachineClass: newMachineClass("foo", testing.SampleProviderSpec),
				Secret:       providerSecret,
			})
			g.Expect(err.Error()).To(ContainSubstring("not supported by the driver"))
		}).Should(Succeed())

		By("failing if the invalid IP is set for dnsServers in provider")
		Eventually(func(g Gomega) {
			SampleProviderSpec := testing.Copy(testing.SampleProviderSpec)
			SampleProviderSpec["dnsServers"] = append(testing.SampleProviderSpec["dnsServers"].([]net.IP), net.ParseIP("invalid-dns-ips"))
			_, err := (*drv).CreateMachine(ctx, &driver.CreateMachineRequest{
				Machine:      newMachine(ns, "machine", -1, nil),
				MachineClass: newMachineClass(v1alpha1.ProviderName, SampleProviderSpec),
				Secret:       providerSecret,
			})
			g.Expect(err.Error()).To(ContainSubstring("invalid IP"))
		}).Should(Succeed())
	})
})
