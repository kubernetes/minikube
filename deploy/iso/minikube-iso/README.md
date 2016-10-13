## minikube ISO image

This includes the configuration for an alternative bootable ISO image meant to be used in conjection with https://github.com/kubernetes/minikube.

It includes:
- systemd as the init system
- rkt
- docker

**Note**: This is currently intended to be a stop-gap solution. In the middleterm this is meant to be replaced by a "slim" version of a bootable CoreOS image.


## Quickstart

To use this ISO image, use the `--iso-url` flag in minikube:

```
$ minikube start \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.5/minikube-v0.0.5.iso
```

To test the minikube rkt container runtime support, make sure you have minikube v0.10 or later, and execute:

```
$ minikube start \
    --container-runtime=rkt \
    --network-plugin=cni \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.5/minikube-v0.0.5.iso
```

Note that the above statement includes `--network-plugin=cni` which is the recommended way of starting rtk+Kubernetes.

## Configurations

The following configurations have been tested:

* OSX
  * Virtualbox
  * VMware Fusion

* Linux
  * Virtualbox
  * KVM

The following configurations are known to have issues currently:

* OSX
  * xhyve (https://github.com/coreos/minikube-iso/issues/17)

## Hacking

To test a locally-built version of the minikube master branch, include a `kubernetes-version` flag with a path to the `localkube` output from your source build directory:

```
$ cd $HOME/src/minikube/src/k8s.io/minikube
$ ./out/minikube start \
    --container-runtime=rkt \
    --network-plugin=cni \
    --kubernetes-version=file://$HOME/minikube/src/k8s.io/minikube/out/localkube \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.3/minikube-v0.0.3.iso
```

### Build instructions
```
$ cd $HOME
$ git clone https://github.com/coreos/minikube-iso
$ git clone https://github.com/buildroot/buildroot
$ cd buildroot
$ git checkout 2016.08
$ make BR2_EXTERNAL=../minikube-iso minikube_defconfig
$ make
```

The bootable ISO image will be available in `output/images/rootfs.iso9660`.

### Buildroot configuration

To change the buildroot configuration, execute:

```
$ cd buildroot
$ make menuconfig
$ make
```

To change the kernel configuration, execute:

```
$ cd buildroot
$ make linux-menuconfig
$ make
```

The last commands copies changes made to the kernel configuration to the minikube-iso defconfig.

### Saving buildroot/kernel configuration changes

To save any buildroot configuration changes made with `make menuconfig`, execute:

```
$ cd buildroot
$ make savedefconfig
```

The changes will be reflected in the `minikube-iso/configs/minikube_defconfig` file.

```
$ cd minikube-iso
$ git stat
## master
 M configs/minikube_defconfig
```

To save any kernel configuration changes made with `make linux-menuconfig`, execute:

```
$ cd buildroot
$ make linux-savedefconfig
$ cp output/build/linux-4.7.2/defconfig \
    ../minikube-iso/board/coreos/minikube/linux-4.7_defconfig
```

The changes will be reflected in the `minikube-iso/configs/minikube_defconfig` file.
