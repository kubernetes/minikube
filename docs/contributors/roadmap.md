# minikube roadmap (2019)

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2019. Please send a PR to suggest any improvements to it.

## (#1) Make minikube the easiest way for developers to learn Kubernetes

- Single step installation
  - Users should not have to separately install supporting binaries
- Creation of a user-centric minikube website for installation & documentation

## (#2) Diversify the minikube community

- Add documentation for new minikube contributors
- Grow the number of maintainers
- Increase community involvement in planning and decision making

## (#3) Make minikube robust and easy to debug

- Pre-flight error checks for common connectivity and configuration errors
- Improve the `minikube status` command so that it can diagnose common issues
- Make minikube usable offline
- Mark features & options not covered by continuous integration as `experimental`

## (#4) Multi-node/multi-cluster support

- Stabilize and improve profiles support (multi-cluster)
- Introduce multi-node support

## (#5) Improve performance

- Reduce guest VM overhead by 50%
- Support lightweight deployment methods for environments where VM's are impractical

## (#6) Reduce technical debt

- Replace built-in machine drivers (virtualbox, kvm2) with their upstream equivalents
- Remove dependency on boot2docker (deprecated)
  