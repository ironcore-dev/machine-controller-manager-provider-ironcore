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
	"github.com/onmetal/machine-controller-manager-provider-onmetal/pkg/onmetal"
	computev1alpha1 "github.com/onmetal/onmetal-api/api/compute/v1alpha1"
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
	OnmetalKubeconfigPath string
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

	onmetalClient, namespace, err := getOnmetalClientAndNamespace()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	drv := onmetal.NewDriver(onmetalClient, namespace)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	if err := app.Run(s, drv); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func getOnmetalClientAndNamespace() (client.Client, string, error) {
	s := runtime.NewScheme()
	utilruntime.Must(scheme.AddToScheme(s))
	utilruntime.Must(computev1alpha1.AddToScheme(s))
	utilruntime.Must(corev1.AddToScheme(s))

	onmetalKubeconfigData, err := os.ReadFile(OnmetalKubeconfigPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read onmetal kubeconfig %s: %w", OnmetalKubeconfigPath, err)
	}
	onmetalKubeconfig, err := clientcmd.Load(onmetalKubeconfigData)
	if err != nil {
		return nil, "", fmt.Errorf("unable to read onmetal cluster kubeconfig: %w", err)
	}
	clientConfig := clientcmd.NewDefaultClientConfig(*onmetalKubeconfig, nil)
	if err != nil {
		return nil, "", fmt.Errorf("unable to serialize onmetal cluster kubeconfig: %w", err)
	}
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, "", fmt.Errorf("unable to get onmetal cluster rest config: %w", err)
	}
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, "", fmt.Errorf("failed to get namespace from onmetal kubeconfig: %w", err)
	}
	if namespace == "" {
		return nil, "", fmt.Errorf("got a empty namespace from onmetal kubeconfig")
	}
	client, err := client.New(restConfig, client.Options{Scheme: s})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create client: %w", err)
	}
	return client, namespace, nil
}

func AddExtraFlags(fs *pflag.FlagSet) {
	fs.StringVar(&OnmetalKubeconfigPath, "onmetal-kubeconfig", "", "Path to the onmetal kubeconfig.")
}
