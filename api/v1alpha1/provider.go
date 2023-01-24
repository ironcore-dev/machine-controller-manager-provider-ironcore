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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"

	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
)

const (
	// V1Alpha1 is the API version
	V1Alpha1 = "mcm.gardener.cloud/v1alpha1"
	// ProviderName is the provider name
	ProviderName = "onmetal"
)

// ProviderSpec is the spec to be used while parsing the calls
type ProviderSpec struct {
	// MachineClassRef is a reference to the MachineClass of the Machine
	MachineClassRef corev1.LocalObjectReference `json:"machineClassRef,omitempty"`

	// MachinePoolSelector selects a suitable MachinePoolRef by the given labels
	MachinePoolSelector map[string]string `json:"machinePoolSelector,omitempty"`

	// MachinePoolRef defines MachinePool to run the Machine on.
	// If empty a scheduler will figure out an appropriate pool to run the Machine on
	MachinePoolRef *corev1.LocalObjectReference `json:"machinePoolRef,omitempty"`

	// Image is the URL pointing to an OCI registry containing the operating system image which should be used to boot the Machine
	Image string `json:"image,omitempty"`

	// ImagePullSecretRef is an optional secret for pulling the image of a Machine
	ImagePullSecretRef *corev1.LocalObjectReference `json:"imagePullSecretRef,omitempty"`

	// NetworkInterfaces defines a list of NetworkInterfaces used by a Machine
	NetworkInterfaces []computev1alpha1.NetworkInterface `json:"networkInterfaces,omitempty"`

	// Volumes is a list of Volumes used by a Machine
	Volumes []computev1alpha1.Volume `json:"volumes,omitempty"`

	// Ignition contains the ignition configuration which should be run on first boot of a Machine.
	Ignition string `json:"ignition,omitempty"`

	// By default if ignition is set it will be merged it with our template
	// If IgnitionOverride is set to true allows to fully override
	IgnitionOverride bool `json:"ignitionOverride,omitempty"`

	// IgnitionSecretKey is optional key field used to identify the ignition content in the Secret
	// If the key is empty, the DefaultIgnitionKey will be used as fallback.
	IgnitionSecretKey string `json:"ignitionSecretKey,omitempty"`
}
