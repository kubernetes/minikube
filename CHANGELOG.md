# Minikube Release Notes

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
 * Generated documentation for all commands. Documentation [is here](https://github.com/kubernetes/minikube/blob/master/docs/minikube.md).


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
