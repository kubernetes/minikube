## Requirements

- macOS 10.11+
- HyperKit

## HyperKit installation

* If Docker for Desktop is installed, you already have HyperKit
* Otherwise, if you have [Brew Package Manager](https://brew.sh/), run:

```shell
brew install hyperkit
```

* As a final alternative, you may [Install HyperKit from GitHub](https://github.com/moby/hyperkit)

## Driver installation

You must also download and install the latest minikube driver:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/docker-machine-driver-hyperkit \
&& sudo install -o root -m 4755 docker-machine-driver-hyperkit /usr/local/bin/
```

## Usage

Start a cluster using the hyperkit driver:

```shell
minikube start --vm-driver=hyperkit
```

To make hyperkit the default driver:

```shell
minikube config set vm-driver hyperkit
```
