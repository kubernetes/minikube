# Minikube Release Notes

## Version 0.3.0 - 6/20/2016
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
