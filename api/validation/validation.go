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
	"fmt"

	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"

	device "github.com/onmetal/machine-controller-manager-provider-onmetal/api/device"
	mcmapivalidation "github.com/onmetal/machine-controller-manager-provider-onmetal/api/validation/helper"
	corev1 "k8s.io/api/core/v1"
	apivalidation "k8s.io/apimachinery/pkg/api/validation"
	metav1validation "k8s.io/apimachinery/pkg/apis/meta/v1/validation"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var ValidateMachinePoolName = apivalidation.NameIsDNSSubdomain
var ValidateVolumePoolName = apivalidation.NameIsDNSSubdomain
var allowedVolumeTemplateObjectMetaFields = sets.New("Annotations", "Labels")

// ValidateProviderSpec validates the provider spec and provider secret
func ValidateProviderSpec(spec *v1alpha1.ProviderSpec, secret *corev1.Secret, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	allErrs = validateOnmetalMachineClassSpec(spec, field.NewPath("spec"))
	allErrs = append(allErrs, validateSecret(secret, field.NewPath("secretRef"))...)

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

	if secret.Data["kubeconfig"] == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("kubeconfig"), "kubeconfig is required"))
	}

	if secret.Data["namespace"] == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("namespace"), "namespace is required"))
	}

	return allErrs
}

func validateOnmetalMachineClassSpec(spec *v1alpha1.ProviderSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.MachineClassRef == (corev1.LocalObjectReference{}) {
		allErrs = append(allErrs, field.Required(fldPath.Child("machineClassRef"), "Machine Class Reference is required"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(spec.MachineClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("machineClassRef").Child("name"), spec.MachineClassRef.Name, msg))
	}

	if spec.MachinePoolRef == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("machinePoolRef"), "Machine Pool Reference is required"))
	}

	if spec.MachinePoolRef != nil {
		for _, msg := range ValidateMachinePoolName(spec.MachinePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("machinePoolRef").Child("name"), spec.MachinePoolRef.Name, msg))
		}
	}

	if spec.Image == "" {
		allErrs = append(allErrs, field.Required(fldPath.Child("Image"), "Image is required"))
	}

	if spec.ImagePullSecretRef == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("imagePullSecretRef"), "Image Pull Secret Reference is required"))
	}

	if spec.ImagePullSecretRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.ImagePullSecretRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("imagePullSecretRef").Child("name"), spec.ImagePullSecretRef.Name, msg))
		}
	}

	if len(spec.MachinePoolSelector) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("MachinePoolSelector"), "Machine pool selector is required"))
	}

	seenNames := sets.New[string]()
	seenDevices := sets.New[string]()
	for i, vol := range spec.Volumes {
		if seenNames.Has(vol.Name) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("volume").Index(i).Child("name"), vol.Name))
		} else {
			seenNames.Insert(vol.Name)
		}
		if seenDevices.Has(*vol.Device) {
			allErrs = append(allErrs, field.Duplicate(fldPath.Child("volume").Index(i).Child("device"), *vol.Device))
		} else {
			seenDevices.Insert(*vol.Device)
		}
		allErrs = append(allErrs, validateMCMVolume(&vol, fldPath.Child("volume").Index(i))...)
	}

	allErrs = append(allErrs, validateOnmetalNetworkInterfaces(spec.NetworkInterfaces, fldPath.Child("networkInterfaces"))...)

	return allErrs
}

func validateOnmetalNetworkInterfaces(interfaces []computev1alpha1.NetworkInterface, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList
	if len(interfaces) == 0 {
		allErrs = append(allErrs, field.Required(fldPath.Child("networkInterfaces"), "at least one network interface is required"))
	}

	for i, nic := range interfaces {
		idxPath := fldPath.Index(i)
		emptyNic := computev1alpha1.NetworkInterfaceSource{}
		if nic.Name == "" && emptyNic == nic.NetworkInterfaceSource {
			allErrs = append(allErrs, field.Required(idxPath, "either network name or network Interface source or both is required"))
		}
	}

	return allErrs
}

func isReservedDeviceName(prefix string, idx int) bool {
	return idx == 0
}

func validateMCMVolume(volume *computev1alpha1.Volume, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	for _, msg := range apivalidation.NameIsDNSLabel(volume.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("name"), volume.Name, msg))
	}

	if volume.Device == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("device"), "must specify device"))
	} else {
		// TODO: Improve validation on prefix.
		prefix, idx, err := device.ParseName(*volume.Device)
		if err != nil {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("device"), volume.Device, fmt.Sprintf("invalid device name: %v", err)))
		} else {
			if isReservedDeviceName(prefix, idx) {
				allErrs = append(allErrs, field.Forbidden(fldPath.Child("device"), fmt.Sprintf("device name %s is reserved", *volume.Device)))
			}
		}
	}

	allErrs = append(allErrs, validateVolumeSource(&volume.VolumeSource, fldPath)...)

	return allErrs
}

