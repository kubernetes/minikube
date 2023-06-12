---
title: "Drivers"
date: 2019-07-31
weight: 4
description: >
  How to create a new VM Driver
---

This document is written for contributors who are familiar with minikube, who would like to add support for a new VM driver.

minikube relies on docker-machine drivers to manage machines. This document discusses how to modify minikube, so that this driver may be used by `minikube start --driver=<new_driver>`.

## Creating a new driver

See [machine-drivers](https://github.com/machine-drivers) , the fork where all new docker-machine drivers are located.

## Builtin vs External Drivers

Most drivers are built-in: they are included into minikube as a code dependency, so no further
installation is required. There are two primary cases you may want to use an external driver:

- The driver has a code dependency which minikube should not rely on due to platform incompatibilities (kvm2) or licensing
- The driver needs to run with elevated permissions (hyperkit)

External drivers are instantiated by executing a command `docker-machine-driver-<name>`, which begins an RPC server which minikube will talk to.

### Integrating a driver

The integration process is effectively 3 steps.

1. Create a driver shim within `k8s.io/minikube/pkg/minikube/drivers`
   - Add Go build tag for the supported operating systems
   - Define the driver metadata to register in `DriverDef`
2. Add import in `pkg/minikube/cluster/default_drivers.go` so that the driver may be included by the minikube build process.

### The driver shim

The primary duty of the driver shim is to register a VM driver with minikube, and translate minikube VM hardware configuration into a format that the driver understands.

### Registering your driver

The docs on registry are available here: <https://pkg.go.dev/k8s.io/minikube/pkg/minikube/registry>

[DriverDef](https://pkg.go.dev/k8s.io/minikube/pkg/minikube/registry#DriverDef) is the main
struct to define a driver metadata. Essentially, you need to define 4 things at most, which is
pretty simple once you understand your driver well:

- Name: unique name of the driver, it will be used as the unique ID in registry and as
`--driver` option in minikube command

- Builtin: `true` if the driver should be builtin to minikube (preferred). `false` otherwise.

- ConfigCreator: how to translate a minikube config to driver config. The driver config will be persistent
on your `$USER/.minikube` directory. Most likely the driver config is the driver itself.

- DriverCreator: Only needed when driver is builtin, to instantiate the driver instance.

Any Questions: please ping your friend [@anfernee](https://github.com/anfernee) or the #minikube Slack channel.
