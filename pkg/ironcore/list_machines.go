// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	"context"
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	apiv1alpha1 "github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (d *ironcoreDriver) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine list request has been received for %q", req.MachineClass.Name)
	defer klog.V(3).Infof("Machine list request has been processed for %q", req.MachineClass.Name)

	providerSpec, err := validateProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("provider spec for requested provider '%s' is invalid: %v", req.MachineClass.Provider, err))
	}

	// Get ironcore machine list
	ironcoreMachineList := &computev1alpha1.MachineList{}
	matchingLabels := client.MatchingLabels{}
	for k, v := range providerSpec.Labels {
		matchingLabels[k] = v
	}
	if err := d.IroncoreClient.List(ctx, ironcoreMachineList, client.InNamespace(d.IroncoreNamespace), matchingLabels); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	//Creating machineList from ironcoreMachineList items
	machineList := make(map[string]string, len(ironcoreMachineList.Items))
	for _, machine := range ironcoreMachineList.Items {
		machineID := getProviderIDForIroncoreMachine(&machine)
		machineList[machineID] = machine.Name
	}

	return &driver.ListMachinesResponse{MachineList: machineList}, nil
}
