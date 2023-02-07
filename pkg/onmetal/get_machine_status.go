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
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	apiv1alpha1 "github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *onmetalDriver) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	if isEmptyMachineStatusRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine status request has been received for %q", req.Machine.Name)
	defer klog.V(3).Infof("Machine status request has been processed for %q", req.Machine.Name)

	// Get namespace from machine secret
	namespace, ok := req.Secret.Data["namespace"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find namespace is machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}

	// Create k8s client for the user provided machine secret. This client will be used
	// to create the resources in the user provided namespace.
	k8sClient, err := d.createK8sClient(req.Secret)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create k8s client for machine secret %s: %v", client.ObjectKeyFromObject(req.Secret), err))
	}

	// Fetch machine key from machine request
	onmetalMachineKey := client.ObjectKey{
		Namespace: string(namespace),
		Name:      req.Machine.Name,
	}
	onmetalMachine := &computev1alpha1.Machine{}
	if err := k8sClient.Get(ctx, onmetalMachineKey, onmetalMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &driver.GetMachineStatusResponse{
		ProviderID: getProviderIDForOnmetalMachine(onmetalMachine),
		NodeName:   onmetalMachine.Name,
	}, nil
}

func isEmptyMachineStatusRequest(req *driver.GetMachineStatusRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}
