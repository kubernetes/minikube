# Minikube Roadmap
This document contains the goals, plans, and priorities for the minikube project.

## Goals
The primary goal of minikube is to make it simple to run Kubernetes on your laptop, both for getting started and day-to-day development workflows.
Here are some specific features that align with our goal:
* Single command setup and teardown UX.
* Unified UX across OSes
* Minimal dependencies on third party software.
* Minimal resource overhead.
* Replace any other alternatives to local cluster deployment.

## Non-Goals
* Simplifying kubernetes production deployment experience. Kube-deploy is attempting to tackle this problem.
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Priorities
This section contains the overall priorities of the minikube project, in rough order.

 * Setting up a well-tested, secure and complete Kubernetes cluster locally.
 * Keeping up with new Kubernetes releases and features.
   * Load Balancer support
   * Persistent disks
 * Mac OSX and Linux support initially, Windows support later.
 * Development-focused features like:
   * Mounting host directories
   * VPN/proxy networking.
 * Support for alternative Kubernetes runtimes, like rkt.
 * Removing the VirtualBox dependency and replacing it with Hypervisor.framework/Hyper-V.
 * Support for multiple nodes.

## Timelines
These are rough dates, on a 3-month schedule. Minikube will release much faster than this, so this section is fairly speculative.
This section is subject to change based on feedback and staffing.

### June 2016
 * Fully-tested, complete release of minikube that supports:
   * Mac OSX and Linux
   * Kubernetes 1.3
   * Docker 1.11.x
   * VirtualBox

### September 2016
 * Support for Windows
 * Kubernetes 1.4, Docker 1.x.y
 * Host Directory mounting
 * Improved networking

### December 2016
 * Native hypervisor integration (Hypervisor.framework, Hyper-V)
 *
