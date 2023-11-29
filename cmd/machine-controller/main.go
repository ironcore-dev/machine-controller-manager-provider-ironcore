// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"github.com/gardener/machine-controller-manager/pkg/client/clientset/versioned/scheme"
	_ "github.com/gardener/machine-controller-manager/pkg/util/client/metrics/prometheus" // for client metric registration
	"github.com/gardener/machine-controller-manager/pkg/util/provider/app"
	mcmoptions "github.com/gardener/machine-controller-manager/pkg/util/provider/app/options"
	_ "github.com/gardener/machine-controller-manager/pkg/util/reflector/prometheus" // for reflector metric registration
	_ "github.com/gardener/machine-controller-manager/pkg/util/workqueue/prometheus" // for workqueue metric registration
	computev1alpha1 "github.com/ironcore-dev/ironcore/api/compute/v1alpha1"
	"github.com/ironcore-dev/machine-controller-manager-provider-ironcore/pkg/ironcore"
	"github.com/spf13/pflag"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	logsv1 "k8s.io/component-base/logs/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	IroncoreKubeconfigPath string
	CSIDriverName          string
)

func main() {
	s := mcmoptions.NewMCServer()
	s.AddFlags(pflag.CommandLine)

	options := logs.NewOptions()
	logs.AddFlags(pflag.CommandLine)
	AddExtraFlags(pflag.CommandLine)

	flag.InitFlags()
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := logsv1.ValidateAndApply(options, nil); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	ironcoreClient, namespace, err := getIroncoreClientAndNamespace()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	drv := ironcore.NewDriver(ironcoreClient, namespace, CSIDriverName)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := app.Run(s, drv); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func getIroncoreClientAndNamespace() (client.Client, string, error) {
	s := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(s))
	utilruntime.Must(computev1alpha1.AddToScheme(s))
	utilruntime.Must(corev1.AddToScheme(s))

	ironcoreKubeconfigData, err := os.ReadFile(IroncoreKubeconfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read ironcore kubeconfig %s: %w", IroncoreKubeconfigPath, err)
	}
	ironcoreKubeconfig, err := clientcmd.Load(ironcoreKubeconfigData)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read ironcore cluster kubeconfig: %w", err)
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*ironcoreKubeconfig, nil)
	if err != nil {
		return nil, "", fmt.Errorf("unable to serialize ironcore cluster kubeconfig: %w", err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get ironcore cluster rest config: %w", err)
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get namespace from ironcore kubeconfig: %w", err)
	}
	if namespace == "" {
		return nil, "", fmt.Errorf("got a empty namespace from ironcore kubeconfig")
	}
	client, err := client.New(restConfig, client.Options{Scheme: s})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create client: %w", err)
	}
	return client, namespace, nil
}

func AddExtraFlags(fs *pflag.FlagSet) {
	fs.StringVar(&IroncoreKubeconfigPath, "ironcore-kubeconfig", "", "Path to the ironcore kubeconfig.")
	fs.StringVar(&CSIDriverName, "csi-driver-name", ironcore.DefaultCSIDriverName, "CSI driver name used to determine the volumes for a Node.")
}
