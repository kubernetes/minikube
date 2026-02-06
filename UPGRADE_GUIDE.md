# Minikube Upgrade Guide

## From v1.37.x to v1.38.0

### Notable Changes:
- Support for Kubernetes v1.35.0
- No sudo required for vfkit/krunkit on macOS 26+
- Hyperkit driver removal warning

### Steps:
1. Backup any custom configurations.
2. Stop and delete existing clusters with `minikube delete`.
3. Download and install v1.38.0 from the official GitHub releases page.
4. Verify installation with `minikube version`.

### Deprecations:
- 32-bit architecture support removed.
- Hyperkit driver will be removed in v1.39.0; switch to Docker or kvm2 now.

### Resources:
- [Release Notes](https://github.com/kubernetes/minikube/releases/tag/v1.38.0)