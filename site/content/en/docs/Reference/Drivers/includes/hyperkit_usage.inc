## Requirements

- macOS 10.11+
- HyperKit

## Installing dependencies

If Docker for Desktop is installed, HyperKit is already installed.

Otherwise, if the [Brew Package Manager](https://brew.sh/) is installed:

```shell
brew install hyperkit
```

As a final alternative, you can [Install HyperKit from GitHub](https://github.com/moby/hyperkit)

## Driver Installation

Download and install the latest minikube driver:

```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/
docker-machine-driver-hyperkit \
&& sudo install -o root -m 4755 docker-machine-driver-hyperkit /usr/local/bin/
```