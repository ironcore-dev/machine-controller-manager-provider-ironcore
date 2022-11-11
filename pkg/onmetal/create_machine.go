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
	"encoding/base64"
	"encoding/json"
	"fmt"

	machinev1alpha1 "github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	apiv1alpha1 "github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/validation"
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/ignition"
	"github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (d *onmetalDriver) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	if isEmptyCreateRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
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

func isEmptyCreateRequest(req *driver.CreateMachineRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}

func (d *onmetalDriver) applyOnMetalMachine(ctx context.Context, req *driver.CreateMachineRequest, providerSpec *apiv1alpha1.ProviderSpec) (*computev1alpha1.Machine, error) {
	// Get userData from machine secret
	userDataSecret, ok := req.Secret.Data["userData"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find user-data in machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}
	// Get namespace from machine secret
	namespace, ok := req.Secret.Data["namespace"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find namespace is machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}

	// TODO: retrieve ssh keys

	// Construct ignition file config
	config := &ignition.Config{
		PasswdHash:     "*", // TODO: handle password for debug later
		Hostname:       req.Machine.Name,
		UserdataBase64: base64.StdEncoding.EncodeToString(userDataSecret),
		SSHKeys:        []string{},
		InstallPath:    "/var/lib/coreos-install",
	}
	ignitionContent, err := ignition.File(config)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create ignition file for machine %s: %v", req.Machine.Name, err))
	}

	ignitionData := map[string][]byte{}
	ignitionData[providerSpec.IgnitionSecretKey] = []byte(ignitionContent)
	ignitionSecret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      getIgnitionNameForMachine(req.Machine.Name),
			Namespace: string(namespace),
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
			Namespace: d.Namespace,
			Labels: map[string]string{
				labelKeyProvider: apiv1alpha1.ProviderName,
				labelKeyApp:      labelValueMachine,
			},
		},
		Spec: computev1alpha1.MachineSpec{
			MachineClassRef:     providerSpec.MachineClassRef,
			MachinePoolSelector: providerSpec.MachinePoolSelector,
			MachinePoolRef:      providerSpec.MachinePoolRef,
			Image:               providerSpec.Image,
			ImagePullSecretRef:  providerSpec.ImagePullSecretRef,
			NetworkInterfaces:   providerSpec.NetworkInterfaces,
			Volumes:             providerSpec.Volumes,
			IgnitionRef: &v1alpha1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: ignitionSecret.Name,
				},
				Key: getIgnitionKeyOrDefault(providerSpec.IgnitionSecretKey),
			},
		},
	}

	// Create k8s client for the user provided machine secret. This client will be used
	// to create the resources in the user provided namespace.
	k8sClient, err := d.createK8sClient(req.Secret)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create k8s client for machine secret %s: %v", client.ObjectKeyFromObject(req.Secret), err))
	}

	if err := k8sClient.Patch(ctx, onmetalMachine, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying onmetal machine: %s", err.Error()))
	}

	if err := controllerutil.SetControllerReference(onmetalMachine, ignitionSecret, k8sClient.Scheme()); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("could not set ignition secret ownership: %s", err.Error()))
	}

	if err := k8sClient.Patch(ctx, ignitionSecret, client.Apply, fieldOwner, client.ForceOwnership); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("error applying ignition secret: %s", err.Error()))
	}

	return onmetalMachine, nil
}

func getIgnitionKeyOrDefault(key string) string {
	if key == "" {
		return computev1alpha1.DefaultIgnitionKey
	}
	return key
}

func validateProviderSpecAndSecret(class *machinev1alpha1.MachineClass, secret *corev1.Secret) (*apiv1alpha1.ProviderSpec, error) {
	if class == nil {
		return nil, status.Error(codes.Internal, "MachineClass in ProviderSpec is not set")
	}

	var providerSpec *apiv1alpha1.ProviderSpec
	if err := json.Unmarshal(class.ProviderSpec.Raw, &providerSpec); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	validationErr := validation.ValidateProviderSpec(providerSpec, secret, field.NewPath("providerSpec"))
	if validationErr.ToAggregate() != nil && len(validationErr.ToAggregate().Errors()) > 0 {
		err := fmt.Errorf("failed to validate provider spec: %v", validationErr.ToAggregate().Errors())
		klog.V(2).Infof("Validation of OnMetalMachineClass '%s' failed: %w", class.Name, err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	return providerSpec, nil
}
