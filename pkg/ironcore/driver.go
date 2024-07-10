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
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	defaultIgnitionKey     = "ignition.json"
	ShootNameLabelKey      = "shoot-name"
	ShootNamespaceLabelKey = "shoot-namespace"
	DefaultCSIDriverName   = "csi.ironcore.dev"
)

var (
	fieldOwner = client.FieldOwner("mcm.ironcore.de/field-owner")
)

type ironcoreDriver struct {
	Schema            *runtime.Scheme
	IroncoreClient    client.Client
	IroncoreNamespace string
	CSIDriverName     string
}

func (d *ironcoreDriver) InitializeMachine(ctx context.Context, request *driver.InitializeMachineRequest) (*driver.InitializeMachineResponse, error) {
	return nil, status.Error(codes.Unimplemented, "IronCore Provider does not yet implement InitializeMachine")
}

// NewDriver returns a new Gardener ironcore driver object
func NewDriver(c client.Client, namespace, csiDriverName string) driver.Driver {
	return &ironcoreDriver{
		IroncoreClient:    c,
		IroncoreNamespace: namespace,
		CSIDriverName:     csiDriverName,
	}
}

func (d *ironcoreDriver) GenerateMachineClassForMigration(_ context.Context, _ *driver.GenerateMachineClassForMigrationRequest) (*driver.GenerateMachineClassForMigrationResponse, error) {
	return &driver.GenerateMachineClassForMigrationResponse{}, nil
}

func getIgnitionNameForMachine(machineName string) string {
	return fmt.Sprintf("%s-%s", machineName, "ignition")
}

func getProviderIDForIroncoreMachine(ironcoreMachine *computev1alpha1.Machine) string {
	return fmt.Sprintf("%s://%s/%s", v1alpha1.ProviderName, ironcoreMachine.Namespace, ironcoreMachine.Name)
}
