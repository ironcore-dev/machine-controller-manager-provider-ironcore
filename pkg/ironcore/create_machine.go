// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

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

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	ipamv1alpha1 "github.com/ironcore-dev/ironcore/api/ipam/v1alpha1"
	networkingv1alpha1 "github.com/ironcore-dev/ironcore/api/networking/v1alpha1"
	storagev1alpha1 "github.com/ironcore-dev/ironcore/api/storage/v1alpha1"
	apiv1alpha1 "github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/validation"
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/ignition"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

// CreateMachine handles a machine creation request
func (d *ironcoreDriver) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
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

	ironcoreMachine, err := d.applyIronCoreMachine(ctx, req, providerSpec)
	if err != nil {
		return nil, err
	}

	return &driver.CreateMachineResponse{
		ProviderID: getProviderIDForIroncoreMachine(ironcoreMachine),
		NodeName:   ironcoreMachine.Name,
	}, nil
}

// isEmptyCreateRequest checks if any of the fields in CreateMachineRequest is empty
func isEmptyCreateRequest(req *driver.CreateMachineRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}

// applyIronCoreMachine takes care of creating actual ironcore Machine object with proper ignition data
func (d *ironcoreDriver) applyIronCoreMachine(ctx context.Context, req *driver.CreateMachineRequest, providerSpec *apiv1alpha1.ProviderSpec) (*computev1alpha1.Machine, error) {
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
			Name:      d.getIgnitionNameForMachine(ctx, req.Machine.Name),
			Namespace: d.IroncoreNamespace,
		},
		Data: ignitionData,
	}

	// Determine pool selection strategy: direct reference (old model) vs topology labels (new model).
	// In the old model, the zone name was the pool name (1:1 logical pools).
	// In the new model, multiple pools are grouped by topology labels.
	zone := req.MachineClass.NodeTemplate.Zone
	region := req.MachineClass.NodeTemplate.Region

	var machinePoolRef *corev1.LocalObjectReference
	var machinePoolSelector map[string]string

	pool := &computev1alpha1.MachinePool{}
	if err := d.IroncoreClient.Get(ctx, client.ObjectKey{Name: zone}, pool); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, status.Error(codes.Internal, fmt.Sprintf("error checking machine pool %q: %s", zone, err.Error()))
		}
		// Pool not found by name -> use topology label-based selection
		machinePoolSelector = map[string]string{
			string(commonv1alpha1.TopologyLabelZone): zone,
		}
		if region != "" {
			machinePoolSelector[string(commonv1alpha1.TopologyLabelRegion)] = region
		}
	} else {
		// Pool found by name -> old behavior (direct reference)
		machinePoolRef = &corev1.LocalObjectReference{Name: zone}
	}

	ironcoreMachine := &computev1alpha1.Machine{
		TypeMeta: metav1.TypeMeta{
			APIVersion: computev1alpha1.SchemeGroupVersion.String(),
			Kind:       "Machine",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Machine.Name,
			Namespace: d.IroncoreNamespace,
			Labels:    providerSpec.Labels,
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef: corev1.LocalObjectReference{
				Name: req.MachineClass.NodeTemplate.InstanceType,
			},
			MachinePoolRef:      machinePoolRef,
			MachinePoolSelector: machinePoolSelector,
			Power:               computev1alpha1.PowerOn,
			NetworkInterfaces: []computev1alpha1.NetworkInterface{
				{
					Name: "nic",
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
			IgnitionRef: &commonv1alpha1.SecretKeySelector{
				Name: ignitionSecret.Name,
				Key:  ignitionSecretKey,
			},
		},
	}

	if providerSpec.RootDisk == nil {
		ironcoreMachine.Spec.Volumes = []computev1alpha1.Volume{
			{
				Name: "root",
				VolumeSource: computev1alpha1.VolumeSource{
					LocalDisk: &computev1alpha1.LocalDiskVolumeSource{
						Image: providerSpec.Image,
					},
				},
			},
		}
	} else {
		ironcoreMachine.Spec.Volumes = []computev1alpha1.Volume{
			{
				Name: "root",
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
								//TODO remove once image field is removed from API spec
								Image: providerSpec.Image,
								DataSource: storagev1alpha1.VolumeDataSource{
									OSImage: &storagev1alpha1.OSDataSource{
										Image: providerSpec.Image,
									},
								},
							},
						},
					},
				},
			},
		}
	}

	if err := d.IroncoreClient.Patch(ctx, ironcoreMachine, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying ironcore machine: %s", err.Error()))
	}

	if err := d.IroncoreClient.Patch(ctx, ignitionSecret, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying ignition secret: %s", err.Error()))
	}

	return ironcoreMachine, nil
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
