---
title: "Using minikube as Docker Desktop Replacement"                     
linkTitle: "Using minikube as Docker Desktop Replacement"
weight: 1
date: 2022-02-02
---

## Overview

- This guide will show you how to use minikube as a Docker Desktop replacement.

## Before You Begin
- This only works with the `docker` container runtime, not with `containerd` or `crio`.

- You need to start minikube with a VM driver instead of `docker`, such as `hyperkit` on macOS and `hyperv` on Windows.

- Alternatively, you can use the [`minikube image build`]({{< ref "/docs/commands/image#minikube-image-build" >}}) command instead of `minikube docker-env` and `docker build`.

## Steps
<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">1</strong></span>Install the Docker CLI</h2>
{{% tabs %}}
{{% mactab %}}
{{% tabs %}}
{{% tab brew %}}
```shell
brew install docker
```
{{% /tab %}}
{{% tab Manual %}}
1. Download the static binary archive. Go to https://download.docker.com/mac/static/stable/ and select `x86_64` (for Mac on Intel chip) or `aarch64` (for Mac on Apple silicon), and then download the `.tgz` file relating to the version of Docker Engine you want to install.

2. Extract the archive using the `tar` utility. The `docker` binary is extracted.
```shell
tar xzvf /path/to/<FILE>.tar.gz
```

3. Clear the extended attributes to allow it run.
```shell
sudo xattr -rc docker
```

4. Move the binary to a directory on your executable path, such as `/usr/local/bin/`.
```shell
sudo cp docker/docker /usr/local/bin/
```
{{% /tab %}}
{{% /tabs %}}
{{% /mactab %}}
{{% windowstab %}}
{{% tabs %}}
{{% tab Chocolatey %}}
**Please Note**: The docker engine requires the Windows Features: Containers and Microsoft-Hyper-V to be installed in order to function correctly. You can install these with the chocolatey command:
```shell
choco install Containers Microsoft-Hyper-V --source windowsfeatures
```

1. Install docker-engine
```shell
choco install docker-engine
```

2. This package creates the group `docker-users` and adds the installing user to it. In order to communicate with docker you will need to log out and back in.
{{% /tab %}}
{{% tab Manual %}}
1. Download the static binary archive. Go to https://download.docker.com/win/static/stable/x86_64 and select the latest version from the list.

2. Run the following PowerShell commands to install and extract the archive to your program files:
```shell
Expand-Archive /path/to/<FILE>.zip -DestinationPath $Env:ProgramFiles
```

3. Add the path to the Docker CLI binary (`C:\Program Files\Docker`) to the `PATH` environment variable, [guide to setting environment variables in Windows](https://www.architectryan.com/2018/08/31/how-to-change-environment-variables-on-windows-10/).

4. Restart Windows for the `PATH` change to take effect.
{{% /tab %}}
{{% /tabs %}}
{{% /windowstab %}}
{{% /tabs %}}

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">2</strong></span>Start minikube</h2>
Start minikube with a VM driver and `docker` container runtime if not already running.

```shell
minikube start --container-runtime=docker --vm=true
```

<h2 class="step"><span class="fa-stack fa-1x"><i class="fa fa-circle fa-stack-2x"></i><strong class="fa-stack-1x text-primary">3</strong></span>Point Docker CLI to minikube</h2>
Use the `minikube docker-env` command to point your terminal's Docker CLI to the Docker instance inside minikube.

<br>Note: the default profile name is `minikube`

{{% tabs %}}
{{% tab "bash/zsh" %}}
```
eval $(minikube -p <profile> docker-env)
```
{{% /tab %}}
{{% tab PowerShell %}}
```
& minikube -p <profile> docker-env --shell powershell | Invoke-Expression
```
{{% /tab %}}
{{% tab cmd %}}
```
@FOR /f "tokens=*" %i IN ('minikube -p <profile> docker-env --shell cmd') DO @%i
```
{{% /tab %}}
{{% tab fish %}}
```
minikube -p <profile> docker-env | source
```
{{% /tab %}}
{{% tab tcsh %}}
```
eval `minikube -p <profile> docker-env`
```
{{% /tab %}}
{{% /tabs %}}
