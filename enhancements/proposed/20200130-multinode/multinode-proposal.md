# Multi-node

* First proposed: 2020-01-31
* Authors: Sharif Elgamal (@sharifelgamal)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

## Summary

Until now minikube has always been a local single node Kubernetes cluster. Having multiple nodes in minikube clusters has been [the most requested feature](https://github.com/kubernetes/minikube/issues/94) in this history of the minikube repository.

## Goals

*   Enabling clusters with any number of control plane and worker nodes.
*   The ability to add and remove nodes from any cluster.
*   The ability to customize config per node.

## Non-Goals

*   Reproducing production environments

## Design Details

Since minikube was designed with only a single node cluster in mind, we need to make some fairly significant refactors, the biggest of which is the introduction of the Node object. Each cluster config will be able to have an abitrary number of Node objects, each of which will have attributes that can define it, similar to what [tstromberg proposed](https://github.com/kubernetes/minikube/pull/5874) but with better backwards compatibility with current config.

Each node will correspond to one VM (or container) and will connect back to the primary control plane via `kubeadm join`.

Also added will be the `node` sub command, e.g. `minikube node start` and `minikube node delete`. This will allow users to control their cluster however they please. Eventually, we will want to support passing in a `yaml` file into `minikube start` that defines all the nodes and their configs in one go. 

Users will be able to start multinode clusters in two ways:
1. `minikube start --nodes=2`
1. * `minikube start`
   * `minikube node add --name=node2`
   * `minikube node start --name=node2`

A note about `docker env`, the initial implementation won't properly support `docker env` in any consistent way, use at your own risk. The plan is to propagate any changes made to all the nodes in the cluster, with the caveat that anything that interrupts the command will cause a potentially corrupt cluster.

## Alternatives Considered

_TBD_
