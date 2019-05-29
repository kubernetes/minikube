# VM Driver plugin installation

Minikube uses Docker Machine to manage the Kubernetes VM so it benefits from the
driver plugin architecture that Docker Machine uses to provide a consistent way to
manage various VM providers. Minikube embeds VirtualBox and VMware Fusion drivers
so there are no additional steps to use them. However, other drivers require an
extra binary to be present in the host PATH.

The following drivers currently require driver plugin binaries to be present in
the host PATH:

* [KVM2](#kvm2-driver)
* [Hyperkit](#hyperkit-driver)
* [HyperV](#hyperv-driver)
* [VMware](#vmware-unified-driver)

## KVM2 driver

To install the KVM2 driver, first install and configure the prerequisites, namely libvirt 1.3.1 or higher, and qemu-kvm:

* Debian or Ubuntu 18.x: `sudo apt install libvirt-clients libvirt-daemon-system qemu-kvm`
* Ubuntu 16.x or older: `sudo apt install libvirt-bin libvirt-daemon-system qemu-kvm`
* Fedora/CentOS/RHEL: `sudo yum install libvirt-daemon-kvm qemu-kvm`

Check your installed virsh version:

`virsh --version`

If your version of virsh is newer than 1.3.1 (January 2016), you may download our pre-built driver:


```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-kvm2 \
  && sudo install docker-machine-driver-kvm2 /usr/local/bin/
```

If your version of virsh is older than 1.3.1 (Januarry 2016), you may build your own driver binary if you have go 1.12+ installed.

```shell
$ sudo apt install libvirt-dev
$ git clone https://github.com/kubernetes/minikube.git
$ cd minikube
$ make out/docker-machine-driver-kvm2
$ sudo install out/docker-machine-driver-kvm2 /usr/local/bin
```

To finish the kvm installation, start and verify the `libvirtd` service

```shell
sudo systemctl enable libvirtd.service
sudo systemctl start libvirtd.service
sudo systemctl status libvirtd.service
```

Add your user to `libvirt` group (older distributions may use `libvirtd` instead)

```shell
sudo usermod -a -G libvirt $(whoami)
```

Join the `libvirt` group with your current shell session:

```shell
newgrp libvirt
```

To use the kvm2 driver:

```shell
minikube start --vm-driver kvm2
```

or, to use kvm2 as a default driver for `minikube start`:

```shell
minikube config set vm-driver kvm2
```

## Hyperkit driver

Install the [hyperkit](http://github.com/moby/hyperkit) VM manager using [brew](https://brew.sh):

```shell
brew install hyperkit
```

Then install the most recent version of minikube's fork of the hyperkit driver:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
&& sudo install -o root -g wheel -m 4755 docker-machine-driver-hyperkit /usr/local/bin/
```

If you are using [dnsmasq](http://www.thekelleys.org.uk/dnsmasq/doc.html) in your setup and cluster creation fails (stuck at kube-dns initialization) you might need to add `listen-address=192.168.64.1` to `dnsmasq.conf`.

*Note: If `dnsmasq.conf` contains `listen-address=127.0.0.1` kubernetes discovers dns at 127.0.0.1:53 and tries to use it using bridge ip address, but dnsmasq replies only to requests from 127.0.0.1*

To use the driver:

```shell
minikube start --vm-driver hyperkit
```

or, to use hyperkit as a default driver for minikube:

```shell
minikube config set vm-driver hyperkit
```

## HyperV driver

Hyper-v users may need to create a new external network switch as described [here](https://docs.docker.com/machine/drivers/hyper-v/). This step may prevent a problem in which `minikube start` hangs indefinitely, unable to ssh into the minikube virtual machine. In this add, add the `--hyperv-virtual-switch=switch-name` argument to the `minikube start` command.

On some machines, having **dynamic memory management** turned on for the minikube VM can cause problems of unexpected and random restarts which manifests itself in simply losing the connection to the cluster, after which `minikube status` would simply state `stopped`. Machine restarts are caused due to following Hyper-V error: `The dynamic memory balancer could not add memory to the virtual machine 'minikube' because its configured maximum has been reached`. **Solution**: turned the dynamic memory management in hyper-v settings off (and allocate a fixed amount of memory to the machine).

To use the driver:

```shell
minikube start --vm-driver hyperv --hyperv-virtual-switch=switch-name
```

or, to use hyperv as a default driver:

```shell
minikube config set vm-driver hyperv && minikube config set hyperv-virtual-switch switch-name
```

and run minikube as usual:

```shell
minikube start
```

## VMware unified driver

The VMware unified driver will eventually replace the existing vmwarefusion driver.
The new unified driver supports both VMware Fusion (on macOS) and VMware Workstation (on Linux and Windows)

To install the vmware unified driver, head over at <https://github.com/machine-drivers/docker-machine-driver-vmware/releases> and download the release for your operating system.

The driver must be:

1. Stored in `$PATH`
2. Named `docker-machine-driver-vmware`
3. Executable (`chmod +x` on UNIX based platforms)

If you're running on macOS with Fusion, this is an easy way install the driver:

```shell
export LATEST_VERSION=$(curl -L -s -H 'Accept: application/json' https://github.com/machine-drivers/docker-machine-driver-vmware/releases/latest | sed -e 's/.*"tag_name":"\([^"]*\)".*/\1/') \
&& curl -L -o docker-machine-driver-vmware https://github.com/machine-drivers/docker-machine-driver-vmware/releases/download/$LATEST_VERSION/docker-machine-driver-vmware_darwin_amd64 \
&& chmod +x docker-machine-driver-vmware \
&& mv docker-machine-driver-vmware /usr/local/bin/
```

To use the driver:

```shell
minikube start --vm-driver vmware
```

or, to use vmware unified driver as a default driver:

```shell
minikube config set vm-driver vmware
```

and run minikube as usual:

```shell
minikube start
```

# Troubleshooting

minikube is currently unable to display the error message received back from the VM driver. Users can however reveal the error by passing `--alsologtostderr -v=8` to `minikube start`. For instance:

```shell
minikube start --vm-driver=kvm2 --alsologtostderr -v=8
```

Output:

```
Found binary path at /usr/local/bin/docker-machine-driver-kvm2
Launching plugin server for driver kvm2
Error starting plugin binary: fork/exec /usr/local/bin/docker-machine-driver-kvm2: exec format error   
```
