## Requirements

* Parallels Desktop 11.0.0+ Pro or Business edition

## Driver Installation

If the [Brew Package Manager](https://brew.sh/) is installed, run:

```shell
brew install docker-machine-parallels
```

Otherwise:

```shell
latest=$(curl -s https://api.github.com/repos/Parallels/docker-machine-parallels/releases/latest \
  | grep -o 'http.*docker-machine-driver-parallels' | head -n1)
curl -LO $latest
install docker-machine-driver-parallels /usr/local/bin/
```


## Usage

Start a cluster using the parallels driver:

```shell
minikube start --driver=parallels
```

To make parallels the default driver:

```shell
minikube config set driver parallels
```
