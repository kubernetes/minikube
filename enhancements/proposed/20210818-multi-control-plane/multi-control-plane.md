# Multiple Control Plane

* First proposed: 2021-08-18
* Authors: Ling Sameul (@lingsamuel)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

## Summary

Minikube now supports multiple worker nodes. But multiple control plane nodes haven't been implemented yet. As part of multinode MEP, this proposal discusses the details of multiple control plane implementation.

## Goals

*   Supports add multiple control planes.
*   The ability to customize config per node.
*   Interact correctly with other features and be user-friendly.

## Non-Goals

*   Reproducing production environments
*   100% backward compatibility

## Design Details

Users should be able to manage control planes at any time. Typical scenarios are as follows:
1. `minikube start --control-planes=2`, the `nodes` parameter will be automatically scaled up if `nodes < control-planes`
2. `minikube node add --control-plane` to add a control plane

Users may use the subcommand `node` to stop or delete a control plane node. In both situations, the node will leave the control plane via `kubeadm reset`. Once the stopped node starts again, it will re-join the control plane via `kubeadm join`.

Since control planes need to communicate with each other at startup, the control planes VM should theoretically start in parallel. However, this approach may increase the complexity of the startup process and logs. To avoid this, the first control planes will be marked as primary. All non-primary control planes will join the primary control plane.

Several backward compatibility cases:
1. start a stopped old, single control plane cluster
2. use the `node` subcommand to manage a running single control plane cluster
3. updates configuration files if necessary

There is no promise of compatibility with any old cluster configuration. There is also no guarantee that older minikube will be compatible with newer configuration.

Also, the `status` command will output an extra IP field that will help users to write scripts to manage their cluster. Primary control plane will be marked.

The existing subcommand `docker-env` way cannot propagate changes to all nodes since it only outputs docker environment variables. In prototype multinode implementation, it will output the primary control plane information by default, or specify nodes by name.

## Alternatives Considered

_TBD_
