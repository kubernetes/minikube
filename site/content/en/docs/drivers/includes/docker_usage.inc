## Requirements

- [Install Docker](https://hub.docker.com/search?q=&type=edition&offering=community&sort=updated_at&order=desc)
- amd64 or arm64 system.

## Usage

Start a cluster using the docker driver:

```shell
minikube start --driver=docker
```
To make docker the default driver:

```shell
minikube config set driver docker
```

