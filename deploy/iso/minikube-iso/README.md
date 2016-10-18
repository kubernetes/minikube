## minikube ISO image

This includes the configuration for an alternative bootable ISO image meant to be used in conjection with minikube.

It includes:
- systemd as the init system
- rkt
- docker

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

### Build instructions

```
$ git clone https://github.com/kubernetes/minikube
$ cd minikube
$ make minikube-iso
```

The bootable ISO image will be available in `out/buildroot/output/images/rootfs.iso9660`.

### Testing local minikube changes

To test a local build of minikube, include a `kubernetes-version` flag with a path to the `localkube` output from your source build directory:

```
$ cd minikube
$ ./out/minikube start \
    --container-runtime=rkt \
    --network-plugin=cni \
    --kubernetes-version=file://$HOME/minikube/src/k8s.io/minikube/out/localkube \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.5/minikube-v0.0.5.iso
```

### Testing local minikube-iso changes

To test a local build of minikube-iso, start a web server (i.e. Caddy) to serve the ISO image, and start `minikube` with an `--iso-url` pointing to localhost:

```
$ cd $HOME/src/minikube/src/k8s.io/minikube
$ cd ./out/buildroot/output/images
$ caddy browse "log stdout"
Activating privacy features... done.
http://:2015
```

In another terminal:

```
$ minikube start --iso-url=http://localhost:2015/rootfs.iso9660
```

### Buildroot configuration

To change the buildroot configuration, execute:

```
$ cd out/buildroot
$ make menuconfig
$ make
```

To change the kernel configuration, execute:

```
$ cd out/buildroot
$ make linux-menuconfig
$ make
```

The last commands copies changes made to the kernel configuration to the minikube-iso defconfig.

### Saving buildroot/kernel configuration changes

To save any buildroot configuration changes made with `make menuconfig`, execute:

```
$ cd out/buildroot
$ make savedefconfig
```

The changes will be reflected in the `minikube-iso/configs/minikube_defconfig` file.

```
$ git stat
## master
 M deploy/iso/minikube-iso/configs/minikube_defconfig
```

To save any kernel configuration changes made with `make linux-menuconfig`, execute:

```
$ cd out/buildroot
$ make linux-savedefconfig
$ cp output/build/linux-4.7.2/defconfig \
    ../../deploy/iso/minikube-iso/board/coreos/minikube/linux-4.7_defconfig
```

The changes will be reflected in the `deploy/iso/minikube-iso/configs/minikube_defconfig` file.
