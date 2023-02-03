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
	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	. "github.com/onmetal/machine-controller-manager-provider-onmetal/testutils/validation"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var fldPath *field.Path
var mydevice string
var mydevice2 = "vda"
var invDevice = "foobar"
var ptrDevice = &mydevice
var ptrDevice2 = &mydevice
var ptrDevice3 = &mydevice2
var invDeviceptr = &invDevice

func mustParseNewQuantity(s string) *resource.Quantity {
	q := resource.MustParse(s)
	return &q
}

var _ = Describe("Machine", func() {
	DescribeTable("ValidateProviderSpec",
		func(spec *v1alpha1.ProviderSpec, secret *corev1.Secret, fldPath *field.Path, match types.GomegaMatcher) {
			errList := ValidateProviderSpec(spec, secret, fldPath)
			Expect(errList).To(match)
		},
		Entry("no machine class ref",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{},
			fldPath,
			ContainElement(RequiredField("spec.machineClassRef")),
		),
		Entry("no machine pool ref",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{},
			fldPath,
			ContainElement(RequiredField("spec.machinePoolRef")),
		),
		Entry("no image pull secret ref",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{},
			fldPath,
			ContainElement(RequiredField("spec.imagePullSecretRef")),
		),
		Entry("no image",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(RequiredField("spec.image"))),
		),
		Entry("invalid machine class ref name",
			&v1alpha1.ProviderSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "foo*"},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.machineClassRef.name")),
		),
		Entry("invalid machine pool ref name",
			&v1alpha1.ProviderSpec{
				MachinePoolRef: &corev1.LocalObjectReference{Name: "foo*"},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.machinePoolRef.name")),
		),
		Entry("invalid image pull secret ref name",
			&v1alpha1.ProviderSpec{
				ImagePullSecretRef: &corev1.LocalObjectReference{Name: "foo*"},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.imagePullSecretRef.name")),
		),
		Entry("valid machine class ref subdomain name",
			&v1alpha1.ProviderSpec{
				MachineClassRef: corev1.LocalObjectReference{Name: "foo-bar"},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(InvalidField("spec.machineClassRef.name"))),
		),
		Entry("valid machine pool ref subdomain name",
			&v1alpha1.ProviderSpec{
				MachinePoolRef: &corev1.LocalObjectReference{Name: "foo.bar.baz"},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(InvalidField("spec.machinePoolRef.name"))),
		),
		Entry("valid image pull secret ref subdomain name",
			&v1alpha1.ProviderSpec{
				ImagePullSecretRef: &corev1.LocalObjectReference{Name: "foo-bar"},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(InvalidField("spec.imagePullSecretRef.name"))),
		),
		Entry("invalid volume name",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{Name: "foo*", Device: ptrDevice},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.volume[0].name")),
		),
		Entry("duplicate volume name",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{Name: "foo", Device: ptrDevice},
					{Name: "foo", Device: ptrDevice2},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(DuplicateField("spec.volume[1].name")),
		),
		Entry("invalid volumeRef name",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{Name: "foo", Device: ptrDevice, VolumeSource: computev1alpha1.VolumeSource{
						VolumeRef: &corev1.LocalObjectReference{Name: "foo*"}}},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.volume[0].volumeRef.name")),
		),
		Entry("invalid empty disk size limit quantity",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{
						Name:   "foo",
						Device: ptrDevice,
						VolumeSource: computev1alpha1.VolumeSource{
							EmptyDisk: &computev1alpha1.EmptyDiskVolumeSource{SizeLimit: mustParseNewQuantity("-1Gi")},
						},
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.volume[0].emptyDisk.sizeLimit")),
		),
		Entry("duplicate machine volume device",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{
						Name:   "foo",
						Device: ptrDevice,
					},
					{
						Name:   "bar",
						Device: ptrDevice,
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(DuplicateField("spec.volume[1].device")),
		),
		Entry("invalid volume device",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{
						Name:   "foo",
						Device: invDeviceptr,
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(InvalidField("spec.volume[0].device")),
		),
		Entry("reserved volume device",
			&v1alpha1.ProviderSpec{
				Volumes: []computev1alpha1.Volume{
					{
						Name:   "foo",
						Device: ptrDevice3,
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			ContainElement(ForbiddenField("spec.volume[0].device")),
		),
		Entry("no secret",
			&v1alpha1.ProviderSpec{},
			nil,
			fldPath,
			Not(ContainElement(RequiredField("corev1 secret"))),
		),
		Entry("no userData in secret",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{
				Data: map[string][]byte{
					"userData": nil,
				},
			},
			fldPath,
			Not(ContainElement(RequiredField("corev1 secret userData"))),
		),
		Entry("no kubeconfig in secret",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{
				Data: map[string][]byte{
					"kubeconfig": nil,
				},
			},
			fldPath,
			Not(ContainElement(RequiredField("corev1 secret kubeconfig"))),
		),
		Entry("no namespace in secret",
			&v1alpha1.ProviderSpec{},
			&corev1.Secret{
				Data: map[string][]byte{
					"namespace": nil,
				},
			},
			fldPath,
			Not(ContainElement(RequiredField("corev1 secret namespace"))),
		),
		Entry("empty network interface",
			&v1alpha1.ProviderSpec{
				NetworkInterfaces: []computev1alpha1.NetworkInterface{},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(RequiredField("atleast 1 network interface required"))),
		),
		Entry("empty name in network interface",
			&v1alpha1.ProviderSpec{
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "",
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(RequiredField("network interface name is required"))),
		),
		Entry("empty name in network interface",
			&v1alpha1.ProviderSpec{
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name: "foo",
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(RequiredField("network interface name is required"))),
		),
		Entry("empty network interface source in network interface",
			&v1alpha1.ProviderSpec{
				NetworkInterfaces: []computev1alpha1.NetworkInterface{
					{
						Name:                   "foo",
						NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{},
					},
				},
			},
			&corev1.Secret{},
			fldPath,
			Not(ContainElement(RequiredField("network interface source is required"))),
		),
	)
})
