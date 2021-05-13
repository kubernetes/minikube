---
title: "minikube start"
linkTitle: "Get Started!"
weight: 1
aliases:
  - /docs/start
---

minikube is local Kubernetes, focusing on making it easy to learn and develop for Kubernetes.

All you need is Docker (or similarly compatible) container or a Virtual Machine environment, and Kubernetes is a single command away: `minikube start`

## What you’ll need

* 2 CPUs or more
* 2GB of free memory
* 20GB of free disk space
* Internet connection
* Container or virtual machine manager, such as: [Docker]({{<ref "/docs/drivers/docker">}}), [Hyperkit]({{<ref "/docs/drivers/hyperkit">}}), [Hyper-V]({{<ref "/docs/drivers/hyperv">}}), [KVM]({{<ref "/docs/drivers/kvm2">}}), [Parallels]({{<ref "/docs/drivers/parallels">}}), [Podman]({{<ref "/docs/drivers/podman">}}), [VirtualBox]({{<ref "/docs/drivers/virtualbox">}}), or [VMWare]({{<ref "/docs/drivers/vmware">}})

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">1</strong></span>Installation</h2>

Click on the buttons that describe your target platform:

{{% quiz_row base="" name="Operating system" %}}
{{% quiz_button option="Linux" %}} {{% quiz_button option="macOS" %}} {{% quiz_button option="Windows" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux" name="Architecture" %}}
{{% quiz_button option="x86-64" %}} {{% quiz_button option="ARM64" %}} {{% quiz_button option="ARMv7" %}} {{% quiz_button option="ppc64" %}} {{% quiz_button option="S390x" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/x86-64" name="Installer type" %}}
{{% quiz_button option="Binary download" %}} {{% quiz_button option="Debian package" %}} {{% quiz_button option="RPM package" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ARM64" name="Installer type" %}}
{{% quiz_button option="Binary download" %}} {{% quiz_button option="Debian package" %}} {{% quiz_button option="RPM package" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ppc64" name="Installer type" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/S390x" name="Installer type" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Linux/ARMv7" name="Installer type" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS" name="Architecture" %}}
{{% quiz_button option="x86-64" %}} {{% quiz_button option="ARM64" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS/x86-64" name="Installer type" %}}
{{% quiz_button option="Homebrew" %}} {{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/macOS/ARM64" name="Installer type" %}}
{{% quiz_button option="Binary download" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Windows" name="Architecture" %}}
{{% quiz_button option="x86-64" %}}
{{% /quiz_row %}}

{{% quiz_row base="/Windows/x86-64" name="Installer type" %}}
{{% quiz_button option="Windows Package Manager" %}} {{% quiz_button option="Chocolatey" %}} {{% quiz_button option=".exe download" %}}
{{% /quiz_row %}}

{{% quiz_instruction id="/Linux/x86-64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/x86-64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_amd64.deb
sudo dpkg -i minikube_latest_amd64.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/x86-64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.x86-64.rpm
sudo rpm -Uvh minikube-latest.x86-64.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-arm64
sudo install minikube-linux-arm64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_arm64.deb
sudo dpkg -i minikube_latest_arm64.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARM64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.aarch64.rpm
sudo rpm -Uvh minikube-latest.aarch64.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-ppc64le
sudo install minikube-linux-ppc64le /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_ppc64le.deb
sudo dpkg -i minikube_latest_ppc64le.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ppc64/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.ppc64el.rpm
sudo rpm -Uvh minikube-latest.ppc64el.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-s390x
sudo install minikube-linux-s390x /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/Debian package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube_latest_s390x.deb
sudo dpkg -i minikube_latest_s390x.deb
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/S390x/RPM package" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.s390x.rpm
sudo rpm -Uvh minikube-latest.s390x.rpm
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Linux/ARMv7/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-arm
sudo install minikube-linux-arm /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/x86-64/Homebrew" %}}
If the [Brew Package Manager](https://brew.sh/) is installed:

```shell
brew install minikube
```

If `which minikube` fails after installation via brew, you may have to remove the old minikube links and link the newly installed binary:

```shell
brew unlink minikube
brew link minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/x86-64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-amd64
sudo install minikube-darwin-amd64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/macOS/ARM64/Binary download" %}}
```shell
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-darwin-arm64
sudo install minikube-darwin-arm64 /usr/local/bin/minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/Windows Package Manager" %}}
If the [Windows Package Manager](https://docs.microsoft.com/en-us/windows/package-manager/) is installed, use the following command to install minikube:

```shell
winget install minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/Chocolatey" %}}
If the [Chocolatey Package Manager](https://chocolatey.org/) is installed, use the following command:

```shell
choco install minikube
```
{{% /quiz_instruction %}}

{{% quiz_instruction id="/Windows/x86-64/.exe download" %}}
Download and run the stand-alone [minikube Windows installer](https://storage.googleapis.com/minikube/releases/latest/minikube-installer.exe).

_If you used a CLI to perform the installation, you will need to close that CLI and open a new one before proceeding._
{{% /quiz_instruction %}}


<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">2</strong></span>Start your cluster</h2>

From a terminal with administrator access (but not logged in as root), run:

```shell
minikube start
```

If minikube fails to start, see the [drivers page]({{<ref "/docs/drivers">}}) for help setting up a compatible container or virtual-machine manager.

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">3</strong></span>Interact with your cluster</h2>

If you already have kubectl installed, you can now use it to access your shiny new cluster:

```shell
kubectl get po -A
```

Alternatively, minikube can download the appropriate version of kubectl, if you don't mind the double-dashes in the command-line:

```shell
minikube kubectl -- get po -A
```

Initially, some services such as the storage-provisioner, may not yet be in a Running state. This is a normal condition during cluster bring-up, and will resolve itself momentarily. For additional insight into your cluster state, minikube bundles the Kubernetes Dashboard, allowing you to get easily acclimated to your new environment:

```shell
minikube dashboard
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">4</strong></span>Deploy applications</h2>

Create a sample deployment and expose it on port 8080:

```shell
kubectl create deployment hello-minikube --image=k8s.gcr.io/echoserver:1.4
kubectl expose deployment hello-minikube --type=NodePort --port=8080
```

It may take a moment, but your deployment will soon show up when you run:

```shell
kubectl get services hello-minikube
```

The easiest way to access this service is to let minikube launch a web browser for you:

```shell
minikube service hello-minikube
```

Alternatively, use kubectl to forward the port:

```shell
kubectl port-forward service/hello-minikube 7080:8080
```

Tada! Your application is now available at [http://localhost:7080/](http://localhost:7080/)

### LoadBalancer deployments

To access a LoadBalancer deployment, use the "minikube tunnel" command. Here is an example deployment:

```shell
kubectl create deployment balanced --image=k8s.gcr.io/echoserver:1.4  
kubectl expose deployment balanced --type=LoadBalancer --port=8080
```

In another window, start the tunnel to create a routable IP for the 'balanced' deployment:

```shell
minikube tunnel
```

To find the routable IP, run this command and examine the `EXTERNAL-IP` column:

```shell
kubectl get services balanced
```

Your deployment is now available at &lt;EXTERNAL-IP&gt;:8080

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">5</strong></span>Manage your cluster</h2>

Pause Kubernetes without impacting deployed applications:

```shell
minikube pause
```

Halt the cluster:

```shell
minikube stop
```

Increase the default memory limit (requires a restart):

```shell
minikube config set memory 16384
```

Browse the catalog of easily installed Kubernetes services:

```shell
minikube addons list
```

Create a second cluster running an older Kubernetes release:

```shell
minikube start -p aged --kubernetes-version=v1.16.1
```

Delete all of the minikube clusters:

```shell
minikube delete --all
```

## Take the next step

* [The minikube handbook]({{<ref "/docs/handbook">}})
* [Community-contributed tutorials]({{<ref "/docs/tutorials">}})
* [minikube command reference]({{<ref "/docs/commands">}})
* [Contributors guide]({{<ref "/docs/contrib">}})
* Take our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A) to share your thoughts 🙏
