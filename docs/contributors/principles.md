# Principles of Minikube

The primary goal of minikube is to make it simple to run Kubernetes locally, for day-to-day development workflows and learning purposes. Here are the guiding principles for minikube, in rough priority order:

1. User-friendly and accessible
2. Inclusive and community-driven
3. Cross-platform
4. Support all Kubernetes features
5. High-fidelity
6. Compatible with all supported Kubernetes releases
7. Support for all Kubernetes-friendly container runtimes
8. Stable and easy to debug

Here are some specific minikube features that align with our goal:

* Single command setup and teardown UX
* Support for local storage, networking, auto-scaling, load balancing, etc.
* Unified UX across operating systems
* Minimal dependencies on third party software
* Minimal resource overhead

## Non-Goals

* Simplifying Kubernetes production deployment experience
* Supporting all possible deployment configurations of Kubernetes like various types of storage, networking, etc.
