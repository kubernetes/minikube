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

## Usage

Start a cluster using the hyperkit driver:

```shell
minikube start --driver=hyperkit
```

To make hyperkit the default driver:

```shell
minikube config set driver hyperkit
```
