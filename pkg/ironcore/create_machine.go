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
	computev1alpha1ac "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/compute/v1alpha1"
	ipamv1alpha1ac "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/ipam/v1alpha1"
	networkingv1alpha1ac "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/networking/v1alpha1"
	storagev1alpha1ac "github.com/ironcore-dev/ironcore/client-go/applyconfigurations/storage/v1alpha1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1alpha1 "github.com/ironcore-dev/ironcore/api/common/v1alpha1"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	corev1alpha1 "github.com/ironcore-dev/ironcore/api/core/v1alpha1"
	apiv1alpha1 "github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/validation"
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/ignition"
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
	userData, err := d.getUserData(req)
	if err != nil {
		return nil, err
	}

	ignitionSecretApplyConfig, ignitionSecretKey, err := d.buildIgnitionSecretApplyConfig(ctx, req, providerSpec, userData)
	if err != nil {
		return nil, err
	}

	machinePoolRef, machinePoolSelector, err := d.resolveMachinePool(ctx, req.MachineClass.NodeTemplate.Zone, req.MachineClass.NodeTemplate.Region)
	if err != nil {
		return nil, err
	}

	if err := d.IroncoreClient.Apply(ctx, ignitionSecretApplyConfig, client.FieldOwner(fieldOwner), client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to apply ignition secret for machine %s: %v", req.Machine.Name, err))
	}

	machineApplyConfig := d.buildMachineApplyConfig(ctx, req, providerSpec, ignitionSecretKey, machinePoolRef, machinePoolSelector)
	if err := d.IroncoreClient.Apply(ctx, machineApplyConfig, client.FieldOwner(fieldOwner), client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying ironcore machine: %s", err.Error()))

	}

	return &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Machine.Name,
			Namespace: d.IroncoreNamespace,
		},
	}, nil
}

func (d *ironcoreDriver) getUserData(req *driver.CreateMachineRequest) ([]byte, error) {
	userData, ok := req.Secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find user-data in machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}

	return userData, nil
}

func (d *ironcoreDriver) buildIgnitionSecretApplyConfig(ctx context.Context, req *driver.CreateMachineRequest, providerSpec *apiv1alpha1.ProviderSpec, userData []byte) (*corev1ac.SecretApplyConfiguration, string, error) {
	config := &ignition.Config{
		Hostname:         req.Machine.Name,
		UserData:         string(userData),
		Ignition:         providerSpec.Ignition,
		DnsServers:       providerSpec.DnsServers,
		IgnitionOverride: providerSpec.IgnitionOverride,
	}
	ignitionContent, err := ignition.File(config)
	if err != nil {
		return nil, "", status.Error(codes.Internal, fmt.Sprintf("failed to create ignition file for machine %s: %v", req.Machine.Name, err))
	}

	ignitionSecretKey := getIgnitionKeyOrDefault(providerSpec.IgnitionSecretKey)

	secret := corev1ac.Secret(
		d.getIgnitionNameForMachine(ctx, req.Machine.Name),
		d.IroncoreNamespace,
	).WithData(map[string][]byte{
		ignitionSecretKey: []byte(ignitionContent),
	})

	return secret, ignitionSecretKey, nil
}

func (d *ironcoreDriver) resolveMachinePool(ctx context.Context, zone string, region string) (*corev1.LocalObjectReference, map[string]string, error) {
	pool := &computev1alpha1.MachinePool{}

	if err := d.IroncoreClient.Get(ctx, client.ObjectKey{Name: zone}, pool); err != nil {
		if !apierrors.IsNotFound(err) {
			return nil, nil, status.Error(codes.Internal, fmt.Sprintf("error checking machine pool %q: %s", zone, err.Error()))
		}

		machinePoolSelector := map[string]string{
			string(commonv1alpha1.TopologyLabelZone): zone,
		}
		if region != "" {
			machinePoolSelector[string(commonv1alpha1.TopologyLabelRegion)] = region
		}

		return nil, machinePoolSelector, nil
	}

	return &corev1.LocalObjectReference{Name: zone}, nil, nil
}

