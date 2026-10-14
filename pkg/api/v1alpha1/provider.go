// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	"net/netip"

	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	// V1Alpha1 is the API version
	V1Alpha1 = "mcm.gardener.cloud/v1alpha1"
	// ProviderName is the provider name
	ProviderName = "ironcore"
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
	// DnsServers is a list of DNS resolvers which should be configured on the host.
	DnsServers []netip.Addr `json:"dnsServers,omitempty"`
}

// RootDisk defines the root disk properties of the Machine.
// With VolumeClassName set, the root disk is an ephemeral Volume of that class, populated from the image.
// Otherwise the root disk is a local disk of the image, with Size as its sizeLimit.
type RootDisk struct {
	// Size defines the volume size of the root disk.
	// Required if VolumeClassName is empty.
	Size resource.Quantity `json:"size"`
	// VolumeClassName defines which volume class to use for the root disk.
	// If empty, the root disk is a local disk.
	VolumeClassName string `json:"volumeClassName"`
	// VolumePoolName defines on which VolumePool a Volume should be scheduled.
	// Must be empty if VolumeClassName is empty.
	VolumePoolName string `json:"volumePoolName,omitempty"`
}
