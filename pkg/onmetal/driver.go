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
	"github.com/onmetal/machine-controller-manager-provider-onmetal/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fieldOwner        = client.FieldOwner("machine-controller-manager-provider-onmetal")
	labelKeyApp       = "app"
	labelKeyProvider  = "machine-provider"
	labelValueMachine = "machine"

	defaultIgnitionKey = "ignition.json"
)

type onmetalDriver struct {
	Schema *runtime.Scheme
}

// NewDriver returns a new Gardener on Metal driver object
func NewDriver(schema *runtime.Scheme) driver.Driver {
	return &onmetalDriver{
		Schema: schema,
	}
}

func (d *onmetalDriver) GenerateMachineClassForMigration(_ context.Context, _ *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	return &driver.GenerateMachineClassForMigrationResponse{}, nil
}

func (d *onmetalDriver) getOnmetalMachineKeyFromMachineRequest(req *driver.GetMachineStatusRequest, namespace string) types.NamespacedName {
	return types.NamespacedName{
		Namespace: namespace,
		Name:      req.Machine.Name,
	}
}

func (d *onmetalDriver) createK8sClient(secret *corev1.Secret) (client.Client, error) {
	kubeconfig, ok := secret.Data["kubeconfig"]
	if !ok {
		return nil, status.Error(codes.Internal, fmt.Sprintf("no kubeconfig found in machine secret %s", client.ObjectKeyFromObject(secret)))
	}

	cfg, err := restConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to load client config from machine secret %s: %v", client.ObjectKeyFromObject(secret), err))
	}
	k8sClient, err := client.New(cfg, client.Options{Scheme: d.Schema})
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create k8s client: %v", err))
	}
	return k8sClient, nil
}

func restConfigFromBytes(kubeconfig []byte) (*rest.Config, error) {
	clientCfg, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}
	cfg, err := clientCfg.ClientConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func getIgnitionNameForMachine(machineName string) string {
	return fmt.Sprintf("%s-%s", machineName, "ignition")
}

func getProviderIDForOnmetalMachine(onmetalMachine *computev1alpha1.Machine) string {
	return fmt.Sprintf("%s://%s/%s", v1alpha1.ProviderName, onmetalMachine.Namespace, onmetalMachine.Name)
}
