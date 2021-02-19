# Standard Linux Distribution

* First proposed: 2020-12-17
* Authors: Anders F Björklund (@afbjorklund)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Change the distribution (OS) for the minikube ISO, from Buildroot to Ubuntu.

## Goals

* Use one of the supported Kubernetes OS, like Ubuntu 20.04
* Use the same operating system for KIC base and ISO image

## Non-Goals

* Making major changes to the new standard operating system
* Support production deployments, still intended for learning

## Design Details

Use external system image and external packages, same as for KIC image.

Keep both images available (one being default), during transition period.

## Alternatives Considered

Continue to support custom distro, instead of using a standard distro.

Make current Buildroot OS into standard supported Kubernetes distribution.
