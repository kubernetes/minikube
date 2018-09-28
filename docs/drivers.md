# Driver plugin installation

Minikube uses Docker Machine to manage the Kubernetes VM so it benefits from the
driver plugin architecture that Docker Machine uses to provide a consistent way to
manage various VM providers. Minikube embeds VirtualBox and VMware Fusion drivers
so there are no additional steps to use them. However, other drivers require an
extra binary to be present in the host PATH.

The following drivers currently require driver plugin binaries to be present in
the host PATH:

* [KVM2](#kvm2-driver)
* [KVM](#kvm-driver)
* [Hyperkit](#hyperkit-driver)
* [xhyve](#xhyve-driver)
* [HyperV](#hyperv-driver)

#### KVM2 driver

The KVM2 driver is intended to replace KVM driver.
The KVM2 driver is maintained by the minikube team, and is built, tested and released with minikube.

To install the KVM2 driver, first install and configure the prereqs:

```shell
# Install libvirt and qemu-kvm on your system, e.g.
# Debian/Ubuntu (for older Debian/Ubuntu versions, you may have to use libvirt-bin instead of libvirt-clients and libvirt-daemon-system)
$ sudo apt install libvirt-clients libvirt-daemon-system qemu-kvm
# Fedora/CentOS/RHEL
$ sudo yum install libvirt-daemon-kvm qemu-kvm

# Add yourself to the libvirt group so you don't need to sudo
# NOTE: For older Debian/Ubuntu versions change the group to `libvirtd`
$ sudo usermod -a -G libvirt $(whoami)

# Update your current session for the group change to take effect
# NOTE: For older Debian/Ubuntu versions change the group to `libvirtd`
$ newgrp libvirt
```

Then install the driver itself:

```shell
curl -Lo docker-machine-driver-kvm2 https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2 \
&& chmod +x docker-machine-driver-kvm2 \
&& sudo cp docker-machine-driver-kvm2 /usr/local/bin/ \
&& rm docker-machine-driver-kvm2
```

To use the driver you would do:

```shell
minikube start --vm-driver kvm2
```

#### KVM driver

Minikube is currently tested against [`docker-machine-driver-kvm` v0.10.0](https://github.com/dhiltgen/docker-machine-kvm/releases).

After following the instructions on the KVM driver releases page, you need to make sure that have the necessary packages and permissions by following these instructions:
```shell
# Install libvirt and qemu-kvm on your system, e.g.
# Debian/Ubuntu (for older Debian/Ubuntu versions, you may have to use libvirt-bin instead of libvirt-clients and libvirt-daemon-system)
$ sudo apt install libvirt-clients libvirt-daemon-system qemu-kvm
# Fedora/CentOS/RHEL
$ sudo yum install libvirt-daemon-kvm qemu-kvm

# Add yourself to the libvirt group so you don't need to sudo
# NOTE: For older Debian/Ubuntu versions change the group to `libvirtd`
$ sudo usermod -a -G libvirt $(whoami)

# Update your current session for the group change to take effect
# NOTE: For older Debian/Ubuntu versions change the group to `libvirtd`
$ newgrp libvirt
```

To use the driver you would do:

```shell
minikube start --vm-driver kvm
```

#### Hyperkit driver

The Hyperkit driver will eventually replace the existing xhyve driver.
It is built from the minikube source tree, and uses [moby/hyperkit](http://github.com/moby/hyperkit) as a Go library.

To install the hyperkit driver:

```shell
curl -Lo docker-machine-driver-hyperkit https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
&& chmod +x docker-machine-driver-hyperkit \
&& sudo cp docker-machine-driver-hyperkit /usr/local/bin/ \
&& rm docker-machine-driver-hyperkit \
&& sudo chown root:wheel /usr/local/bin/docker-machine-driver-hyperkit \
&& sudo chmod u+s /usr/local/bin/docker-machine-driver-hyperkit
```

The hyperkit driver currently requires running as root to use the vmnet framework to setup networking.

If you encountered errors like `Could not find hyperkit executable`, you might need to install [Docker for Mac](https://store.docker.com/editions/community/docker-ce-desktop-mac)

If you are using [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) in your setup and cluster creation fails (stuck at kube-dns initialization) you might need to add `listen-address=192.168.64.1` to `dnsmasq.conf`.

*Note: If `dnsmasq.conf` contains `listen-address=127.0.0.1` kubernetes discovers dns at 127.0.0.1:53 and tries to use it using bridge ip address, but dnsmasq replies only to reqests from 127.0.0.1*

#### xhyve driver

From https://github.com/zchee/docker-machine-driver-xhyve#install:

```shell
$ brew install docker-machine-driver-xhyve

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
$ sudo chmod u+s $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
```

#### HyperV driver

Hyper-v users may need to create a new external network switch as described [here](https://docs.docker.com/machine/drivers/hyper-v/). This step may prevent a problem in which `minikube start` hangs indefinitely, unable to ssh into the minikube virtual machine. In this add, add the `--hyperv-virtual-switch=switch-name` argument to the `minikube start` command.

On some machines, having **dynamic memory management** turned on for the minikube VM can cause problems of unexpected and random restarts which manifests itself in simply losing the connection to the cluster, after which `minikube status` would simply state `stopped`. Machine restarts are caused due to following Hyper-V error: `The dynamic memory balancer could not add memory to the virtual machine 'minikube' because its configured maximum has been reached`. **Solution**: turned the dynamic memory management in hyper-v settings off (and allocate a fixed amount of memory to the machine).
