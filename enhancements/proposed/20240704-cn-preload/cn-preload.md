# CN-Preload and offline minikube

* First proposed: July 4, 2024
* Authors: 锦南路之花 (@ComradeProgrammer)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

<!-- 
_(1 paragraph) What are you proposing, and why is it important to users and/or developers?_ -->

Currently minikube provide preload mechanism. It build one single tarball for kicbase container, which contains all images required by kubeadm. However this kind of preload is file-system relevant (e.g. only available for overlayfs) and cannot be used by all users. This MEP proposes an idea of **adding a new kind of file-system irrelevant preload as a fall-back option**, which can bring benefits to developers in the following aspects:
- Network issue in some regions(Main purpose): In some part of the world neither gcrio nor Dockerhub is not available, and minikube cannot start a kubernetes cluster without images from them. Since June 2024, using mirror registry has been no longer feasible either. Preload can help solve this problem.
- Performance issue (Additional benefits): It will allow more users using different docker filesystem to use preload and get an acceleration for running `minikube start`
- Offline availability (Possible stretch goals): Based on this new preload method, **it is possible that we can make minikube runnable completely offline**, thus make minikube available in more developing or testing environment(e.g. internal on-premise clusters).

## Goals

<!-- *   _A bulleted list of specific goals for this proposal_
*   _How will we know that this proposal has succeeded?_ -->

- Goal 1: `minikube start` can be run in areas where no docker registry is available but `http_proxy` are set.
- Goal 2: All user can use preload for acceleration when using docker/containerd runtime, regardless of their UnionFS.
- Goal 3 (stretch goal): `minikube start` can be run **completely offline** (no internection connection) when preload is also provided locally.

## Non-Goals

- We are not going to change or remove any existing preload methods or features. 

## Design Details

<!-- 
_(2+ paragraphs) A short overview of your implementation idea, containing only as much detail as required to convey your idea._

_If you have multiple ideas, list them concisely._

_Include a testing plan to ensure that your enhancement is not broken by future changes._ -->

### 1. Background: How did all this come up?
Recently we notice that there are more and more issue about failing to start minikube in some internet restriced regions. Previously in those internet restriced area, minikube have a mirror docker registry to pull the docker image(`registry.cn-hangzhou.aliyuncs.com/google_containers`). This registry was not a google or kubernetes official mirror site but a third-party contributor's effort. However images on that mirror site seems outdated. That is most common reason of these issues.

However, since June 2024, using mirror registry has been no longer feasible either due to local legality reason, therefore it is highly not recommended to spend effort to fix the sync problem. We should implement some new approches to use minikube for development, and avoid pulling images from gcrio or dockerhub since they are unavailable in those areas. Then I thought of preloads. 

Preload mechanism should be a very promising approach to solve this problem, because it was exactly designed to bypass image pulling. But currently this approach still cannot work in those regions because:
- When downloading preload, minikube don't use HTTP_PROXY at all if it is a localhost address
- Even if we use HTTP_PROXY, our preload is FS-relevant (actually, for overlayfs only). But there are still many people using other file systems.

Due to reasons above, many users still cannot use preload and fall back to image pulling logic, and eventually fail to start a cluster. This proposal aims to solve these two problems, so that we can provide a new approch to use minikube for development in those internet restriced regions. 

Actually, we can go one step further and add a powerful offline mode for minikube start, so that minikube can be used even there is no network connection at all(e.g. some on-premise cluster). 


### 2. FS-Irrelevant preloads
Currently we provide an all-in-one preload. This is a tarball of a kic container, which has a docker and all image required by kubeadm By doing so, we can preload all images, so that we don't need to preload any images but using local cache. 

However, this tarball is FS-specific. It is because that the docker (inside the kic) will store all images in `/var/lib/docker/<file system name>` (e.g. `/var/lib/docker/overlayfs`). That is the reason why a preload built for a specific file system cannot be used by other filesystem, and why preload cannot always be used for docker container runtime(and containerd...) 

A tarball of a docker image/container itself is irrlevant with the docker file system. Therefore, if we provide a new perload tarball which contains the following things seperately
- A tarball of kic image 
- Tarballs of each images required by kubeadm
- All other binaries required
Then this preload will become fs-irrelevant.

When using this kind of preload, minikube can 
- first load the kic tarball
- start the kic container
- copy all tarballs into the kic container
- run `docker image load` inside the kic container
- run kubeadm

This kind of preload can be used as a fall-back option, since overlay fs is still the most widely used docker fs. We can simply add new logic to `hack/preload-images` and `pkg/minikube/download/preload.go` to achieve this. Existing minikube behavior for preloads will remain intact.


### 3. Respect the value of HTTP_PROXY when downloading minikube preload

Currently minikube cannot properly utilize HTTP_PROXY, because the stardard Golang http client will automatically ignore HTTP_PROXY **if this proxy is on localhost or 127.0.0.1** , unless explicitly specified otherwise. We can change this behavior by passing proxy parameters to the http client if we detect that HTTP_PROXY is set. 

### 4. Offline minikube (stretch goal)

Kubeadm can also be run without internet connection, if all images required are provided. See https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init/#without-internet-connection

Actually we can consider including all other binaries which requires downloading on-the-fly into this universal preload,
so if the users downloaded the minikube binary and preload together, then they don't need any utilize internet connection to start a cluster. This is very useful for some on-premise development enviornment. 

Thus we can add a new offline mode of minikube start by adding a special flag "--offline-preload" to specify the location of pre-download universal preload.

### 5. alter the behavior of "--image-mirror-country" flag

Currently we have a --image-mirror-country flag for those users in internet restricted area. We can alter the behavior of this flag into using the universal fs-irrlevant preload.

This will now break any existing features because user can still use `--registry-mirror` if they have an internal docker mirror registry. 



## Alternatives Considered

<!-- _Alternative ideas that you are leaning against._ -->
One alternative idea is to pull image from dockerhub via HTTP_PROXY, but the problem is that configuring docker engine to use HTTP_PROXY requires restart of docker engine. That means we need to restart it twice(once on host machine and once inside kic). That will be extremely time consuming and unreliable.