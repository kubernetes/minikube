# minikube roadmap (2019)

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2019, divided by how they apply to our [guiding principles](principles.md)

Please send a PR to suggest any improvements to it.

## (#1) User-friendly and accessible

- [ ] Creation of a user-centric minikube website for installation & documentation [#4388](https://github.com/kubernetes/minikube/issues/4388)
- [ ] Localized output to 5+ written languages [#4186](https://github.com/kubernetes/minikube/issues/4186) [#4185](https://github.com/kubernetes/minikube/issues/4185)
- [x] Make minikube usable in environments with challenging connectivity requirements
- [ ] Support lightweight deployment methods for environments where VM's are impractical [#4389](https://github.com/kubernetes/minikube/issues/4389) [#4390](https://github.com/kubernetes/minikube/issues/4390)
- [x] Add offline support

## (#2) Inclusive and community-driven

- [x] Increase community involvement in planning and decision making
- [ ] Make the continuous integration and release infrastructure publicly available [#3256](https://github.com/kubernetes/minikube/issues/4390)
- [x] Double the number of active maintainers

## (#3) Cross-platform

- [ ] Users should never need to separately install supporting binaries [#3975](https://github.com/kubernetes/minikube/issues/3975) [#4391](https://github.com/kubernetes/minikube/issues/4391)
- [ ] Simplified installation process across all supported platforms

## (#4) Support all Kubernetes features

- [ ] Add multi-node support [#94](https://github.com/kubernetes/minikube/issues/94)

## (#5) High-fidelity

- [ ] Reduce guest VM overhead by 50% [#3207](https://github.com/kubernetes/minikube/issues/3207)
- [x] Disable swap in the guest VM

## (#6) Compatible with all supported Kubernetes releases

- [x] Continuous Integration testing across all supported Kubernetes releases
- [ ] Automatic PR generation for updating the default Kubernetes release minikube uses [#4392](https://github.com/kubernetes/minikube/issues/4392)

## (#7) Support for all Kubernetes-friendly container runtimes

- [x] Run all integration tests across all supported container runtimes
- [ ] Support for Kata Containers [#4347](https://github.com/kubernetes/minikube/issues/4347)

## (#8) Stable and easy to debug

- [x] Pre-flight error checks for common connectivity and configuration errors
- [ ] Improve the `minikube status` command so that it can diagnose common issues
- [ ] Mark all features not covered by continuous integration as `experimental`
- [x] Stabilize and improve profiles support (AKA multi-cluster)
