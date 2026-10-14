// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net/netip"

	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
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
			ContainElement(field.Required(fldPath.Child("spec.secretRef"), "secretRef is required")),
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
			ContainElement(field.Required(fldPath.Child("userData"), "userData is required")),
		),
		Entry("no image",
			&v1alpha1.ProviderSpec{
				Image:    "",
				RootDisk: &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(field.Required(fldPath.Child("spec.image"), "image is required")),
		),
		Entry("rootDisk with size and no volumeClassName is valid",
			newProviderSpec(&v1alpha1.RootDisk{
				Size: resource.MustParse("10Gi"),
			}),
			newSecret(),
			fldPath,
			BeEmpty(),
		),
		Entry("rootDisk without size or volumeClassName",
			newProviderSpec(&v1alpha1.RootDisk{}),
			newSecret(),
			fldPath,
			ConsistOf(field.Required(fldPath.Child("spec.rootDisk.size"), "size is required when volumeClassName is not set")),
		),
		Entry("rootDisk with negative size and no volumeClassName",
			newProviderSpec(&v1alpha1.RootDisk{
				Size: resource.MustParse("-1Gi"),
			}),
			newSecret(),
			fldPath,
			ConsistOf(field.Invalid(fldPath.Child("spec.rootDisk.size"), "-1Gi", "must be greater than zero")),
		),
		Entry("rootDisk with volumePoolName and no volumeClassName",
			newProviderSpec(&v1alpha1.RootDisk{
				Size:           resource.MustParse("10Gi"),
				VolumePoolName: "bar",
			}),
			newSecret(),
			fldPath,
			ConsistOf(field.Forbidden(fldPath.Child("spec.rootDisk.volumePoolName"), "volumePoolName requires volumeClassName")),
		),
		Entry("rootDisk with volumeClassName, size and volumePoolName is valid",
			newProviderSpec(&v1alpha1.RootDisk{
				Size:            resource.MustParse("10Gi"),
				VolumeClassName: "foo",
				VolumePoolName:  "bar",
			}),
			newSecret(),
			fldPath,
			BeEmpty(),
		),
		Entry("no network name",
			&v1alpha1.ProviderSpec{
				NetworkName: "",
				RootDisk:    &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(field.Required(fldPath.Child("spec.networkName"), "networkName is required")),
		),
		Entry("no prefix name",
			&v1alpha1.ProviderSpec{
				PrefixName: "",
				RootDisk:   &v1alpha1.RootDisk{},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(field.Required(fldPath.Child("spec.prefixName"), "prefixName is required")),
		),
		Entry("invalid dns server ip",
			&v1alpha1.ProviderSpec{
				RootDisk:   &v1alpha1.RootDisk{},
				DnsServers: []netip.Addr{invalidIP},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(field.Invalid(fldPath.Child("spec.dnsServers[0]"), invalidIP, "ip is invalid")),
		),
	)
})

// newProviderSpec returns a spec that is valid except for the given root disk.
func newProviderSpec(rootDisk *v1alpha1.RootDisk) *v1alpha1.ProviderSpec {
	return &v1alpha1.ProviderSpec{
		Image:       "foo-image",
		NetworkName: "bar-network",
		PrefixName:  "baz-prefix",
		RootDisk:    rootDisk,
	}
}

func newSecret() *corev1.Secret {
	return &corev1.Secret{Data: map[string][]byte{"userData": []byte("abcd")}}
}
