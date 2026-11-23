## Experimental

This is an experimental driver. Please use it only for experimental reasons until it has reached maturity. For a more reliable minikube experience, use a non-experimental driver, like [Docker](https://minikube.sigs.k8s.io/docs/drivers/docker/).

## Usage

It's recommended to run minikube with the podman driver and [CRI-O container runtime](https://cri-o.io/) (except when using Rootless Podman):

```shell
minikube start --driver=podman --container-runtime=cri-o
```

Alternatively, start minikube with the podman driver only:

```shell
minikube start --driver=podman
```

To make podman the default driver:

```shell
minikube config set driver podman
```

## Rootless Podman

By default, minikube executes Podman with `sudo`.
To use Podman without `sudo` (i.e., Rootless Podman), set the `rootless` property to `true`:

```shell
minikube config set rootless true
```

Alternatively, pass the `--rootless` flag when starting minikube:

```shell
minikube start --driver=podman --rootless
```

### Host prerequisites

When running the podman driver in rootless mode, minikube runs as your non-root user. The host must configure subordinate UID and GID ranges in `/etc/subuid` and `/etc/subgid` for that user. Without them, rootless Podman fails to allocate user namespaces and `minikube start` fails.

Tested configuration on Ubuntu 24.04:

```shell
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER
podman system migrate
```

After configuring subordinate IDs, log out and log back in (or restart your session) for the changes to take effect.

> **Note:** This configuration was verified on Ubuntu 24.04; please validate and report results for other distributions (Fedora, Debian, etc.). See the upstream [Podman rootless tutorial](https://github.com/podman-container-tools/podman/blob/main/docs/tutorials/rootless_tutorial.md#etcsubuid-and-etcsubgid-configuration) for details on `/etc/subuid` and `/etc/subgid` configuration.

For Rootless Podman, it is recommended to set `--container-runtime` to `containerd`:

```shell
minikube start --driver=podman --container-runtime=containerd
```

See the [Rootless Docker](https://minikube.sigs.k8s.io/docs/drivers/docker/#rootless-docker) section for additional requirements and restrictions.