func validateVolumeSource(source *computev1alpha1.VolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	var numDefs int
	if source.VolumeRef != nil {
		numDefs++
		for _, msg := range apivalidation.NameIsDNSLabel(source.VolumeRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeRef").Child("name"), source.VolumeRef.Name, msg))
		}
	}
	if source.EmptyDisk != nil {
		if numDefs > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("emptyDisk"), "must only specify one volume source"))
		} else {
			numDefs++
			allErrs = append(allErrs, validateEmptyDiskVolumeSource(source.EmptyDisk, fldPath.Child("emptyDisk"))...)
		}
	}
	if source.Ephemeral != nil {
		if numDefs > 0 {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("ephemeral"), "must only specify one volume source"))
		} else {
			numDefs++
			allErrs = append(allErrs, validateEphemeralVolumeSource(source.Ephemeral, fldPath.Child("ephemeral"))...)
		}
	}
	if numDefs == 0 {
		allErrs = append(allErrs, field.Invalid(fldPath, source, "must specify at least one volume source"))
	}

	return allErrs
}

func validateEmptyDiskVolumeSource(source *computev1alpha1.EmptyDiskVolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if sizeLimit := source.SizeLimit; sizeLimit != nil {
		allErrs = append(allErrs, mcmapivalidation.ValidateNonNegativeQuantity(*sizeLimit, fldPath.Child("sizeLimit"))...)
	}

	return allErrs
}

func validateEphemeralVolumeSource(source *computev1alpha1.EphemeralVolumeSource, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if source.VolumeTemplate == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeTemplate"), "must specify volume template"))
	} else {
		allErrs = append(allErrs, validateVolumeTemplateSpecForMachine(source.VolumeTemplate, fldPath.Child("volumeTemplate"))...)
	}

	return allErrs
}

func validateVolumeTemplateSpecForMachine(template *storagev1alpha1.VolumeTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if template == nil {
		allErrs = append(allErrs, field.Required(fldPath, ""))
	} else {
		allErrs = append(allErrs, validateVolumeTemplateSpec(template, fldPath)...)

		if template.Spec.ClaimRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("spec", "claimRef"), "may not specify claimRef"))
		}
		if template.Spec.Unclaimable {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("spec", "unclaimable"), "may not specify unclaimable"))
		}
	}

	return allErrs
}

func validateVolumeTemplateSpec(spec *storagev1alpha1.VolumeTemplateSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	allErrs = append(allErrs, metav1validation.ValidateLabels(spec.ObjectMeta.Labels, fldPath.Child("labels"))...)
	allErrs = append(allErrs, apivalidation.ValidateAnnotations(spec.ObjectMeta.Annotations, fldPath.Child("annotations"))...)
	allErrs = append(allErrs, mcmapivalidation.ValidateFieldAllowList(spec.ObjectMeta, allowedVolumeTemplateObjectMetaFields, "cannot be set for a volume template", fldPath)...)
	allErrs = append(allErrs, validateVolumeSpec(&spec.Spec, fldPath.Child("spec"))...)
	return allErrs
}

func validateVolumeSpec(spec *storagev1alpha1.VolumeSpec, fldPath *field.Path) field.ErrorList {
	var allErrs field.ErrorList

	if spec.VolumeClassRef == nil {
		allErrs = append(allErrs, field.Required(fldPath.Child("volumeClassRef"), "must specify a volume class ref"))
	}

	for _, msg := range apivalidation.NameIsDNSLabel(spec.VolumeClassRef.Name, false) {
		allErrs = append(allErrs, field.Invalid(fldPath.Child("volumeClassRef").Child("name"), spec.VolumeClassRef.Name, msg))
	}
	allErrs = append(allErrs, metav1validation.ValidateLabels(spec.VolumePoolSelector, fldPath.Child("volumePoolSelector"))...)

	if spec.VolumePoolRef != nil {
		for _, msg := range ValidateVolumePoolName(spec.VolumePoolRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("volumePoolRef").Child("name"), spec.VolumePoolRef.Name, msg))
		}
	}

	if spec.Unclaimable {
		if spec.ClaimRef != nil {
			allErrs = append(allErrs, field.Forbidden(fldPath.Child("claimRef"), "cannot specify unclaimable and claimRef"))
		}
	} else {
		if spec.ClaimRef != nil {
			for _, msg := range apivalidation.NameIsDNSLabel(spec.ClaimRef.Name, false) {
				allErrs = append(allErrs, field.Invalid(fldPath.Child("claimRef").Child("name"), spec.ClaimRef.Name, msg))
			}
		}
	}

	storageValue, ok := spec.Resources[corev1.ResourceStorage]
	if !ok {
		allErrs = append(allErrs, field.Required(fldPath.Child("resources").Key(string(corev1.ResourceStorage)), ""))
	} else {
		allErrs = append(allErrs, mcmapivalidation.ValidatePositiveQuantity(storageValue, fldPath.Child("resources").Key(string(corev1.ResourceStorage)))...)
	}

	if spec.ImagePullSecretRef != nil {
		for _, msg := range apivalidation.NameIsDNSLabel(spec.ImagePullSecretRef.Name, false) {
			allErrs = append(allErrs, field.Invalid(fldPath.Child("imagePullSecretRef").Child("name"), spec.ImagePullSecretRef.Name, msg))
		}
	}

	return allErrs
}
