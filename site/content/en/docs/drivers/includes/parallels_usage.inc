## Requirements

* Parallels Desktop 11.0.0+ Pro or Business edition

## Driver Installation

If the [Brew Package Manager](https://brew.sh/) is installed, run:

```shell
brew install docker-machine-parallels
```

Otherwise:

```shell
r=https://api.github.com/repos/Parallels/docker-machine-parallels
curl -LO $(curl -s $r/releases/latest | grep -o 'http.*parallels' | head -n1) \
 && install docker-machine-driver-parallels /usr/local/bin/
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
