## Requirements

- macOS 10.11+
- HyperKit

## HyperKit installation

* If Docker for Desktop is installed, you already have HyperKit
* Otherwise minikube downloads HyperKit automatically to a suitable version if missing (on the host), there is no need to manually upgrade the driver

## Usage

Start a cluster using the hyperkit driver:

```shell
minikube start --driver=hyperkit
```

To make hyperkit the default driver:

```shell
minikube config set driver hyperkit
```
