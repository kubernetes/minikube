# Vendor in drivers, remove libmachine RPC interface

### Goals

 * Simplify the local developer experience.  Minikube requires you to download a separate driver binary.  With this solution, only the minikube binary and the appropriate hypervisor are needed.
 * Remove hard docker dependency.  With the help of CoreOS, Minikube has a custom ISO to run either rkt or docker as the runtime for Kubernetes.  As a product of Docker, libmachine is inherently docker-specific and not aim at a generic runtime interface like CNI
 * More granular control over provisioning and machine configuration
 * Versioning control between minikube and drivers


Libmachine creates an RPC client/server for the different driver binaries to call methods on the driver interface.  We can bypass the RPC calls by directly vendoring in the drivers, all of which are written in golang.  Docker already does this for their “core” drivers (e.g. virtualbox) and in their new products Docker for Mac (i.e. xhyve).  The core drivers are determined by a whitelist [here](https://github.com/docker/machine/blob/master/libmachine/drivers/plugin/localbinary/plugin.go#L20-L23).

### Vendoring in the drivers

Vendoring in the drivers through Godep gives the ability to control versioning and releasing.  Not dependent on upstream driver maintainers to do a binary release.  

Drivers are small, kvm-driver (11 MiB), xhyve-driver (7 MiB), the rest are already “core” drivers in libmachine.  Even with adding the drivers, removing libmachine will most likely net a measurable net reduction in minikube binary size.

KVM and Xhyve require CGO.  Xhyve is nontrivial to cross compile on linux because it needs darwin headers.  In the event that cross compilation does not work, minikube for darwin must be compiled on our mac build slave.

Xhyve also has a dependency on OCaml for its vendored in docker/hyperkit.

Xhyve requires the driver to be installed with root:wheel but does not require sudo.

### Removal of Host type and API

#### Consolidation of configuration persistence boilerplate
Because of libmachine, we maintain two overlapping configuration files.


Minikube already reads and writes a config (~/.minikube/config/config.json) through the `minikube config view/set/get` commands. However, libmachine additionally writes its own serialized Host file (~/.minikube/machines/minikube/config.json).  The current minikube config may need some additional fields that contain information about the stateful machine (e.g. cert directory, IPAddress)  


We would need to reimplement api.Load and api.Save to write to one config file.

**Option #1** Continue to write a separate configuration file for the current machine state

**Option #2** Merge the minikube config and the machine config.  While it would simply a lot of the logic around reading and writing, it may conflate the state of the machine with startup options.  However, docker-machine only uses one config file for both driver state and options.

#### The Driver Interface

Currently, the driver interface has some methods defined such as SetConfigFromFlags() and GetCreateFlags() that won’t make sense when the code is called directly.  There are two options:


**Option #1** Use the driver interface as is and never call those functions.  At the expense of having vacuous implementations if we create new drivers, we maintain some compatibility with other libmachine code.

**Option #2** Create a new, restricted interface with a subset of functions.  This gives us more control going forward, but there is no clear gain unless we write our own drivers.

#### Loss of Compatibility with libmachine
Probably the greatest concern with this proposal is the loss of compatibility with libmachine packages.  Although the point of this proposal is the removal of this library, libmachine provides many useful functions - from provisioning different ISOs to bug fixes and boot2docker support.  As minikube moves to its own custom ISO, and its focus on Kubernetes rather than the docker ecosystem, this shouldn’t be too much of an issue.


If this proposal is implemented, there will still be references to libmachine.  One clear dependency will be the boot2docker provisioner and the drivers that are vendored into libmachine already.
