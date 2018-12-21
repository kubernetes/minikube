# minikube roadmap (2019)

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2019. Please send a PR to suggest any improvements to it.

## (#1) Make minikube the easiest way for developers to learn Kubernetes

- Single step installation
  - Users should not have to seperately install supporting binaries
- Creation of a user-centric minikube website for installation & documentation

## (#2) Diversify the minikube community

- Add documentation for new minikube contributors
- Ensure that all decisions are ratified publically by the minikube community

## (#3) Make minikube robust and debuggable

- Add pre-flight error checks for common connectivity issues 
- Add pre-flight error checks for common configuration errors
- Mark features & options not covered by continuous integration as `experimental`
- Improve the `status` command so that it can diagnose common environmental issues
- Make minikube usable offline

## (#4) Official multi-node/multi-cluster support

- Rebrand profiles as multi-cluster support
- Integrate KIND for multi-node support
- Add commands to add/remove nodes within an existing cluster

## (#5) Improve minikube performance

- Add support for lighter-weight deployment methods, such as container-based (LXD, Docker) or chroot
- Reduce guest VM overhead by 50% for macOS users
