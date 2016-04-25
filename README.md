# Minikube

## Background

Minikube is a tool that makes it easy to run Kubernetes locally. Minikube runs
a single-node Kubernetes cluster inside a VM on your laptop for users looking
to try out Kubernetes or develop with it day-to-day.

## Design

Minikube uses [libmachine](https://github.com/docker/machine/tree/master/libmachine) for provisioning VMs, and [localkube](https://github.com/redspread/localkube)
for running the cluster.

For more information about minikube, see the [proposal](https://github.com/kubernetes/kubernetes/blob/master/docs/proposals/local-cluster-ux.md).

## Goals

* Works across multiple OSes - OS X, Linux and Windows primarily.
* Single command setup and teardown UX.
* Unified UX across OSes
* Minimal dependencies on third party software.
* Minimal resource overhead.
* Replace any other alternatives to local cluster deployment.

## Non Goals

* Simplifying kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.
