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

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"k8s.io/klog/v2"
)

// IronCoreCSIDriver is the CSI driver for ironcore provisioner
const (
	IroncoreCSIDriver = "csi.ironcore.dev"
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
