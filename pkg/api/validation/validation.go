// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net/netip"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
)

// ValidateProviderSpecAndSecret validates the provider spec and provider secret
func ValidateProviderSpecAndSecret(spec *v1alpha1.ProviderSpec, secret *corev1.Secret, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = validateIroncoreMachineClassSpec(spec, field.NewPath("spec"))
	allErrs = append(allErrs, validateSecret(secret, field.NewPath("spec"))...)

	return allErrs
}

func validateSecret(secret *corev1.Secret, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if secret == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("secretRef"), "secretRef is required"))
		return allErrs
	}

	if secret.Data["userData"] == nil {
		allErrs = append(allErrs, field.Required(field.NewPath("userData"), "userData is required"))
	}

	return allErrs
}

func validateIroncoreMachineClassSpec(spec *v1alpha1.ProviderSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.RootDisk != nil && spec.RootDisk.VolumeClassName == "" && spec.RootDisk.Size.IsZero() {
		allErrs = append(allErrs, field.Required(fldPath.Child("rootDisk").Child("size"), "size is required when volumeClassName is not set"))
	}

	if spec.RootDisk != nil && spec.RootDisk.VolumeClassName == "" && spec.RootDisk.Size.Sign() < 0 {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("rootDisk").Child("size"), spec.RootDisk.Size.String(), "must be greater than zero"))
	}

	if spec.RootDisk != nil && spec.RootDisk.VolumeClassName == "" && spec.RootDisk.VolumePoolName != "" {
		allErrs = append(allErrs, field.Forbidden(fldPath.Child("rootDisk").Child("volumePoolName"), "volumePoolName requires volumeClassName"))
	}

	if spec.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("image"), "image is required"))
	}

	if spec.NetworkName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkName"), "networkName is required"))
	}

	if spec.PrefixName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("prefixName"), "prefixName is required"))
	}

	for i, ip := range spec.DnsServers {
		if !netip.Addr.IsValid(ip) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("dnsServers").Index(i), ip, "ip is invalid"))
		}
	}

	return allErrs
}
