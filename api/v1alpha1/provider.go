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
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// V1Alpha1 is the API version
	V1Alpha1 = "mcm.gardener.cloud/v1alpha1"
	// ProviderName is the provider name
	ProviderName = "onmetal"
)

// ProviderSpec is the spec to be used while parsing the calls
type ProviderSpec struct {
	// Image is the URL pointing to an OCI registry containing the operating system image which should be used to boot the Machine
	Image string `json:"image,omitempty"`
	// Ignition contains the ignition configuration which should be run on first boot of a Machine.
	Ignition string `json:"ignition,omitempty"`
	// By default, if ignition is set it will be merged it with our template
	// If IgnitionOverride is set to true allows to fully override
	IgnitionOverride bool `json:"ignitionOverride,omitempty"`
	// IgnitionSecretKey is optional key field used to identify the ignition content in the Secret
	// If the key is empty, the DefaultIgnitionKey will be used as fallback.
	IgnitionSecretKey string `json:"ignitionSecretKey,omitempty"`
	// RootDisk defines the root disk properties of the Machine.
	RootDisk *RootDisk `json:"rootDisk,omitempty"`
	// NetworkName is the Network to be used for the Machine's NetworkInterface.
	NetworkName string `json:"networkName"`
	// PrefixName is the parent Prefix from which an IP should be allocated for the Machine's NetworkInterface.
	PrefixName string `json:"prefixName"`
	// Labels are used to tag resources which the MCM creates, so they can be identified later.
	Labels map[string]string `json:"labels,omitempty"`
}

// RootDisk defines the root disk properties of the Machine.
type RootDisk struct {
	// Size defines the volume size of the root disk.
	Size resource.Quantity `json:"size"`
	// VolumeClassName defines which volume class to use for the root disk.
	VolumeClassName string `json:"volumeClassName"`
	// VolumePoolName defines on which VolumePool a Volume should be scheduled.
	VolumePoolName string `json:"volumePoolName,omitempty"`
}
