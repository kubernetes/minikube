## Requirements

* VMware Fusion

## Driver Installation

If the [Brew Package Manager](https://brew.sh/) is installed, run:

```shell
brew install docker-machine-driver-vmware
```

Otherwise:

```shell
r=https://api.github.com/repos/machine-drivers/docker-machine-driver-vmware
d=docker-machine-driver-vmware_darwin_amd64
u=$(curl -s $r/releases/latest | grep -o 'http.*Darwin_amd64.tar.gz' | head -n1)
mkdir $d \
 && (cd $d && curl -L $u > $d.tar.gz && tar -xf $d.tar.gz) \
 && install $d/docker-machine-driver-vmware /usr/local/bin/docker-machine-driver-vmware \
 && rm -rf $d
```

## Usage

Start a cluster using the vmware driver:

```shell
minikube start --driver=vmware
```
To make vmware the default driver:

```shell
minikube config set driver vmware
```
