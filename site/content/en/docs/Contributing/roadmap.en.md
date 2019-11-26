---
title: "Roadmap"
date: 2019-07-31
weight: 4
description: >
  Development Roadmap
---

This roadmap is a living document outlining the major technical improvements which we would like to see in minikube during 2019, divided by how they apply to our [guiding principles](principles.md)

Please send a PR to suggest any improvements to it.

# 2019

## (#1) Inclusive and community-driven

- [x] Increase community involvement in planning and decision making
- [x] Double the number of active maintainers
- [ ] Make the continuous integration and release infrastructure publicly available

## (#2) User-friendly and accessible

- [x] Creation of a user-centric minikube website for installation & documentation
- [x] Make minikube usable in environments with challenging connectivity requirements
- [x] Localized output to 5+ written languages

## (#3) Support all Kubernetes features

- [x] Continuous Integration testing across all supported Kubernetes releases
- [x] Run all integration tests across all supported container runtimes
- [ ] Add multi-node support
- [ ] Automatic PR generation for updating the default Kubernetes release minikube uses

## (#4) Cross-platform

- [x] Simplified installation process across all supported platforms
- [x] Users should never need to separately install supporting binaries
- [ ] Support lightweight deployment methods for environments where VM's are impractical

## (#5) Reliable

- [x] Pre-flight error checks for common connectivity and configuration errors
- [x] Stabilize and improve profiles support (AKA multi-cluster)
- [ ] Improve the `minikube status` command so that it can diagnose common issues

## (#6) High Performance

- [ ] Reduce guest VM overhead by 50%
- [x] Disable swap in the guest VM

## (#7) Developer focused

- [x] Add offline support

# 2020 (draft)

## (#1) Inclusive and community-driven

- [ ] Maintainers from 4 countries, 4 companies
- [ ] Installation documentation in 5+ written languages
- [ ] Enhancements approved by a community-driven process

## (#2) User-friendly

- [ ] Automatic installation of hypervisor dependencies
- [ ] Graphical User Interface
- [ ] Built-in 3rd Party ecosystem with 50+ entries

## (#3) Support all Kubernetes features

- [ ] Multi-node
- [ ] IPv6
- [ ] Usage documentation for 3 leading CNI providers
- [ ] Automatically publish conformance test results after a release

## (#4) Cross-platform

- [ ] VM-free deployment to containers (Docker, Podman)
- [ ] Windows as a first-class citizen
- [ ] WSL2 support (no additional VM required)
- [ ] Firecracker VM support
- [ ] Generic (SSH) driver support

## (#5) Reliable

- [ ] Resource alerts
- [ ] Time synchronization on HyperKit
- [ ] Prototype post-libmachine implementation of minikube

## (#6) High Performance

- [ ] Startup latency under 30s
- [ ] Kernel-assisted mounts (CIFS, NFS) by default
- [ ] Suspend and Resume
- [ ] <25% CPU overhead on a single core

## (#7) Developer Focused

- [ ] Container build integration
- [ ] Documented workflow for Kubernetes development
