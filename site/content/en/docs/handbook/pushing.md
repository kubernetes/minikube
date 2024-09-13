---
title: "Pushing images"
weight: 5
description: >
 comparing 8 ways to push your image into a minikube cluster.
aliases:
 - /docs/tasks/building
 - /docs/tasks/caching
 - /docs/tasks/podman_service
 - /docs/tasks/docker_daemon
---

Glossary:

**Pull** means downloading a container image directly from a remote registry.

**Push** means uploading a container image directly to a remote registry.

**Load** takes an image that is available as an archive, and makes it available in the cluster.

**Save** saves an image into an archive.

**Build** takes a "build context" (directory) and creates a new image in the cluster from it.

**Tag** means assigning a name and tag.

## Comparison table for different methods

The best method to push your image to minikube depends on the container-runtime you built your cluster with (the default is docker).
Here is a comparison table to help you choose:

| Method | Supported Runtimes | Performance | Load | Build |
|--- |--- |--- |--- |--- |--- |--- |
|  [docker-env command](/docs/handbook/pushing/#1-pushing-directly-to-the-in-cluster-docker-daemon-docker-env) |   docker & containerd |  good  | yes | yes |
|  [cache command](/docs/handbook/pushing/#2-push-images-using-cache-command) |  all  |  ok  | yes | no |
|  [podman-env command](/docs/handbook/pushing/#3-pushing-directly-to-in-cluster-cri-o-podman-env) |   only cri-o |  good  | yes | yes |
|  [registry addon](/docs/handbook/pushing/#4-pushing-to-an-in-cluster-using-registry-addon)   |   all |  ok  | yes | no |
|  [minikube ssh](/docs/handbook/pushing/#5-building-images-inside-of-minikube-using-ssh)   |   all | best  | yes\* | yes\* |
|  [ctr/buildctl command](/docs/handbook/pushing/#6-pushing-directly-to-in-cluster-containerd-buildkitd) |   only containerd |  good  | yes | yes |
|  [image load command](/docs/handbook/pushing/#7-loading-directly-to-in-cluster-container-runtime)  |  all  |  ok  | yes | no |
|  [image build command](/docs/handbook/pushing/#8-building-images-to-in-cluster-container-runtime)  |  all  |  ok  | no | yes |

* Note 1: The default container-runtime on minikube is `docker`.
* Note 2: The `none` driver (bare metal) does not need pushing image to the cluster, as any image on your system is already available to the Kubernetes cluster.
* Note 3: When using ssh to run the commands, the files to load or build must already be available on the node (not only on the client host).

---

## 1. Pushing directly to the in-cluster Docker daemon (docker-env)

This is similar to podman-env but only for Docker runtime.
When using a container or VM driver (all drivers except none), you can reuse the Docker daemon inside minikube cluster.
This means you don't have to build on your host machine and push the image into a docker registry. You can just build inside the same docker daemon as minikube which speeds up local experiments.

To point your terminal to use the docker daemon inside minikube run this:

{{% tabs %}}
{{% linuxtab %}}
```shell
eval $(minikube docker-env)
```
{{% /linuxtab %}}
{{% mactab %}}
```shell
eval $(minikube docker-env)
```
{{% /mactab %}}
{{% windowstab %}}
PowerShell
```shell
& minikube -p minikube docker-env --shell powershell | Invoke-Expression
```

cmd
```shell
@FOR /f "tokens=*" %i IN ('minikube -p minikube docker-env --shell cmd') DO @%i
```
{{% /windowstab %}}
{{% /tabs %}}

Now any 'docker' command you run in this current terminal will run against the docker inside minikube cluster.

So if you do the following commands, it will show you the containers inside the minikube, inside minikube's VM or Container.

```shell
docker ps
```

Now you can 'build' against the docker inside minikube, which is instantly accessible to kubernetes cluster.

```shell
docker build -t my_image .
```

To verify your terminal is using minikube's docker-env you can check the value of the environment variable MINIKUBE_ACTIVE_DOCKERD to reflect the cluster name.

{{% pageinfo color="info" %}}
Tip 1:
Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`) in your yaml file. Otherwise Kubernetes won't use your locally build image and it will pull from the network.
{{% /pageinfo %}}

{{% pageinfo color="info" %}}
Tip 2:
Evaluating the docker-env is only valid for the current terminal.
By closing the terminal, you will go back to using your own system's docker daemon.
{{% /pageinfo %}}

{{% pageinfo color="info" %}}
Tip 3:
In container-based drivers such as Docker or Podman, you will need to re-do docker-env each time you restart your minikube cluster.
{{% /pageinfo %}}

More information on [docker-env](https://minikube.sigs.k8s.io/docs/commands/docker-env/)

---

## 2. Push images using 'cache' command.

From your host, you can push a Docker image directly to minikube. This image will be cached and automatically pulled into all future minikube clusters created on the machine

```shell
minikube cache add alpine:latest
```

The add command will store the requested image to `$MINIKUBE_HOME/cache/images`, and load it into the minikube cluster's container runtime environment automatically.

{{% pageinfo color="info" %}}
Tip 1 :
If your image changes after your cached it, you need to do 'cache reload'.
{{% /pageinfo %}}

minikube refreshes the cache images on each start. However to reload all the cached images on demand, run this command :
```shell
minikube cache reload
```

{{% pageinfo color="info" %}}
Tip 2 :
If you have multiple clusters, the cache command will load the image for all of them.
{{% /pageinfo %}}

To display images you have added to the cache:

```shell
minikube cache list
```

This listing will not include the images minikube's built-in system images.

```shell
minikube cache delete <image name>
```

For more information, see:

* [Reference: cache command]({{< ref "/docs/commands/cache.md" >}})

---

## 3. Pushing directly to in-cluster CRI-O. (podman-env)

{{% tabs %}}
{{% linuxtab %}}
This is similar to docker-env but only for CRI-O runtime.
To push directly to CRI-O, configure podman client on your host using the podman-env command in your shell:

```shell
eval $(minikube podman-env)
```

You should now be able to use podman client on the command line on your host machine talking to the podman service inside the minikube VM:

```shell
podman-remote help
```

Now you can 'build' against the storage inside minikube, which is instantly accessible to kubernetes cluster.

```shell
podman-remote build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On Linux the remote client is called "podman-remote", while the local program is called "podman".
{{% /pageinfo %}}

{{% /linuxtab %}}
{{% mactab %}}
This is similar to docker-env but only for CRI-O runtime.
To push directly to CRI-O, configure Podman client on your host using the podman-env command in your shell:

```shell
eval $(minikube podman-env)
```

You should now be able to use Podman client on the command line on your host machine talking to the Podman service inside the minikube VM:

```shell
podman help
```

Now you can 'build' against the storage inside minikube, which is instantly accessible to Kubernetes cluster.

```shell
podman build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On macOS the remote client is called "podman", since there is no local "podman" program available.
{{% /pageinfo %}}

{{% /mactab %}}
{{% windowstab %}}
This is similar to docker-env but only for CRI-O runtime.
To push directly to CRI-O, configure Podman client on your host using the podman-env command in your shell:

PowerShell
```shell
& minikube -p minikube podman-env --shell powershell | Invoke-Expression
```

cmd
```shell
@FOR /f "tokens=*" %i IN ('minikube -p minikube podman-env --shell cmd') DO @%i
```

You should now be able to use Podman client on the command line on your host machine talking to the Podman service inside the minikube VM:

Now you can 'build' against the storage inside minikube, which is instantly accessible to Kubernetes cluster.

```shell
podman help
```

```shell
podman build -t my_image .
```

{{% pageinfo color="info" %}}
Note: On Windows the remote client is called "podman", since there is no local "podman" program available.
{{% /pageinfo %}}

{{% /windowstab %}}
{{% /tabs %}}

Remember to turn off the `imagePullPolicy:Always` (use `imagePullPolicy:IfNotPresent` or `imagePullPolicy:Never`), as otherwise Kubernetes won't use images you built locally.

---

## 4. Pushing to an in-cluster using Registry addon

For illustration purpose, we will assume that minikube VM has one of the ip from `192.168.39.0/24` subnet. If you have not overridden these subnets as per [networking guide](https://minikube.sigs.k8s.io/reference/networking/), you can find out default subnet being used by minikube for a specific OS and driver combination [here](https://github.com/kubernetes/minikube/blob/dfd9b6b83d0ca2eeab55588a16032688bc26c348/pkg/minikube/cluster/cluster.go#L408) which is subject to change. Replace `192.168.39.0/24` with appropriate values for your environment wherever applicable.

Ensure that docker is configured to use `192.168.39.0/24` as insecure registry. Refer [here](https://docs.docker.com/registry/insecure/) for instructions.

Ensure that `192.168.39.0/24` is enabled as insecure registry in minikube. Refer [here](https://minikube.sigs.k8s.io/docs/handbook/registry/#enabling-insecure-registries/) for instructions..

Enable minikube registry addon:

```shell
minikube addons enable registry
```

Build docker image and tag it appropriately:

```shell
docker build --tag $(minikube ip):5000/test-img .
```

Push docker image to minikube registry:

```shell
docker push $(minikube ip):5000/test-img
```

---

## 5. Building images inside of minikube using SSH

Use `minikube ssh` to run commands inside the minikube node, and run the build command directly there.
Any command you run there will run against the same daemon / storage that kubernetes cluster is using.

For Docker, use:

```shell
docker build
```

For more information on the `docker build` command, read the [Docker documentation](https://docs.docker.com/engine/reference/commandline/build/) (docker.com).

For CRI-O, use:

```shell
sudo podman build
```

For more information on the `podman build` command, read the [Podman documentation](https://github.com/containers/podman/blob/master/docs/source/markdown/podman-build.1.md) (podman.io).

For Containerd, use:

```shell
sudo ctr images import
```

```shell
sudo buildctl build
```

For more information on the `ctr images` command, read the [containerd documentation](https://containerd.io/docs/getting-started/) (containerd.io)

For more information on the `buildctl build` command, read the [Buildkit documentation](https://github.com/moby/buildkit#quick-start) (mobyproject.org).

To exit minikube ssh and come back to your terminal type:

```shell
exit
```

## 6. Pushing directly to in-cluster containerd (buildkitd)

This is similar to docker-env and podman-env but only for Containerd runtime.

Currently it requires starting the daemon and setting up the tunnels manually.

### `ctr` instructions

In order to access containerd, you need to log in as `root`.
This requires adding the ssh key to `/root/authorized_keys`..

```console
docker@minikube:~$ sudo mkdir /root/.ssh
docker@minikube:~$ sudo chmod 700 /root/.ssh
docker@minikube:~$ sudo cp .ssh/authorized_keys /root/.ssh/authorized_keys
docker@minikube:~$ sudo chmod 600 /root/.ssh
```

Note the flags that are needed for the `ssh` command.

```bash
minikube --alsologtostderr ssh --native-ssh=false
```

Tunnel the containerd socket to the host, from the machine.
(_Use above ssh flags (most notably the -p port and root@host)_)

```bash
ssh -nNT -L ./containerd.sock:/run/containerd/containerd.sock ... &
```

Now you can run command to this unix socket, tunneled over ssh.

```bash
ctr --address ./containerd.sock help
```

Images in "k8s.io" namespace are accessible to kubernetes cluster.

### `buildctl` instructions

Start the BuildKit daemon, using the containerd backend.

```console
docker@minikube:~$ sudo -b buildkitd --oci-worker=false --containerd-worker=true --containerd-worker-namespace=k8s.io
```

Make the BuildKit socket accessible to the regular user.

```console
docker@minikube:~$ sudo groupadd buildkit
docker@minikube:~$ sudo chgrp -R buildkit /run/buildkit
docker@minikube:~$ sudo usermod -aG buildkit $USER
docker@minikube:~$ exit
```

Note the flags that are needed for the `ssh` command.

```bash
minikube --alsologtostderr ssh --native-ssh=false
```

Tunnel the BuildKit socket to the host, from the machine.
(_Use above ssh flags (most notably the -p port and user@host)_)

```bash
ssh -nNT -L ./buildkitd.sock:/run/buildkit/buildkitd.sock ... &
```

After that, it should now be possible to use `buildctl`:

```bash
buildctl --addr unix://buildkitd.sock build \
    --frontend=dockerfile.v0 \
    --local context=. \
    --local dockerfile=. \
    --output type=image,name=registry.k8s.io/username/imagename:latest
```

Now you can 'build' against the storage inside minikube. which is instantly accessible to kubernetes cluster.

---

## 7. Loading directly to in-cluster container runtime

The minikube client will talk directly to the container runtime in the
cluster, and run the load commands there - against the same storage.

```shell
minikube image load my_image
```

For more information, see:

* [Reference: image load command]({{< ref "/docs/commands/image.md#minikube-image-load" >}})

---

## 8. Building images to in-cluster container runtime

The minikube client will talk directly to the container runtime in the
cluster, and run the build commands there - against the same storage.

```shell
minikube image build -t my_image .
```

For more information, see:

* [Reference: image build command]({{< ref "/docs/commands/image.md#minikube-image-build" >}})
