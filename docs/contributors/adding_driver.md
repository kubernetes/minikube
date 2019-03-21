# Adding new driver (Deprecated)

New drivers should be added into <https://github.com/machine-drivers>

Minikube relies on docker machine drivers to manage machines. This document talks about how to
add an existing docker machine driver into minikube registry, so that minikube can use the driver
by `minikube create --vm-driver=<new_driver>`. This document is not going to talk about how to
create a new docker machine driver.

## Understand your driver

First of all, before started, you need to understand your driver in terms of:

- Which operating system is your driver running on?
- Is your driver builtin the minikube binary or triggered through RPC?
- How to translate minikube config to driver config?
- If builtin, how to instantiate the driver instance?

Builtin basically means whether or not you need separate driver binary in your `$PATH` for minikube to
work. For instance, `hyperkit` is not builtin, because you need `docker-machine-driver-hyperkit` in your
`$PATH`. `vmwarefusion` is builtin, because you don't need anything.

## Understand registry

Registry is what minikube uses to register all the supported drivers. The driver author registers
their drivers in registry, and minikube runtime will look at the registry to find a driver and use the
driver metadata to determine what workflow to apply while those drivers are being used.

The godoc of registry is available here: <https://godoc.org/k8s.io/minikube/pkg/minikube/registry>

[DriverDef](https://godoc.org/k8s.io/minikube/pkg/minikube/registry#DriverDef) is the main
struct to define a driver metadata. Essentially, you need to define 4 things at most, which is
pretty simple once you understand your driver well:

- Name: unique name of the driver, it will be used as the unique ID in registry and as
`--vm-driver` option in minikube command

- Builtin: `true` if the driver is builtin minikube binary. `false` otherwise.

- ConfigCreator: how to translate a minikube config to driver config. The driver config will be persistent
on your `$USER/.minikube` directory. Most likely the driver config is the driver itself.

- DriverCreator: Only needed when driver is builtin, to instantiate the driver instance.

## An example

All drivers are located in `k8s.io/minikube/pkg/minikube/drivers`. Take `vmwarefusion` as an example:

```golang
// +build darwin

package vmwarefusion

import (
    "github.com/docker/machine/drivers/vmwarefusion"
    "github.com/docker/machine/libmachine/drivers"
    cfg "k8s.io/minikube/pkg/minikube/config"
    "k8s.io/minikube/pkg/minikube/constants"
    "k8s.io/minikube/pkg/minikube/registry"
)

func init() {
    registry.Register(registry.DriverDef{
        Name:          "vmwarefusion",
        Builtin:       true,
        ConfigCreator: createVMwareFusionHost,
        DriverCreator: func() drivers.Driver {
            return vmwarefusion.NewDriver("", "")
        },
    })
}

func createVMwareFusionHost(config cfg.MachineConfig) interface{} {
    d := vmwarefusion.NewDriver(cfg.GetMachineName(), constants.GetMinipath()).(*vmwarefusion.Driver)
    d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
    d.Memory = config.Memory
    d.CPU = config.CPUs
    d.DiskSize = config.DiskSize
    d.SSHPort = 22
    d.ISO = d.ResolveStorePath("boot2docker.iso")
    return d
}
```

- In init function, register a `DriverDef` in registry. Specify the metadata in the `DriverDef`. As mentioned
earlier, it's builtin, so you also need to specify `DriverCreator` to tell minikube how to create a `drivers.Driver`.
- Another important thing is `vmwarefusion` only runs on MacOS. You need to add a build tag on top so it only
runs on MacOS, so that the releases on Windows and Linux won't have this driver in registry.
- Last but not least, import the driver in `pkg/minikube/cluster/default_drivers.go` to include it in build.

## Summary

In summary, the process includes the following steps:

1. Add the driver under `k8s.io/minikube/pkg/minikube/drivers`
   - Add build tag for supported operating system
   - Define driver metadata in `DriverDef`
2. Add import in `pkg/minikube/cluster/default_drivers.go`

Any Questions: please ping your friend [@anfernee](https://github.com/anfernee)
