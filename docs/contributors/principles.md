# Principles of Minikube

The primary goal of minikube is to make it simple to run Kubernetes locally, for day-to-day development workflows and learning purposes.

Here are some specific minikube features that align with our goal:

* Single command setup and teardown UX
* Support most portable Kubernetes core features (local storage, networking, auto-scaling, load balancing, etc.)
* Unified UX across OSes
* Minimal dependencies on third party software
* Minimal resource overhead

## Non-Goals

* Simplifying Kubernetes production deployment experience
  * Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.

## Priorities

These are the overall priorities of the minikube project, roughly in order:

1. Setting up a well-tested, secure and complete Kubernetes cluster locally.
2. Cross Platform support (macOS, Linux, Windows)
3. Supporting existing Kubernetes features, such as Load Balancer support and Persistent disks
4. Keeping up with new Kubernetes releases and features
5. Development-focused features like mounting host directories and support for VPN/proxy networking
6. Native hypervisor integration
7. Support for multiple container runtimes
8. User-friendly and accessible to non-technical users
