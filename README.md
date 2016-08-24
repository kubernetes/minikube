## minikube ISO image

This includes the configuration for an alternative bootable ISO image meant to be used in conjection with https://github.com/kubernetes/minikube.

It includes:
- systemd as the init system
- rkt
- docker

### Build instructions
```
$ cd $HOME
$ git clone https://github.com/s-urbaniak/rktminikube
$ cd ..
$ git clone https://github.com/buildroot/buildroot
$ cd buildroot
$ git checkout 2016.08-rc2
$ make BR2_EXTERNAL=../rktminikube minikube_defconfig
$ make
```

The bootable ISO image will be available in `output/images/rootfs.iso9660`.

**Note**: This is currently intended to be a stop-gap solution. In the middleterm this is meant to be replaced by a "slim" version of a bootable CoreOS image.
