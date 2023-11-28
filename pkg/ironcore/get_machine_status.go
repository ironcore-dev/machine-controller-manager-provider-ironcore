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

package ironcore

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	apiv1alpha1 "github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetMachineStatus handles a machine get status request
func (d *ironcoreDriver) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	if isEmptyMachineStatusRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine status request has been received for %q", req.Machine.Name)
	defer klog.V(3).Infof("Machine status request has been processed for %q", req.Machine.Name)

	// Fetch machine key from machine request
	ironcoreMachineKey := client.ObjectKey{
		Namespace: d.IroncoreNamespace,
		Name:      req.Machine.Name,
	}
	ironcoreMachine := &computev1alpha1.Machine{}
	if err := d.IroncoreClient.Get(ctx, ironcoreMachineKey, ironcoreMachine); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &driver.GetMachineStatusResponse{
		ProviderID: getProviderIDForIroncoreMachine(ironcoreMachine),
		NodeName:   ironcoreMachine.Name,
	}, nil
}

func isEmptyMachineStatusRequest(req *driver.GetMachineStatusRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}