func (d *ironcoreDriver) buildMachineApplyConfig(ctx context.Context, req *driver.CreateMachineRequest, providerSpec *apiv1alpha1.ProviderSpec, ignitionSecretKey string, machinePoolRef *corev1.LocalObjectReference, machinePoolSelector map[string]string) *computev1alpha1ac.MachineApplyConfiguration {
	volumes := d.buildMachineVolumes(providerSpec)

	spec := computev1alpha1ac.MachineSpec().
		WithMachineClassRef(corev1.LocalObjectReference{Name: req.MachineClass.NodeTemplate.InstanceType}).
		WithPower(computev1alpha1.PowerOn).
		WithNetworkInterfaces(computev1alpha1ac.NetworkInterface().
			WithName("nic").
			WithEphemeral(computev1alpha1ac.EphemeralNetworkInterfaceSource().
				WithNetworkInterfaceTemplate(networkingv1alpha1ac.NetworkInterfaceTemplateSpec().
					WithLabels(providerSpec.Labels).
					WithSpec(networkingv1alpha1ac.NetworkInterfaceSpec().
						WithNetworkRef(corev1.LocalObjectReference{Name: providerSpec.NetworkName}).
						WithIPFamilies(corev1.IPv4Protocol).
						WithIPs(networkingv1alpha1ac.IPSource().
							WithEphemeral(networkingv1alpha1ac.EphemeralPrefixSource().
								WithPrefixTemplate(ipamv1alpha1ac.PrefixTemplateSpec().
									WithSpec(ipamv1alpha1ac.PrefixSpec().
										WithPrefixLength(32).
										WithParentRef(corev1.LocalObjectReference{Name: providerSpec.PrefixName}),
									),
								),
							),
						),
					),
				),
			),
		).
		WithIgnitionRef(commonv1alpha1.SecretKeySelector{
			Name: d.getIgnitionNameForMachine(ctx, req.Machine.Name),
			Key:  ignitionSecretKey,
		}).
		WithVolumes(volumes...)

	if machinePoolRef != nil {
		spec.WithMachinePoolRef(corev1.LocalObjectReference{Name: machinePoolRef.Name})
	}

	if machinePoolSelector != nil {
		spec.WithMachinePoolSelector(machinePoolSelector)
	}

	return computev1alpha1ac.Machine(req.Machine.Name, d.IroncoreNamespace).
		WithLabels(providerSpec.Labels).
		WithSpec(spec)
}

func (d *ironcoreDriver) buildMachineVolumes(providerSpec *apiv1alpha1.ProviderSpec) []*computev1alpha1ac.VolumeApplyConfiguration {
	if providerSpec.RootDisk == nil {
		return []*computev1alpha1ac.VolumeApplyConfiguration{computev1alpha1ac.Volume().
			WithName("root").
			WithLocalDisk(computev1alpha1ac.LocalDiskVolumeSource().
				WithImage(providerSpec.Image),
			),
		}
	}

	return []*computev1alpha1ac.VolumeApplyConfiguration{computev1alpha1ac.Volume().
		WithName("root").
		WithEphemeral(computev1alpha1ac.EphemeralVolumeSource().
			WithVolumeTemplate(storagev1alpha1ac.VolumeTemplateSpec().
				WithSpec(storagev1alpha1ac.VolumeSpec().
					WithVolumeClassRef(corev1.LocalObjectReference{Name: providerSpec.RootDisk.VolumeClassName}).
					WithResources(corev1alpha1.ResourceList{corev1alpha1.ResourceStorage: providerSpec.RootDisk.Size}).
					WithDataSource(storagev1alpha1ac.VolumeDataSource().
						WithOSImage(storagev1alpha1ac.OSDataSource().
							WithImage(providerSpec.Image),
						),
					),
				),
			),
		),
	}
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
