---
title: "FAQ"
linkTitle: "FAQ"
weight: 3
description: >
  Frequently Asked Questions
---


## Can I run an older Kubernetes version with minikube? Do I have to downgrade my minikube version?

You do not need to download an older minikube to run an older kubernetes version.
You can create a Kubernetes cluster with any version you desire using `--kubernetes-version` flag.

Example:

```bash
minikube start --kubernetes-version=v1.15.0
```

## How can I create more than one cluster with minikube?

By default, `minikube start` creates a cluster named "minikube". If you would like to create a different cluster or change its name, you can use the `--profile` (or `-p`) flag, which will create a cluster with the specified name. Please note that you can have multiple clusters on the same machine.

To see the list of your current clusters, run:
```
minikube profile list
```

minikube profiles are meant to be isolated from one another, with their own settings and drivers. If you want to create a single cluster with multiple nodes, try the [multi-node feature]({{< ref "/docs/tutorials/multi_node" >}}) instead.

## Can I use minikube as a Docker Desktop replacement?

Yes! Follow our tutorial on [Using minikube as a Docker Desktop Replacement]({{< ref "/docs/tutorials/docker_desktop_replacement" >}}).

## Can I start minikube without Kubernetes running?

Yes! If you want to use minikube only as a Docker Desktop replacement without starting Kubernetes itself, try:
```
minikube start --container-runtime=docker --no-kubernetes
```

Alternatively, if you want to temporarily turn off Kubernetes, you can pause and later unpause Kubernetes 
```
minikube pause
```

minikube also has an addon that automatically pauses Kubernetes after a certain amount of inactivity:

```
minikube addons enable auto-pause
```



## Docker Driver: How can I set minikube's cgroup manager?

For non-VM and non-SSH drivers, minikube will try to auto-detect your system's cgroups driver/manager and configure all other components accordingly.
For VM and SSH drivers, minikube will use cgroupfs cgroups driver/manager by default.
To force the `systemd` cgroup manager, run:

```bash
minikube start --force-systemd=true
```

## How can I run minikube with the Docker driver if I have an existing cluster with a VM driver?

First please ensure your Docker service is running. Then you need to either:  

(a) Delete the existing cluster and create a new one

```bash
minikube delete
minikube start --driver=docker
```

Alternatively, (b) Create a second cluster with a different profile name:

```bash
minikube start -p p1 --driver=docker 
```

## Does minikube support IPv6?

minikube currently doesn't support IPv6. However, it is on the [roadmap]({{< ref "/docs/contrib/roadmap.en.md" >}}). You can also refer to the [open issue](https://github.com/kubernetes/minikube/issues/8535).

## How can I prevent password prompts on Linux?

The easiest approach is to use the `docker` driver, as the backend service always runs as `root`.

`none` users may want to try `CHANGE_MINIKUBE_NONE_USER=true`, where kubectl and such will work without `sudo`. See [environment variables]({{< ref "/docs/handbook/config.md#environment-variables" >}}) for more details.  

Alternatively, you can configure `sudo` to never prompt for commands issued by minikube.

## How can I ignore system verification?

