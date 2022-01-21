## Requirements

- [Install Docker](https://hub.docker.com/search?q=&type=edition&offering=community&sort=updated_at&order=desc) 18.09 or higher
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

## Rootless Docker
### Requirements
- Docker 20.10 or higher, see https://rootlesscontaine.rs/getting-started/docker/
- Cgroup v2 delegation, see https://rootlesscontaine.rs/getting-started/common/cgroup2/

### Usage

Start a cluster using the rootless docker driver:

```shell
dockerd-rootless-setuptool.sh install -f
docker context use rootless

minikube start --driver=docker --container-runtime=containerd
```

The `--container-runtime` flag must be set to "containerd" or "cri-o".

The restrictions of rootless `kind` apply to minikube with rootless docker as well.

See https://kind.sigs.k8s.io/docs/user/rootless/ .
