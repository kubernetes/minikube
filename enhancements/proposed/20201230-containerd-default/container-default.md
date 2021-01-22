# CRI: Containerd by default

* First proposed: 2020-11-08
* Authors: Anders F Björklund (@afbjorklund)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Change default container runtime, from current "docker" to replacement "containerd".

## Goals

* Change from docker to containerd as default
* Still allow fast-building images with minikube

## Non-Goals

* Remove the docker support from minikube
* Change anything in the docker driver

## Design Details

The containerd container runtime is already included, and is passing certification.

Unlike [Docker](https://www.docker.com/products/docker-engine), will need to include the CRI (runtime) and CNI (network) by default.

Use [BuildKit](https://github.com/moby/buildkit) as a complement to [Containerd](https://containerd.io/), for producing an image from Dockerfile.

Only run buildkitd on-demand (i.e. when building), default to running only containerd.

## Alternatives Considered

Keep Docker as the default, and add the new CRI-Docker to replace the old dockershim.

Use [CRI-O](https://cri-o.io/)/[Podman](https://podman.io/) as default, which is a bigger change (since dockerd uses containerd).