[kubeadm](https://github.com/kubernetes/kubeadm), minikube's bootstrapper, verifies a list of features on the host system before installing Kubernetes. In the case you get an error and still want to try minikube despite your system's limitation, you can skip verification by starting minikube with this extra option:

```shell
minikube start --extra-config kubeadm.ignore-preflight-errors=SystemVerification
```

## What is the minimum resource allocation necessary for a Knative setup using minikube?

Please allocate sufficient resources for Knative setup using minikube, especially when running minikube cluster on your local machine. We recommend allocating at least 3 CPUs and 3G memory.

```shell
minikube start --cpus 3 --memory 3072
```

## Do I need to install kubectl locally?

No, minikube comes with a built-in kubectl installation. See [minikube's kubectl documentation]({{< ref "docs/handbook/kubectl.md" >}}).

## How can I opt-in to beta release notifications?

Simply run the following command to be enrolled into beta notifications:
```
minikube config set WantBetaUpdateNotification true
```

## Can I get rid of the emoji in minikube's output?

Yes! If you prefer not having emoji in your minikube output 😔 , just set the `MINIKUBE_IN_STYLE` environment variable to `0` or `false`:

```
MINIKUBE_IN_STYLE=0 minikube start

```

## How can I access a minikube cluster from a remote network?

minikube's primary goal is to quickly set up local Kubernetes clusters, and therefore we strongly discourage using minikube in production or for listening to remote traffic. By design, minikube is meant to only listen on the local network.

However, it is possible to configure minikube to listen on a remote network. This will open your network to the outside world and is not recommended. If you are not fully aware of the security implications, please avoid using this.

For the docker and podman driver, use `--listen-address` flag:

```
minikube start --listen-address=0.0.0.0
```

## How can I allocate maximum resources to minikube?

Setting the `memory` and `cpus` flags on the start command to `max` will use maximum available resources:
```
minikube start --memory=max --cpus=max
```

## How can I run minikube on a different hard drive?

Set the `MINIKUBE_HOME` env to a path on the drive you want minikube to run, then run `minikube start`.

```
# Unix
export MINIKUBE_HOME=/otherdrive/.minikube

# Windows
$env:MINIKUBE_HOME = "D:\.minikube"

minikube start
```

## Can I set a static IP for the minikube cluster?

Currently a static IP can only be set when using the Docker or Podman driver.

For more details see the [static IP tutorial]({{< ref "docs/tutorials/static_ip.md" >}}).

## How to ignore the kubeadm requirements and pre-flight checks (such as minimum CPU count)?

Kubeadm has certain software and hardware requirements to maintain a stable Kubernetes cluster. However, these requirements can be ignored (such as when running minikube on a single CPU) by running the following:
```
minikube start --force --extra-config=kubeadm.skip-phases=preflight
```
This is not recommended, but for some users who are willing to accept potential performance or stability issues, this may be the only option.
## I am in China, and I encounter errors when trying to start minikube, what should I do?

After executing `minikube start`, minikube will try to pulling images from `gcr.io` or Docker Hub. However, it has been confirmed that Chinese (mainland) users may not have access to `gcr.io` or Docker Hub. So in China mainland, it is very likely that `minikube start` will fail.

For Chinese users, the reason is that China mainland government has set up GFW firewall to block any access to `gcr.io` or Docker Hub from China mainland. 

Minikube is an open community and we are always willing to help users from any corner of the world to use our open-source software, and provide possible assistance when possible. Here are 3 possible ways to resolve the blockade.

1. Use `minikube start --image-mirror-country='cn'` instead. Aliyun (a Chinese corporation) provides a mirror repository (`registry.cn-hangzhou.aliyuncs.com/google_containers`) for those images, to which Chinese users have access. By using `--image-mirror-country='cn'` flag, minikube will try to pull the image from Aliyun mirror site as first priority. <br/><br/> *Note: when a new image is published on gcr.io, it may take several days for the image to be synchronized to Aliyun mirror repo. However, minikube will always try to pull the newest image by default, which will cause a failure of pulling image. Under this circumstance, you HAVE TO use `--kubernetes-version` flag AS WELL to tell minikube to use an older version image which is available on Aliyun repo.* <br/><br/> *For example, `minikube start --image-mirror-country='cn'  --kubernetes-version=v1.23.8` will tell minikube to pull v1.23.8 k8s image from Aliyun.*

2. If you have a private mirror repository provided by your own cloud provider, you can specify that via `--image-repository` flag. For example, using `minikube start --image-repository='registry.cn-hangzhou.aliyuncs.com/google_containers'` will tell minikube to try to pull images from `registry.cn-hangzhou.aliyuncs.com/google_containers` mirror repository as first priority. 
  
3. Use a proxy server/VPN, if you have one. <br/> *Note: please obey the local laws. In some area, using an unauthorized proxy server/VPN is ILLEGAL* 

## How do I install containernetworking-plugins for none driver?

Go to [containernetworking-plugins](https://github.com/containernetworking/plugins/releases) to find the latest version.

Then execute the following:
```shell
CNI_PLUGIN_VERSION="<version_here>"
CNI_PLUGIN_TAR="cni-plugins-linux-amd64-$CNI_PLUGIN_VERSION.tgz" # change arch if not on amd64
CNI_PLUGIN_INSTALL_DIR="/opt/cni/bin"

curl -LO "https://github.com/containernetworking/plugins/releases/download/$CNI_PLUGIN_VERSION/$CNI_PLUGIN_TAR"
sudo mkdir -p "$CNI_PLUGIN_INSTALL_DIR"
sudo tar -xf "$CNI_PLUGIN_TAR" -C "$CNI_PLUGIN_INSTALL_DIR"
rm "$CNI_PLUGIN_TAR"
```
