// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package ironcore

import (
	"context"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"k8s.io/klog/v2"
)

// IronCoreCSIDriver is the CSI driver for ironcore provisioner
const (
	IroncoreCSIDriver = "ironcore-csi-driver"
)

func (d *ironcoreDriver) GetVolumeIDs(_ context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	klog.V(2).Infof("Get VolumeIDs request has been received")
	klog.V(4).Infof("PVSpecList = %q", req.PVSpecs)

	var volumeIDs []string
	for _, pvSpec := range req.PVSpecs {
		if pvSpec.CSI != nil && pvSpec.CSI.Driver == IroncoreCSIDriver && pvSpec.CSI.VolumeHandle != "" {
			volumeID := pvSpec.CSI.VolumeHandle
			volumeIDs = append(volumeIDs, volumeID)
		}
	}

	klog.V(2).Infof("Get VolumeIDs request has been processed successfully (%d/%d).", len(volumeIDs), len(req.PVSpecs))
	klog.V(4).Infof("VolumneIDs: %v", volumeIDs)

	response := &driver.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}
	return response, nil
}
