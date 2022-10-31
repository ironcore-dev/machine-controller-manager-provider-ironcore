# machine-controller-manager-provider-onmetal

[![reuse compliant](https://reuse.software/badge/reuse-compliant.svg)](https://reuse.software/)

Out of tree (controller based) implementation for `onmetal` as a new provider.

## About
- The `onmetal` Out Of Tree provider implements the interface defined at [MCM OOT driver](https://github.com/gardener/machine-controller-manager/blob/master/pkg/util/provider/driver/driver.go).

## Fundamental Design Principles:

Following are the basic principles kept in mind while developing the external plugin.
* Communication between this Machine Controller (MC) and Machine Controller Manager (MCM) is achieved using the Kubernetes native declarative approach.
* Machine Controller (MC) behaves as the controller used to interact with the cloud provider `onmetal` and manage the resources corresponding to the machine objects.
* Machine Controller Manager (MCM) deals with higher level objects such as machine-set and machine-deployment objects.

## Support for a new provider

- Steps to be followed while implementing/testing a new provider are mentioned [here](https://github.com/gardener/machine-controller-manager/blob/master/docs/development/cp_support_new.md)

## Testing the Gardener on Metal OOT

TODO
