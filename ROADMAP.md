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
* Simplifying kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Priorities
This section contains the overall priorities of the minikube project, in rough order.

 * Setting up a well-tested, secure and complete Kubernetes cluster locally.
 * Mac OSX and Linux support.
 * Supporting existing Kubernetes features:
    * Load Balancer support.
    * Persistent disks.
 * Keeping up with new Kubernetes releases and features.
 * Development-focused features like:
   * Mounting host directories.
   * VPN/proxy networking.
 * Windows support.
 * Native hypervisor integration.
 * Support for alternative Kubernetes runtimes, like rkt.
 * Removing the VirtualBox dependency and replacing it with Hypervisor.framework/Hyper-V.
 * Support for multiple nodes.

## Timelines
These are rough dates, on a 3-month schedule. Minikube will release much faster than this, so this section is fairly speculative.
This section is subject to change based on feedback and staffing.

### June 2016
 * Fully-tested, complete release of minikube that supports:
   * Mac OSX and Linux.
   * Kubernetes 1.3.
   * Docker 1.11.x.
   * VirtualBox.

### September 2016
 * Support for Windows.
 * Kubernetes 1.4, Docker 1.x.y.
 * Host Directory mounting.
 * Improved networking (Ingress, proxies, VPN...).

### December 2016
 * Native hypervisor integration (Hypervisor.framework for OSX, Hyper-V for Windows).
 * Support Rkt.
 * Remove hypervisor on Linux systems.
