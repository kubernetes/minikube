---
title: "Roadmap"
date: 2019-07-31
weight: 4
description: >
  Development Roadmap
---

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2020, divided by how they apply to our [guiding principles]({{< ref "/docs/contrib/principles" >}})

Please send a PR to suggest any improvements to it.

# 2020 

## (#1) Inclusive and community-driven

- [x] Maintainers from 4 countries, 4 companies
- [ ] Installation documentation in 5+ written languages
- [x] Enhancements approved by a community-driven process

## (#2) User-friendly

- [ ] Automatic installation of hypervisor dependencies
- [ ] Graphical User Interface
- [ ] Built-in 3rd Party ecosystem with 50+ entries

## (#3) Support all Kubernetes features

- [x] Multi-node
- [ ] IPv6
- [ ] Usage documentation for 3 leading CNI providers
- [ ] Automatically publish conformance test results after a release

## (#4) Cross-platform

- [x] VM-free deployment to containers (Docker, Podman)
- [x] Windows as a first-class citizen
- [x] WSL2 support (no additional VM required)
- [ ] Firecracker VM support
- [ ] Generic (SSH) driver support

## (#5) Reliable

- [ ] Resource alerts
- [ ] Time synchronization on HyperKit
- [ ] Prototype post-libmachine implementation of minikube

## (#6) High Performance

- [x] Startup latency under 30s
- [ ] Kernel-assisted mounts (CIFS, NFS) by default
- [x] Pause support
- [x] <25% CPU overhead on a single core

## (#7) Developer Focused

- [ ] Container build integration
- [ ] Documented workflow for Kubernetes development
