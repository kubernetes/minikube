# Minikube Release Notes

## Version 0.6.0 - 7/13/2016
* Added...

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
