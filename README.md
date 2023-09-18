README
`"# minikubot

[![Actions Status](https://github.com/kubernetes/minikube/workflows/build/badge.svg)](https://github.com/kubernetes/minikube/actions)
[![GoReport Widget]][GoReport Status]
[![GitHub All Releases](https://img.shields.io/github/downloads/kubernetes/minikube/total.svg)](https://github.com/kubernetes/minikube/releases/latest)
[![Latest Release](https://img.shields.io/github/v/release/kubernetes/minikube?include_prereleases)](https://github.com/kubernetes/minikube/releases/latest)
 

[GoReport Status]: https://goreportcard.com/report/github.com/kubernetes/minikube
[GoReport Widget]: https://goreportcard.com/badge/github.com/kubernetes/minikube

<img src="https://github.com/kubernetes/minikube/raw/master/images/logo/logo.png" width="100" alt="minikube logo">

minikube implements a local Kubernetes cluster on macOS, Linux, and Windows. minikube's [primary goals](https://minikube.sigs.k8s.io/docs/concepts/principles/) are to be the best tool for local Kubernetes application development and to support all Kubernetes features that fit. 

<img src="https://raw.githubusercontent.com/kubernetes/minikube/master/site/static/images/screenshot.png" width="575" height="322" alt="screenshot">

## Features

minikube runs the latest stable release of Kubernetes, with support for standard Kubernetes features like:

* [LoadBalancer](https://minikube.sigs.k8s.io/docs/handbook/accessing/#loadbalancer-access) - using `minikube tunnel`
* Multi-cluster - using `minikube start -p <name>`
* [NodePorts](https://minikube.sigs.k8s.io/docs/handbook/accessing/#nodeport-access) - using `minikube service`
* [Persistent Volumes](https://minikube.sigs.k8s.io/docs/handbook/persistent_volumes/)
* [Ingress](https://kubernetes.io/docs/tasks/access-application-cluster/ingress-minikube/)
* [Dashboard](https://minikube.sigs.k8s.io/docs/handbook/dashboard/) - `minikube dashboard`
* [Container runtimes](https://minikube.sigs.k8s.io/docs/handbook/config/#runtime-configuration) - `minikube start --container-runtime`
* [Configure apiserver and kubelet options](https://minikube.sigs.k8s.io/docs/handbook/config/#modifying-kubernetes-defaults) via command-line flags
* Supports common [CI environments](https://github.com/minikube-ci/examples)

As well as developer-friendly features:

* [Addons](https://minikube.sigs.k8s.io/docs/handbook/deploying/#addons) - a marketplace for developers to share configurations for running services on minikube
* [NVIDIA GPU support](https://minikube.sigs.k8s.io/docs/handbook/addons/nvidia/) - for machine learning
* [Filesystem mounts](https://minikube.sigs.k8s.io/docs/handbook/mount/)

**For more information, see the official [minikube website](https://minikube.sigs.k8s.io)**

## Installation

See the [Getting Started Guide](https://minikube.sigs.k8s.io/docs/start/)

:mega: **Please fill out our [fast 5-question survey](https://forms.gle/Gg3hG5ZySw8c1C24A)** so that we can learn how & why you use minikube, and what improvements we should make. Thank you! :dancers:

## Documentation

See https://minikube.sigs.k8s.io/docs/

## More Examples

See minikube in action [here](https://minikube.sigs.k8s.io/docs/handbook/controls/)

## Community

minikube is a Kubernetes [#sig-cluster-lifecycle](https://github.com/kubernetes/community/tree/master/sig-cluster-lifecycle)  project.

* [**#minikube on Kubernetes Slack**](https://kubernetes.slack.com/messages/minikube) - Live chat with minikube developers!
* [minikube-users mailing list](https://groups.google.com/g/minikube-users)
* [minikube-dev mailing list](https://groups.google.com/g/minikube-dev)

### RELEASE ###
# Release Notes

## Version 1.31.2 - 2023-08-16

docker-env Regression:
* Create `~/.ssh` directory if missing [#16934](https://github.com/kubernetes/minikube/pull/16934)
* Fix adding guest to `~/.ssh/known_hosts` when not needed [#17030](https://github.com/kubernetes/minikube/pull/17030)

Minor Improvements:
* Verify containerd storage separately from docker [#16972](https://github.com/kubernetes/minikube/pull/16972)

Version Upgrades:
* Bump Kubernetes version default: v1.27.4 and latest: v1.28.0-rc.1 [#17011](https://github.com/kubernetes/minikube/pull/17011) [#17051](https://github.com/kubernetes/minikube/pull/17051)
* Addon cloud-spanner: Update cloud-spanner-emulator/emulator image from 1.5.7 to 1.5.9 [#17017](https://github.com/kubernetes/minikube/pull/17017) [#17044](https://github.com/kubernetes/minikube/pull/17044)
* Addon headlamp: Update headlamp-k8s/headlamp image from v0.18.0 to v0.19.0 [#16992](https://github.com/kubernetes/minikube/pull/16992)
* Addon inspektor-gadget: Update inspektor-gadget image from v0.18.1 to v0.19.0 [#17016](https://github.com/kubernetes/minikube/pull/17016)
* Addon metrics-server: Update metrics-server/metrics-server image from v0.6.3 to v0.6.4 [#16969](https://github.com/kubernetes/minikube/pull/16969)
* CNI flannel: Update from v0.22.0 to v0.22.1 [#16968](https://github.com/kubernetes/minikube/pull/16968)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Alex Serbul
- Anders F Björklund
- Jeff MAURY
- Medya Ghazizadeh
- Michelle Thompson
- Predrag Rogic
- Seth Rylan Gainey
- Steven Powell
- aiyijing
- joaquimrocha
- renyanda
- shixiuguo
- sunyuxuan
- Товарищ программист

Thank you to our PR reviewers for this release!

- medyagh (8 comments)
- spowelljr (2 comments)
- ComradeProgrammer (1 comments)
- Lyllt8 (1 comments)
- aiyijing (1 comments)

Thank you to our triage members for this release!

- afbjorklund (6 comments)
- vaibhav2107 (5 comments)
- kundan2707 (3 comments)
- spowelljr (3 comments)
- ao390 (2 comments)

Check out our [contributions leaderboard](https://minikube.sigs.k8s.io/docs/contrib/leaderboard/v1.31.2/) for this release!

## Version 1.31.1 - 2023-07-20

* cni: Fix regression in auto selection [#16912](https://github.com/kubernetes/minikube/pull/16912)

For a more detailed changelog, see [CHANGELOG.md](https://github.com/kubernetes/minikube/blob/master/CHANGELOG.md).

Thank you to our contributors for this release!

- Jeff MAURY
- Medya Ghazizadeh
- Steven Powell

Thank you to our triage members for this release!

- afbjorklund (5 comments)
- torenware (5 comments)
- mprimeaux (3 comments)
- prezha (3 comments)
- spowelljr (1 comments)

* Fixed dashboard command
* Added support for net.IP to the configurator
* Updated dashboard version to 1.4.2

## Version 0.12.1 - 10/28/2016

* Added docker-env support to the buildroot provisioner
* `minikube service` command now supports multiple ports
* Added `minikube service list` command
* Added `minikube completion bash` command to generate bash completion
* Add progress bars for downloading, switch to go-download
* Run kube-dns as addon instead of vendored in kube2sky
* Remove static UUID for xhyve driver
* Add option to specify network name for KVM

## Version 0.12.0 - 10/21/2016

* Added support for the KUBECONFIG env var during 'minikube start'
* Updated default k8s version to v1.4.3
* Updated addon-manager to v5.1
* Added `config view` subcommand
* Increased memory default to 2048 and cpus default to 2
* Set default `log_dir` to `~/.minikube/logs`
* Added `minikube addons` command to enable or disable cluster addons
* Added format flag to service command
* Added flag Hyper-v Virtual Switch
* Added support for IPv6 addresses in docker env

## Version 0.11.0 - 10/6/2016

* Added a "configurator" allowing users to configure the Kubernetes components with arbitrary values.
* Made Kubernetes v1.4.0 the default version in minikube
* Pre-built binaries are now built with go 1.7.1
* Added opt-in error reporting
* Bug fixes

## Version 0.10.0 - 9/15/2016

* Updated the Kubernetes dashboard to v1.4.0
* Added experimental rkt support
* Enabled DynamicProvisioning of volumes
* Improved the output of the `minikube status` command
* Added `minikube config get` and `minikube config set` commands
* Fixed a bug ensuring that the node IP is routable
* Renamed the created VM from minikubeVM to minikube

## Version 0.9.0 - 9/1/2016

* Added Hyper-V support for Windows
* Added debug-level logging for show-libmachine-logs
* Added ISO checksum validation for cached ISOs
* New .minikube/addons directory where users can put addons to be initialized in minikube
* --https flag on `minikube service` for services that run over ssl/tls
* xhyve driver will now receive the same IP across starts/delete

## Version 0.8.0 - 8/17/2016

* Added a --registry-mirror flag to `minikube start`.
* Updated Kubernetes components to v1.3.5.
* Changed the `dashboard` and `service` commands to wait for the underlying services to be ready.
* Added the `DOCKER_API_VERSION` environment variable to `minikube docker-env`.
* Updated the Kubernetes dashboard to v1.1.1.
* Improved error messages during `minikube start`.
* Added the ability to specify a CIDR for the virtualbox driver.
* Configured the `/data` directory inside the Minikube VM to be persisted across reboots.
* Added the ability for minikube to accept environment variables of the form `MINIKUBE_` in place of certain command line flags.
* Minikube will now cache downloaded localkube versions.

## Version 0.7.1 - 7/27/2016

* Fixed a filepath issue which caused `minikube start` to not work properly on Windows

## Version 0.7.0 - 7/26/2016

* Added experimental support for Windows.
* Changed the etc DNS port to avoid a conflict with deis/router.
* Added a `insecure-registry` flag to `minikube start` to support insecure docker registries.
* Added a `--docker-env` flag to `minikube start` which allows for environment variables to be passed to the Docker daemon.
* Updated Kubernetes components to 1.3.3.
* Enabled all available (including alpha) Kubernetes APIs.
* Added ISO caching.
* Added a `--unset` flag to `minikube docker-env` to unset the environment variables.
* Added a `--no-proxy` flag to `minikube docker-env` to add a machine IP to NO_PROXY environment variable.
* Added additional supported shells for `minikube docker-env` (fish, cmd, powershell, tcsh, bash, zsh).

## Version 0.6.0 - 7/13/2016

* Added a `--disk-size` flag to `minikube start`.
* Fixed a bug regarding auth tokens not being reconfigured properly after VM restart
* Added a new `get-k8s-versions` command, to get the available kubernetes versions so that users know what versions are available when trying to select the kubernetes version to use
* Makefile Updates
* Documentation Updates

## Version 0.5.0 - 7/6/2016

* Updated Kubernetes components to v1.3.0
* Added experimental support for KVM and XHyve based drivers. See the [drivers documentation](DRIVERS.md) for usage.
* Fixed a bug causing cluster state to be deleted after a `minikube stop`.
* Fixed a bug causing the minikube logs to fill up rapidly.
* Added a new `minikube service` command, to open a browser to the URL for a given service.
* Added a `--cpus` flag to `minikube start`.

## Version 0.4.0 - 6/27/2016

* Updated Kubernetes components to v1.3.0-beta.1
* Updated the Kubernetes Dashboard to v1.1.0
* Added a check for updates to minikube.
* Added a driver for VMWare Fusion on OSX.
* Added a flag to customize the memory of the minikube VM.
* Documentation updates
* Fixed a bug in Docker certificate generation. Certificates will now be
  regenerated whenever `minikube start` is run.

## Version 0.3.0 - 6/10/2016

* Added a `minikube dashboard` command to open the Kubernetes Dashboard.
* Updated Docker to version 1.11.1.
* Updated Kubernetes components to v1.3.0-alpha.5-330-g760c563.
* Generated documentation for all commands. Documentation [is here](https://minikube.sigs.k8s.io/docs/).

## Version 0.2.0 - 6/3/2016

* conntrack is now bundled in the ISO.
* DNS is now working.
* Minikube now uses the iptables based proxy mode.
* Internal libmachine logging is now hidden by default.
* There is a new `minikube ssh` command to ssh into the minikube VM.
* Dramatically improved integration test coverage
* Switched to glog instead of fmt.Print*

## Version 0.1.0 - 5/29/2016

* Initial minikube release.
* [Contributing](https://minikube.sigs.k8s.io/docs/contrib/)
* [Development Roadmap](https://minikube.sigs.k8s.io/docs/contrib/roadmap/)

Join our meetings:
* [Bi-weekly office hours, Mondays @ 11am PST](https://tinyurl.com/minikube-oh)
* [Triage Party](https://minikube.sigs.k8s.io/docs/contrib/triage/)"`
