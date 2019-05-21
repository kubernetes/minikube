# Building images within the VM

When using a single VM of Kubernetes it's really handy to build inside the VM; as this means you don't have to build on your host machine and push the image into a docker registry - you can just build inside the same machine as minikube which speeds up local experiments.

## Docker (containerd)

For Docker, you can either set up your host docker client to communicate by [reusing the docker daemon](reusing_the_docker_daemon.md).

Or you can use `minikube ssh` to connect to the virtual machine, and run the `docker build` there:

```shell
docker build
```

For more information on the `docker build` command, read the [Docker documentation](https://docs.docker.com/engine/reference/commandline/build/) (docker.com).

## Podman (cri-o)

For Podman, there is no daemon running. The processes are started by the user, monitored by `conmon`.

So you need to use `minikube ssh`, and you will also make sure to run the command as the root user:

```shell
sudo -E podman build
```

For more information on the `podman build` command, read the [Podman documentation](https://github.com/containers/libpod/blob/master/docs/podman-build.1.md) (podman.io).

## Build context

For the build context you can use any directory on the virtual machine, or any address on the network.
