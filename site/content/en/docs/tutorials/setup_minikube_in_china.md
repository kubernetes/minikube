---
title: "Setting up minikube in China"
linkTitle: "Setup minikube in China"
weight: 1
---

This Tutorial explains how to setup minikube in China.

### Binary download

For more detail please refer to [minikube installation guide]({{<ref "/docs/start">}})

Open a terminal. If on Windows, open a terminal with administrator access. Run:

```shell
minikube start --image-repository=registry.cn-hangzhou.aliyuncs.com/google_containers
```
If minikube fails to start, see the [drivers page]({{<ref "/docs/drivers">}}) for help setting up a compatible container or virtual-machine manager.

For users who use the [Hyper-V](https://minikube.sigs.k8s.io/docs/drivers/hyperv/) environment, you should first open the Hyper-V manager to create an external virtual switch. After that, we can use the following commands to create a Hyper-V-based Kubernetes test environment:

```shell
.\minikube.exe start --image-mirror-country cn \
    --iso-url=https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-v1.12.0.iso \
    --registry-mirror=https://xxxxxx.mirror.aliyuncs.com \
    --vm-driver="hyperv"
```

For users who use hyperkit environment:

### Download the Aliyun version minikube

```shell
curl -Lo minikube https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/releases/v1.14.2/minikube-darwin-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/
```

NOTE: You can find if there is a new version to replace v1.14.2 in the command above at
https://github.com/AliyunContainerService/minikube/wiki#%E5%AE%89%E8%A3%85minikube

### Start minikube

```shell
minikube start --image-mirror-country cn \
 --driver=hyperkit \
 --iso-url=https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-v1.15.0.iso \
 --registry-mirror=https://xxxxxxxx.mirror.aliyuncs.com
```

NOTE 1: You can find latest minikube version at https://github.com/kubernetes/minikube/releases.
However, Aliyun's minikube version is a little behind. To verify if a new version exists, you can replace the version in the URL of https://kubernetes.oss-cn-hangzhou.aliyuncs.com/minikube/iso/minikube-v1.15.0.iso to different new versions, such as v1.15.1, and then open it in your browser.

NOTE 2:  For the xxxxxxxx in the command above, you can find yours at
https://cr.console.aliyun.com/cn-hangzhou/instances/mirrors
(Need to register an Aliyun account first)

NOTE 3: You can pass more parameters to this Aliyun version of minikube start, check at https://github.com/AliyunContainerService/minikube/wiki#%E5%90%AF%E5%8A%A8
In this case, driver used is hyperkit on macOS, Aliyun's iso-url and registry-mirror to speed it up.

Please follow [Interact with your cluster guide]({{<ref "/docs/start">}}) to interact with your cluster.

