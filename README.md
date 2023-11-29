# machine-controller-manager-provider-ironcore

[![REUSE status](https://api.reuse.software/badge/github.com/ironcore-dev/machine-controller-manager-provider-ironcore)](https://api.reuse.software/info/github.com/ironcore-dev/machine-controller-manager-provider-ironcore)
[![Go Report Card](https://goreportcard.com/badge/github.com/ironcore-dev/machine-controller-manager-provider-ironcore)](https://goreportcard.com/report/github.com/ironcore-dev/machine-controller-manager-provider-ironcore)
[![GitHub License](https://img.shields.io/static/v1?label=License&message=Apache-2.0&color=blue)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://makeapullrequest.com)

Out of tree (controller based) implementation for `ironcore` as a new provider.

## About
- The `ironcore` Out Of Tree provider implements the interface defined at [MCM OOT driver](https://github.com/gardener/machine-controller-manager/blob/master/pkg/util/provider/driver/driver.go).

## Fundamental Design Principles:

Following are the basic principles kept in mind while developing the external plugin.
* Communication between this Machine Controller (MC) and Machine Controller Manager (MCM) is achieved using the Kubernetes native declarative approach.
* Machine Controller (MC) behaves as the controller used to interact with the cloud provider `ironcore` and manage the resources corresponding to the machine objects.
* Machine Controller Manager (MCM) deals with higher level objects such as machine-set and machine-deployment objects.

## Support for a new provider

- Steps to be followed while implementing/testing a new provider are mentioned [here](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md)

## Testing the Gardener on Metal OOT

1. Open terminal pointing to `$GOPATH/src/github.com/gardener`. Clone this repository.

2. Navigate to `$GOPATH/src/github.com/ironcore-dev/machine-controller-manager-provider-ironcore`.
    - In the `MAKEFILE` make sure `$TARGET_KUBECONFIG` points to the cluster where you wish to manage machines. `$CONTROL_NAMESPACE` represents the namespaces where MCM is looking for machine CR objects, and `$CONTROL_KUBECONFIG` points to the cluster which holds these machine CRs.
    - Run the machine controller (driver) using the command below.
        ```bash
        make start
        ```
3. On the second terminal pointing to `$GOPATH/src/github.com/gardener`,
    - Clone the [latest MCM code](https://github.com/gardener/machine-controller-manager)
        ```bash
        git clone git@github.com:gardener/machine-controller-manager.git
        ```
    - Navigate to the newly created directory.
        ```bash
        cd machine-controller-manager
        ```
    - Deploy the required CRDs from the machine-controller-manager repo,
        ```bash
        kubectl apply -f kubernetes/crds.yaml
        ```
    - Run the machine-controller-manager in the `cmi-client` branch
        ```bash
        make start
        ```
4. On the third terminal pointing to `$GOPATH/src/github.com/gardener/machine-controller-manager-provider-ironcore`
    - Fill in the object files given below and deploy them as described below.
    - Deploy the `machine-class`
        ```bash
        kubectl apply -f kubernetes/machine-class.yaml
        ```
    - Deploy the `kubernetes secret` if required.
        ```bash
        kubectl apply -f kubernetes/secret.yaml
        ```
    - Deploy the `machine` object and make sure it joins the cluster successfully.
        ```bash
        kubectl apply -f kubernetes/machine.yaml
        ```
    - Once machine joins, you can test by deploying a machine-deployment.
    - Deploy the `machine-deployment` object and make sure it joins the cluster successfully.
        ```bash
        kubectl apply -f kubernetes/machine-deployment.yaml
        ```
    - Make sure to delete both the `machine` and `machine-deployment` object after use.
        ```bash
        kubectl delete -f kubernetes/machine.yaml
        kubectl delete -f kubernetes/machine-deployment.yaml
