# minikube roadmap (2019)

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2019, divided by how they apply to our [guiding principles](principles.md)

Please send a PR to suggest any improvements to it.

## (#1) User-friendly and accessible

- [ ] Creation of a user-centric minikube website for installation & documentation
- [ ] Localized output to 5+ written languages
- [ ] Make minikube usable in environments with challenging connectivity requirements
- [ ] Support lightweight deployment methods for environments where VM's are impractical
- [x] Add offline support

## (#2) Inclusive and community-driven

- [x] Increase community involvement in planning and decision making
- [ ] Make the continuous integration and release infrastructure publicly available
- [x] Double the number of active maintainers

## (#3) Cross-platform

- [ ] Simplified installation process across all supported platforms
- [ ] Users should never need to separately install supporting binaries

## (#4) Support all Kubernetes features

- [ ] Add multi-node support

## (#5) High-fidelity

- [ ] Reduce guest VM overhead by 50%
- [x] Disable swap in the guest VM

## (#6) Compatible with all supported Kubernetes releases

- [x] Continuous Integration testing across all supported Kubernetes releases
- [ ] Automatic PR generation for updating the default Kubernetes release minikube uses

## (#7) Support for all Kubernetes-friendly container runtimes

- [x] Run all integration tests across all supported container runtimes
- [ ] Support for Kata Containers (help wanted!)

## (#8) Stable and easy to debug

- [ ] Pre-flight error checks for common connectivity and configuration errors
- [ ] Improve the `minikube status` command so that it can diagnose common issues
- [ ] Mark all features not covered by continuous integration as `experimental`
- [x] Stabilize and improve profiles support (AKA multi-cluster)
