## minikube ISO image

This includes the configuration for an alternative bootable ISO image meant to be used in conjection with https://github.com/kubernetes/minikube.

It includes:
- systemd as the init system
- rkt
- docker

### Build instructions
```
$ cd $HOME
$ git clone https://github.com/coreos/minikube-iso
$ git clone https://github.com/buildroot/buildroot
$ cd buildroot
$ git checkout 2016.08-rc3
$ make BR2_EXTERNAL=../minikube-iso minikube_defconfig
$ make
```

The bootable ISO image will be available in `output/images/rootfs.iso9660`.

**Note**: This is currently intended to be a stop-gap solution. In the middleterm this is meant to be replaced by a "slim" version of a bootable CoreOS image.

## Quickstart

To use this ISO image, use the `--iso-url` flag in minikube:

```
$ minikube start --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.1/minikube-v0.0.1.iso
```

To test the minikube rkt container runtime support, make sure you have a locally built version of minikube including https://github.com/kubernetes/minikube/pull/511, and execute:

```
$ cd $HOME/src/minikube/src/k8s.io/minikube
$ ./out/mininikube start \
    --container-runtime=rkt \
    --kubernetes-version=file://$HOME/minikube/src/k8s.io/minikube/out/localkube \
    --iso-url=https://github.com/coreos/minikube-iso/releases/download/v0.0.1/minikube-v0.0.1.iso
```
