---
title: "Roadmap"
date: 2019-07-31
weight: 4
description: >
  Development Roadmap
---

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube over the next year, divided by how they apply to our [guiding principles]({{< ref "/docs/contrib/principles" >}})

Please send a PR to suggest any improvements to it.

# 2024

## (#1) AI

- [ ] Support Nvidia-Docker runtime
- [ ] Try Kubeflow addon

## (#2) Documentation

- [x] Consolidate Kubernetes documentation that references minikube
- [ ] Delete outdated documentation
- [ ] Add documentation for new features

## (#3) Docker
- [ ] Remove the Docker Desktop requirement on Mac and Windows
- [ ] Support the Docker Desktop environment on Linux as well

## (#4) Podman
- [ ] Improve support for rootless containers with Podman Engine
- [ ] Support the Podman Desktop environment on Mac and Windows

## (#5) libmachine Refactor

- [ ] Add new driver with Virtualization.framework, as QEMU alternative on Mac arm64
- [ ] Fix the provisioner, remove legacy Swarm, and add support for other runtimes
