---
title: "Roadmap"
date: 2019-07-31
weight: 4
description: >
  Development Roadmap
---

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2022, divided by how they apply to our [guiding principles]({{< ref "/docs/contrib/principles" >}})

Please send a PR to suggest any improvements to it.

# 2022

## (#1) GUI

- [x] Be able to start, stop, pause, and delete clusters via a GUI (prototype state)
- [x] Application available for all supported platforms: Linux, macOS, Windows

## (#2) Documentation

- [ ] Consolidate Kubernetes documentation that references minikube
- [ ] Delete outdated documentation
- [ ] Add documentation for new features

## (#3) ARM64 Support

- [x] Add Linux VM support
- [x] Add Mac M1 VM support (experimental, will improve by end of 2022)

## (#4) Docker
- [ ] Remove the Docker Desktop requirement on Mac and Windows
- [x] Continue supporting Docker as a container runtime (with CRI)

## (#5) libmachine Refactor

- [x] Add new driver (with QEMU) to replace HyperKit, primarily for Mac arm64 (experimental, will improve by end of 2022)
- [ ] Fix the provisioner, remove legacy Swarm, and add support for other runtimes
