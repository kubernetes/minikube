# Test Infrastructure Configuration

This directory contains modified configuration files from the [kubernetes/test-infra](https://github.com/kubernetes/test-infra) repository.

## Changes Made

### minikube-presubmits.yaml

Added `optional: true` to the following Prow presubmit jobs to make them optional rather than required:

- `integration-kvm-docker-linux-x86`
- `integration-kvm-containerd-linux-x86`

These changes need to be submitted as a PR to the kubernetes/test-infra repository at:
`config/jobs/kubernetes/minikube/minikube-presubmits.yaml`

## Rationale

Making these KVM tests optional allows PRs to be merged without waiting for these potentially resource-intensive tests to complete, while still maintaining the ability to run them on demand.
