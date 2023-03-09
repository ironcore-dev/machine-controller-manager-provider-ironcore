/*
 * Copyright (c) 2022 by the OnMetal authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package validation

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"

	"net/netip"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

var fldPath *field.Path

var _ = Describe("Machine", func() {
	invalidIP := netip.Addr{}

	DescribeTable("ValidateProviderSpecAndSecret",
		func(spec *v1alpha1.ProviderSpec, secret *corev1.Secret, fldPath *field.Path, match types.GomegaMatcher) {
			errList := ValidateProviderSpecAndSecret(spec, secret, fldPath)
			Expect(errList).To(match)
		},
		Entry("no secret",
			&v1alpha1.ProviderSpec{
				RootDisk: &v1alpha1.RootDisk{},
			},
			nil,
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child(""), "corev1 secret required"))),
		),
		Entry("no userData in secret",
			&v1alpha1.ProviderSpec{
				RootDisk: &v1alpha1.RootDisk{},
			},
			&corev1.Secret{
				Data: map[string][]byte{
					"userData": nil,
				},
			},
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child("userData"), "corev1 secret userData required"))),
		),
		Entry("no image",
			&v1alpha1.ProviderSpec{
				Image:    "",
				RootDisk: &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child("spec.image"), "provider's spec image required"))),
		),
		Entry("no volumeclass name",
			&v1alpha1.ProviderSpec{
				RootDisk: &v1alpha1.RootDisk{
					VolumeClassName: "",
				},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child("spec.rootdisk.volumeclassname"), "provider spec's volumeclassname required"))),
		),
		Entry("no network name",
			&v1alpha1.ProviderSpec{
				NetworkName: "",
				RootDisk:    &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child("spec.networkname"), "provider spec's network name required"))),
		),
		Entry("no prefix name",
			&v1alpha1.ProviderSpec{
				PrefixName: "",
				RootDisk:   &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(field.Required(fldPath.Child("spec.prefixname"), "provider spec's prefixname required"))),
		),
		Entry("invalid dns server ip",
			&v1alpha1.ProviderSpec{
				RootDisk:   &v1alpha1.RootDisk{},
				DnsServers: []netip.Addr{invalidIP},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(field.Invalid(fldPath.Child("spec.dnsservers"), invalidIP, "provider spec's dnsserver valid ip required"))),
		),
	)
})
