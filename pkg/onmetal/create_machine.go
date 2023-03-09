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
	"context"
	"encoding/json"
	"fmt"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/validation"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/ignition"
	"github.com/onmetal/onmetal-api/api/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	corev1alpha1 "github.com/onmetal/onmetal-api/api/core/v1alpha1"
	ipamv1alpha1 "github.com/onmetal/onmetal-api/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/onmetal/onmetal-api/api/networking/v1alpha1"
	storagev1alpha1 "github.com/onmetal/onmetal-api/api/storage/v1alpha1"
)

// CreateMachine handles a machine creation request
func (d *onmetalDriver) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	if isEmptyCreateRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not supported by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine creation request has been received for %s", req.Machine.Name)
	defer klog.V(3).Infof("Machine creation request has been processed for %s", req.Machine.Name)

	providerSpec, err := validateProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, err
	}

	onmetalMachine, err := d.applyOnMetalMachine(ctx, req, providerSpec)
	if err != nil {
		return nil, err
	}

	return &driver.CreateMachineResponse{
		ProviderID: getProviderIDForOnmetalMachine(onmetalMachine),
		NodeName:   onmetalMachine.Name,
	}, nil
}

// isEmptyCreateRequest checks if any of the fields in CreateMachineRequest is empty
func isEmptyCreateRequest(req *driver.CreateMachineRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}

// applyOnMetalMachine takes care of creating actual onemetal Machine object with proper ignition data
func (d *onmetalDriver) applyOnMetalMachine(ctx context.Context, req *driver.CreateMachineRequest, providerSpec *apiv1alpha1.ProviderSpec) (*computev1alpha1.Machine, error) {
	// Get userData from machine secret
	userData, ok := req.Secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find user-data in machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}

	// Construct ignition file config
	config := &ignition.Config{
		Hostname:         req.Machine.Name,
		UserData:         string(userData),
		Ignition:         providerSpec.Ignition,
		DnsServers:       providerSpec.DnsServers,
		IgnitionOverride: providerSpec.IgnitionOverride,
	}
	ignitionContent, err := ignition.File(config)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create ignition file for machine %s: %v", req.Machine.Name, err))
	}

	ignitionSecretKey := getIgnitionKeyOrDefault(providerSpec.IgnitionSecretKey)
	ignitionData := map[string][]byte{}
	ignitionData[ignitionSecretKey] = []byte(ignitionContent)
	ignitionSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getIgnitionNameForMachine(req.Machine.Name),
			Namespace: d.OnmetalNamespace,
		},
		Data: ignitionData,
	}

	onmetalMachine := &computev1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: computev1alpha1.SchemeGroupVersion.String(),
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Machine.Name,
			Namespace: d.OnmetalNamespace,
			Labels:    providerSpec.Labels,
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef: corev1.LocalObjectReference{
				Name: req.MachineClass.NodeTemplate.InstanceType,
			},
			MachinePoolRef: &corev1.LocalObjectReference{
				Name: req.MachineClass.NodeTemplate.Zone,
			},
			Power: computev1alpha1.PowerOn,
			NetworkInterfaces: []computev1alpha1.NetworkInterface{
				{
					Name: "primary",
					NetworkInterfaceSource: computev1alpha1.NetworkInterfaceSource{
						Ephemeral: &computev1alpha1.EphemeralNetworkInterfaceSource{
							NetworkInterfaceTemplate: &networkingv1alpha1.NetworkInterfaceTemplateSpec{
								ObjectMeta: metav1.ObjectMeta{
									Labels: providerSpec.Labels,
								},
								Spec: networkingv1alpha1.NetworkInterfaceSpec{
									NetworkRef: corev1.LocalObjectReference{
										Name: providerSpec.NetworkName,
									},
									IPFamilies: []corev1.IPFamily{corev1.IPv4Protocol},
									IPs: []networkingv1alpha1.IPSource{
										{
											Ephemeral: &networkingv1alpha1.EphemeralPrefixSource{
												PrefixTemplate: &ipamv1alpha1.PrefixTemplateSpec{
													Spec: ipamv1alpha1.PrefixSpec{
														// request single IP
														PrefixLength: 32,
														ParentRef: &corev1.LocalObjectReference{
															Name: providerSpec.PrefixName,
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
				},
			},
			IgnitionRef: &v1alpha1.SecretKeySelector{
				Name: ignitionSecret.Name,
				Key:  ignitionSecretKey,
			},
		},
	}

	if providerSpec.RootDisk == nil {
		onmetalMachine.Spec.Image = providerSpec.Image
	} else {
		onmetalMachine.Spec.Volumes = []computev1alpha1.Volume{
			{
				Name: "primary",
				VolumeSource: computev1alpha1.VolumeSource{
					Ephemeral: &computev1alpha1.EphemeralVolumeSource{
						VolumeTemplate: &storagev1alpha1.VolumeTemplateSpec{
							Spec: storagev1alpha1.VolumeSpec{
								VolumeClassRef: &corev1.LocalObjectReference{
									Name: providerSpec.RootDisk.VolumeClassName,
								},
								Resources: corev1alpha1.ResourceList{
									corev1alpha1.ResourceStorage: providerSpec.RootDisk.Size,
								},
								Image: providerSpec.Image,
							},
						},
					},
				},
			},
		}
	}

	if err := d.OnmetelClient.Patch(ctx, onmetalMachine, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying onmetal machine: %s", err.Error()))
	}

	if err := d.OnmetelClient.Patch(ctx, ignitionSecret, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying ignition secret: %s", err.Error()))
	}

	return onmetalMachine, nil
}

// getIgnitionKeyOrDefault checks if key is empty otherwise return default ingintion key
func getIgnitionKeyOrDefault(key string) string {
	if key == "" {
		return computev1alpha1.DefaultIgnitionKey
	}
	return key
}

// validateProviderSpecAndSecret Validates providerSpec and provider secret
func validateProviderSpecAndSecret(class *machinev1alpha1.MachineClass, secret *corev1.Secret) (*apiv1alpha1.ProviderSpec, error) {
	if class == nil {
		return nil, status.Error(codes.Internal, "MachineClass in ProviderSpec is not set")
	}

	var providerSpec *apiv1alpha1.ProviderSpec
	if err := json.Unmarshal(class.ProviderSpec.Raw, &providerSpec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	validationErr := validation.ValidateProviderSpecAndSecret(providerSpec, secret, field.NewPath("providerSpec"))
	if validationErr.ToAggregate() != nil && len(validationErr.ToAggregate().Errors()) > 0 {
		err := fmt.Errorf("failed to validate provider spec and secret: %v", validationErr.ToAggregate().Errors())
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}
