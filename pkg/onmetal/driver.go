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
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/api/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultIgnitionKey     = "ignition.json"
	ShootNameLabelKey      = "shoot-name"
	ShootNamespaceLabelKey = "shoot-namespace"
)

var (
	fieldOwner = client.FieldOwner("mcm.onmetal.de/field-owner")
)

type onmetalDriver struct {
	Schema           *runtime.Scheme
	OnmetelClient    client.Client
	OnmetalNamespace string
}

// NewDriver returns a new Gardener on Metal driver object
func NewDriver(c client.Client, namespace string) driver.Driver {
	return &onmetalDriver{
		OnmetelClient:    c,
		OnmetalNamespace: namespace,
	}
}

func (d *onmetalDriver) GenerateMachineClassForMigration(_ context.Context, _ *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	return &driver.GenerateMachineClassForMigrationResponse{}, nil
}

func getIgnitionNameForMachine(machineName string) string {
	return fmt.Sprintf("%s-%s", machineName, "ignition")
}

func getProviderIDForOnmetalMachine(onmetalMachine *computev1alpha1.Machine) string {
	return fmt.Sprintf("%s://%s/%s", v1alpha1.ProviderName, onmetalMachine.Namespace, onmetalMachine.Name)
}
