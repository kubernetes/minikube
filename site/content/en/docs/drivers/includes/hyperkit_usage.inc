## Requirements

- macOS 10.11+, darwin/x86_64
  - Note: no support for arm64.
- HyperKit

## HyperKit installation

* With [Brew Package Manager](https://formulae.brew.sh/formula/hyperkit#default), run:

```shell
brew install hyperkit
```

* Else, [Install HyperKit from GitHub](https://github.com/moby/hyperkit).

## Usage

Start a cluster using the hyperkit driver:

```shell
minikube start --driver=hyperkit
```

To make hyperkit the default driver:

```shell
minikube config set driver hyperkit
```
