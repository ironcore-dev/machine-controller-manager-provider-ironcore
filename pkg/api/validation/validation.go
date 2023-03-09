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

package validation

import (
	"net/netip"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
)

// ValidateProviderSpecAndSecret validates the provider spec and provider secret
func ValidateProviderSpecAndSecret(spec *v1alpha1.ProviderSpec, secret *corev1.Secret, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = validateOnmetalMachineClassSpec(spec, field.NewPath("spec"))
	allErrs = append(allErrs, validateSecret(secret, fldPath.Child("secretRef"))...)

	return allErrs
}

func validateSecret(secret *corev1.Secret, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if secret == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child(""), "secretRef is required"))
		return allErrs
	}

	if secret.Data["userData"] == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("userData"), "userData is required"))
	}

	return allErrs
}

func validateOnmetalMachineClassSpec(spec *v1alpha1.ProviderSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.RootDisk != nil && spec.RootDisk.VolumeClassName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("rootdisk volumeclassname"), "volumeclassname is required"))
	}

	if spec.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("image"), "image is required"))
	}

	if spec.NetworkName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkname"), "networkname is required"))
	}

	if spec.PrefixName == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("prefixname"), "prefixname is required"))
	}

	for i, ip := range spec.DnsServers {
		if !netip.Addr.IsValid(ip) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("dnsserver").Index(i), ip, "must specify a valid ip for dnsserver"))
		}
	}

	return allErrs
}
