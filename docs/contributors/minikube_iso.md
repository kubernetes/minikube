# minikube ISO image

This includes the configuration for an alternative bootable ISO image meant to be used in conjunction with minikube.

It includes:

- systemd as the init system
- docker
- CRI-O

## Hacking

### Requirements

* Linux

```shell
sudo apt-get install build-essential gnupg2 p7zip-full git wget cpio python \
    unzip bc gcc-multilib automake libtool locales
```

Either import your private key or generate a sign-only key using `gpg2 --gen-key`.
Also be sure to have an UTF-8 locale set up in order to build the ISO.

### Build instructions

```shell
$ git clone https://github.com/kubernetes/minikube.git
$ cd minikube
$ make buildroot-image
$ make out/minikube.iso
```

The build will occur inside a docker container. If you want to do this on
baremetal, replace `make out/minikube.iso` with `IN_DOCKER=1 make out/minikube.iso`.
The bootable ISO image will be available in `out/minikube.iso`.

### Testing local minikube-iso changes

```shell
$ ./out/minikube start --iso-url=file://$(pwd)/out/minikube.iso
```

### Buildroot configuration

To change the buildroot configuration, execute:

```shell
$ cd out/buildroot
$ make menuconfig
$ make
```

To save any buildroot configuration changes made with `make menuconfig`, execute:

```shell
$ cd out/buildroot
$ make savedefconfig
```

The changes will be reflected in the `minikube-iso/configs/minikube_defconfig` file.

```shell
$ git status
## master
 M deploy/iso/minikube-iso/configs/minikube_defconfig
```

### Saving buildroot/kernel configuration changes

To make any kernel configuration changes and save them, execute:

```shell
$ make linux-menuconfig
```

This will open the kernel configuration menu, and then save your changes to our
iso directory after they've been selected.
