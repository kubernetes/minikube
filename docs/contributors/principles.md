# Principles of Minikube

The primary goal of minikube is to make it simple to run Kubernetes on your local machine, both for getting started and day-to-day development workflows.

Here are some specific features that align with our goal:

* Single command setup and teardown UX.
* Support most portable Kubernetes core features (local storage, networking, auto-scaling, loadbalancing, etc.)
* Unified UX across OSes.
* Minimal dependencies on third party software.
* Minimal resource overhead.
* Becoming the default local-cluster setup for Kubernetes

## Non-Goals
* Simplifying Kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.

* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Priorities

This section contains the overall priorities of the minikube project, in rough order.

 * Setting up a well-tested, secure and complete Kubernetes cluster locally.
 * Cross Platform support (macOS, Linux, Windows)
 * Supporting existing Kubernetes features:
    * Load Balancer support.
    * Persistent disks.
 * Keeping up with new Kubernetes releases and features.
 * Development-focused features like:
   * Mounting host directories.
   * VPN/proxy networking.
 * Native hypervisor integration.
 * Support for alternative Kubernetes runtimes, like rkt.
