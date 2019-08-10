## Requirements

- [VirtualBox](https://www.virtualbox.org/wiki/Downloads) 5.2 or higher

## Usage

minikube currently uses VirtualBox by default, but it can also be explicitly set:

```shell
minikube start --vm-driver=virtualbox
```
To make virtualbox the default driver:

```shell
minikube config set vm-driver virtualbox
```
