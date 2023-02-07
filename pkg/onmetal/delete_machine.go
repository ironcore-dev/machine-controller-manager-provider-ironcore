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
	"time"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/codes"
	"github.com/gardener/machine-controller-manager/pkg/util/provider/machinecodes/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1alpha1 "github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
)

// DeleteMachine handles a machine deletion request and also deletes ignitionSecret associated with it
func (d *onmetalDriver) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	if isEmptyDeleteRequest(req) {
		return nil, status.Error(codes.InvalidArgument, "received empty request")
	}
	if req.MachineClass.Provider != apiv1alpha1.ProviderName {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("requested provider '%s' is not suppored by the driver '%s'", req.MachineClass.Provider, apiv1alpha1.ProviderName))
	}

	klog.V(3).Infof("Machine deletion request has been received for %q", req.Machine.Name)
	defer klog.V(3).Infof("Machine deletion request has been processed for %q", req.Machine.Name)

	// Get namespace from machine secret
	namespace, ok := req.Secret.Data["namespace"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to find namespace in machine secret %s", client.ObjectKeyFromObject(req.Secret)))
	}

	// Create k8s client for the user provided machine secret. This client will be used
	// to create the resources in the user provided namespace.
	k8sClient, err := d.createK8sClient(req.Secret)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create k8s client for machine secret %s: %v", client.ObjectKeyFromObject(req.Secret), err))
	}

	ignitionSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getIgnitionNameForMachine(req.Machine.Name),
			Namespace: string(namespace),
		},
	}

	if err := k8sClient.Delete(ctx, ignitionSecret); client.IgnoreNotFound(err) != nil {
		// Unknown leads to short retry in machine controller
		return nil, status.Error(codes.Unknown, fmt.Sprintf("error deleting ignition secret: %s", err.Error()))
	}

	onmetalMachine := &computev1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Machine.Name,
			Namespace: string(namespace),
		},
	}

	if err := k8sClient.Delete(ctx, onmetalMachine); err != nil {
		if !apierrors.IsNotFound(err) {
			// Unknown leads to short retry in machine controller
			return nil, status.Error(codes.Unknown, fmt.Sprintf("error deleting pod: %s", err.Error()))
		}
		return nil, status.Error(codes.NotFound, err.Error())
	}

	// Actively wait until the onmetal machine is deleted since the extension contract in machine-controller-manager expects drivers to
	// do so. If we would not wait until the onmetal machine is gone it might happen that the kubelet could re-register the Node
	// object even after it was already deleted by machine-controller-manager.
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	if err := wait.PollUntilWithContext(timeoutCtx, 5*time.Second, func(ctx context.Context) (bool, error) {
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(onmetalMachine), onmetalMachine); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			// Unknown leads to short retry in machine controller
			return false, status.Error(codes.Unknown, err.Error())
		}
		return false, nil
	}); err != nil {
		// will be retried with short retry by machine controller
		return nil, status.Error(codes.DeadlineExceeded, err.Error())
	}

	return &driver.DeleteMachineResponse{}, nil
}

func isEmptyDeleteRequest(req *driver.DeleteMachineRequest) bool {
	return req == nil || req.MachineClass == nil || req.Machine == nil || req.Secret == nil
}
