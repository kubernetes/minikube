# vm-driver=none

## Overview

This document is written for system integrators who are familiar with minikube, and wish to run it within a customized VM environment.

The `none` driver allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

## What operating systems are supported?

The `none` driver supports releases of Debian, Ubuntu, and Fedora that are less than 2 years old. In practice, any systemd-based modern distribution is likely to work, and we will accept pull requests which improve compatibility with other systems.

## Example: basic usage

`sudo minikube start --vm-driver=none`

NOTE: The none driver requires minikube to be run as root, until [#3760](https://github.com/kubernetes/minikube/issues/3760) can be addressed.

## Example: Using minikube for continuous integration testing

Most continuous integration environments are already running inside a VM, and may not supported nested virtualization. The `none` driver was designed for this use case. Here is an example, that runs minikube from a non-root user, and ensures that the latest stable kubectl is installed:

```shell
curl -Lo minikube \
  https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 \
  && sudo install minikube /usr/local/bin/

kv=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
curl -Lo kubectl \
  https://storage.googleapis.com/kubernetes-release/release/$kv/bin/linux/amd64/kubectl \
  && sudo install kubectl /usr/local/bin/

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export KUBECONFIG=$HOME/.kube/config

mkdir -p $HOME/.kube $HOME/.minikube
touch $KUBECONFIG

sudo -E minikube start --vm-driver=none
```

At this point, kubectl should be able to interact with the minikube cluster.

## Can the none driver be used outside of a VM?

Yes, *but please avoid doing so if at all possible.*

minikube was designed to run Kubernetes within a dedicated VM, and assumes that it has complete control over the machine it is executing on.  With the `none` driver, minikube and Kubernetes run in an environment with very limited isolation, which could result in:

* Decreased security
* Decreased reliability
* Data loss

We'll cover these in detail below:

### Decreased security

* minikube starts services that may be available on the Internet. Please ensure that you have a firewall to protect your host from unexpected access. For instance:
  * apiserver listens on TCP *:8443
  * kubelet listens on TCP *:10250 and *:10255
  * kube-scheduler listens on TCP *:10259
  * kube-controller listens on TCP *:10257
* Containers may have full access to your filesystem.
* Containers may be able to execute arbitrary code on your host, by using container escape vulnerabilities such as [CVE-2019-5736](https://access.redhat.com/security/vulnerabilities/runcescape). Please keep your release of minikube up to date.

### Decreased reliability

* minikube with the none driver may be tricky to configure correctly at first, because there are many more chances for interference with other locally run services, such as dnsmasq.

* When run in `none` mode, minikube has no built-in resource limit mechanism, which means you could deploy pods which would consume all of the hosts resources.

* minikube and the Kubernetes services it starts may interfere with other running software on the system. For instance, minikube will start and stop container runtimes via systemd, such as docker, containerd, cri-o.

### Data loss

With the `none` driver, minikube will overwrite the following system paths:

* /usr/bin/kubeadm - Updated to match the exact version of Kubernetes selected
* /usr/bin/kubelet - Updated to match the exact version of Kubernetes selected
* /etc/kubernetes - configuration files

These paths will be erased when running `minikube delete`:

* /data/minikube
* /etc/kubernetes/manifests
* /var/lib/minikube

As Kubernetes has full access to both your filesystem as well as your docker images, it is possible that other unexpected data loss issues may arise.

## Environment variables

Some environment variables may be useful for using the `none` driver:

* **CHANGE_MINIKUBE_NONE_USER**: Sets file ownership to the user running sudo ($SUDO_USER)
* **MINIKUBE_HOME**: Saves all files to this directory instead of $HOME
* **MINIKUBE_WANTUPDATENOTIFICATION**: Toggles the notification that your version of minikube is obsolete
* **MINIKUBE_WANTREPORTERRORPROMPT**: Toggles the error reporting prompt
* **MINIKUBE_IN_STYLE**: Toggles color output and emoji usage

## Known Issues

* `systemctl` is required. [#2704](https://github.com/kubernetes/minikube/issues/2704)
* `-p` (profiles) are unsupported: It is not possible to run more than one `--vm-driver=none` instance
* Many `minikube` commands are not supported, such as: `dashboard`, `mount`, `ssh`
* minikube with the `none` driver has a confusing permissions model, as some commands need to be run as root ("start"), and others by a regular user ("dashboard")
* CoreDNS detects resolver loop, goes into CrashloopBackoff - [#3511](https://github.com/kubernetes/minikube/issues/3511)
* Some versions of Linux have a version of docker that is newer then what Kubernetes expects. To overwrite this, run minikube with the following parameters: `sudo -E minikube start --vm-driver=none --kubernetes-version v1.11.8 --extra-config kubeadm.ignore-preflight-errors=SystemVerification`
* On Ubuntu 18.04 (and probably others), because of how `systemd-resolve` is configured by default, one needs to bypass the default `resolv.conf` file and use a different one instead.
  - In this case, you should use this file: `/run/systemd/resolve/resolv.conf`
  - `sudo -E minikube --vm-driver=none start --extra-config=kubelet.resolv-conf=/run/systemd/resolve/resolv.conf`
  - Apperently, though, if `resolve.conf` is too big (about 10 lines!!!), one gets the following error: `Waiting for pods: apiserver proxy! Error restarting cluster: wait: waiting for k8s-app=kube-proxy: timed out waiting for the condition`
  - This error happens in Kubernetes 0.11.x, 0.12.x and 0.13.x, but *not* in 0.14.x
  - If that's your case, try this:
  - `grep -E "^nameserver" /run/systemd/resolve/resolv.conf  |head -n 3 > /tmp/resolv.conf && sudo -E minikube --vm-driver=none start --extra-config=kubelet.resolv-conf=/tmp/resolv.conf`

* [Full list of open 'none' driver issues](https://github.com/kubernetes/minikube/labels/co%2Fnone-driver)
