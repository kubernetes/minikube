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
curl -LO $(curl -s $r/releases/latest | grep -o 'http.*darwin_amd64' | head -n1) \
 && install docker-machine-driver-vmware_darwin_amd64 \
  /usr/local/bin/docker-machine-driver-vmware
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
