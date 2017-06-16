# Minikube Roadmap
This document contains the goals, plans, and priorities for the minikube project.
Note that these priorities are not set in stone. Please file an issue if you'd like to discuss adding or reordering these :)

## Goals
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
 * Removing the VirtualBox dependency and replacing it with Hypervisor.framework/Hyper-V.

## Timelines
Minikube will release much faster than this, so this section is fairly speculative.
This section is subject to change based on feedback and staffing.

### Q1 2017

* Release Kubernetes 1.6.0 alpha and beta releases packaged with minikube
* Release Kubernetes 1.6.0 packaged with minikube within two days of GA upstream build
* Run local e2e Kubernetes tests with minikube
* Minikube no longer depends on libmachine
* Minikube no longer depends on existing KVM driver
* Native drivers are made default and packaged with minikube
* Improve minikube start time by 30%
* Add a no-vm driver for linux CI environments
