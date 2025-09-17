---
title: "Configuration"
weight: 4
description: >
  Configuring your cluster
aliases:
  - /docs/reference/environment_variables/
  - /docs/reference/configuration/kubernetes/
  - /docs/reference/runtimes
---

## Basic Configuration

Most minikube configuration is done via the flags interface. To see which flags are possible for the start command, run:

```shell
minikube start --help
```

## Persistent Configuration

minikube allows users to persistently store new default values to be used across all profiles, using the `minikube config` command. This is done providing a property name, and a property value.

For example, to persistently configure minikube to use hyperkit:

```shell
minikube config set driver hyperkit
```

You can get a complete list of configurable fields using:

```shell
minikube config --help
```

To get a list of the currently set config properties:

```shell
minikube config view
```

## Kubernetes configuration

minikube allows users to configure the Kubernetes components with arbitrary values. To use this feature, you can use the `--extra-config` flag on the `minikube start` command.

This flag is repeated, so you can pass it several times with several different values to set multiple options.

### Selecting a Kubernetes version

By default, minikube installs the latest stable version of Kubernetes that was available at the time of the minikube release. You may select a different Kubernetes release by using the `--kubernetes-version` flag, for example:

```shell
minikube start --kubernetes-version=v1.34.0
```

minikube follows the [Kubernetes Version and Version Skew Support Policy](https://kubernetes.io/docs/setup/version-skew-policy/), so we guarantee support for the latest build for the last 3 minor Kubernetes releases. When practical, minikube aims to support older releases as well so that users can emulate legacy environments.

For up to date information on supported versions, see `OldestKubernetesVersion` and `NewestKubernetesVersion` in [constants.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/constants/constants.go)

### Enabling feature gates

Kubernetes alpha/experimental features can be enabled or disabled by the `--feature-gates` flag on the `minikube start` command. It takes a string of the form `key=value` where key is the `component` name and value is the `status` of it.

```shell
minikube start --feature-gates=EphemeralContainers=true
```

### Modifying Kubernetes defaults

The kubeadm bootstrapper can be configured by the `--extra-config` flag on the `minikube start` command.  It takes a string of the form `component.key=value` where `component` is one of the strings

* kubeadm
* kubelet
* apiserver
* controller-manager
* scheduler

and `key=value` is a flag=value pair for the component being configured.  For example,

```shell
minikube start --extra-config=apiserver.v=10 --extra-config=kubelet.max-pods=100
```

For instance, to allow Kubernetes to launch on an unsupported Docker release:

```shell
minikube start --extra-config=kubeadm.ignore-preflight-errors=SystemVerification
```

## Runtime configuration

The default container runtime in minikube varies. You can select one explicitly by using:

```shell
minikube start --container-runtime=docker
```

Options available are:

* [containerd]({{<ref "/docs/runtimes/containerd">}})
* [cri-o]({{<ref "/docs/runtimes/cri-o">}})
* [docker]({{<ref "/docs/runtimes/docker">}})

See <https://kubernetes.io/docs/setup/production-environment/container-runtimes/>

## Environment variables

minikube supports passing environment variables instead of flags for every value listed in `minikube config`.  This is done by passing an environment variable with the prefix `MINIKUBE_`.

For example the `minikube start --iso-url="$ISO_URL"` flag can also be set by setting the `MINIKUBE_ISO_URL="$ISO_URL"` environment variable.

### Exclusive environment tunings

Some features can only be accessed by minikube specific environment variables, here is a list of these features:

* **MINIKUBE_HOME** - (string) sets the path for the .minikube directory that minikube uses for state/configuration. If you specify it to `/path/to/somewhere` and `somewhere` is not equal to `.minikube`, the final MINIKUBE_HOME will be `/path/to/somewhere/.minikube`. Defaults to `~/.minikube` if unspecified. *Please note: this is used only by minikube and does not affect anything related to Kubernetes tools such as kubectl.*

* **MINIKUBE_IN_STYLE** - (bool) manually sets whether or not emoji and colors should appear in minikube. Set to false or 0 to disable this feature, true or 1 to force it to be turned on.

* **CHANGE_MINIKUBE_NONE_USER** - (bool) automatically change ownership of ~/.minikube to the value of $SUDO_USER

* **MINIKUBE_ENABLE_PROFILING** - (int, `1` enables it) enables trace profiling to be generated for minikube

* **MINIKUBE_SUPPRESS_DOCKER_PERFORMANCE** - (bool) suppresses Docker performance warnings when Docker is slow

### Example: Disabling emoji

{{% tabs %}}

{{% linuxtab %}}

```shell
export MINIKUBE_IN_STYLE=false
minikube start
```

{{% /linuxtab %}}

{{% mactab %}}

```shell
export MINIKUBE_IN_STYLE=false
minikube start
```

{{% /mactab %}}

{{% windowstab %}}

```shell
$env:MINIKUBE_IN_STYLE=false
minikube start
```

{{% /windowstab %}}
{{% /tabs %}}

### Making environment values persistent

To make the exported variables persistent across reboots:

* Linux and macOS: Add these declarations to `~/.bashrc` or wherever your shells environment variables are stored.
* Windows: Either add these declarations to your `~\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1` or run the following in a PowerShell terminal:
```shell
[Environment]::SetEnvironmentVariable("key", "value", [EnvironmentVariableTarget]::User)
```
