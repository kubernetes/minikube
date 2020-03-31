## Install Docker

- [Docker Desktop](https://hub.docker.com/search?q=&type=edition&offering=community&sort=updated_at&order=desc)

## Usage

Start a cluster using the docker driver:

```shell
minikube start --driver=docker
```
To make docker the default driver:

```shell
minikube config set driver docker
```
