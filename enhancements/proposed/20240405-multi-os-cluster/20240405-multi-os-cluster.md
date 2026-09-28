# Windows Node Support for Multi-OS Cluster in minikube 

* First proposed: 2017-09-28 by Ali Kahoot(@kahootali)
* Authors: Vinicius Apolinario(@vrapolinario), Ian King'ori(@iankingori), Bob Sira(@bobsira)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Struggling with comprehensive testing for applications across Linux and Windows environments? We are proposing to enhance minikube's capabilities by introducing Windows node support. minikube start should have the option of setting up a multi-OS cluster with a Linux control plane and both Linux and Windows worker nodes. This is [one of the most requested features](https://github.com/kubernetes/minikube/issues/2015) among the issues opened in the minikube repository. Developers will experience a significant boost in testing capabilities, streamlined workflows, and faster deployment, and a more user-friendly, developer-focused environment, including reliable testing infrastructure. 

## Goals

*   Launching a cluster with both Windows and Linux nodes using the Hyper-V driver.
*   Provide the ability to add Windows nodes to an existing minikube cluster. 

## Non-Goals

*   Support for non-hyper-v drivers on Windows
*  Advanced CNI configurations
*  Simultaneous Multi-OS Node Addition

## Design Details

The initial design creates a mixed-OS cluster on a Windows host using Hyper-V, with one Linux control
plane and one Windows Server 2025 worker. A repeatable `--node` flag specifies each node's role and
operating system, defaulting to Linux:

```powershell
.\out\minikube.exe start --node role=control-plane --node role=worker,os=windows
```

Node roles and guest OS information extend the existing node configuration model. Unsupported topologies
and conflicting settings are rejected before provisioning. Existing Linux `--nodes` and `--ha` workflows
remain unchanged when `--node` is not used.

Windows workers boot from prepared VHDX images built with
[windows-node-image-builder](https://github.com/bobsira/windows-node-image-builder), using either a hosted
or custom image. minikube uses its existing SSH transport to configure and join the worker, with
containerd as the runtime and Flannel for networking.

Windows, Kubernetes, and workload image versions must be compatible. Testing requires unit coverage,
Linux regression coverage, and Windows Hyper-V end-to-end coverage for node readiness, networking,
Windows workloads, and cleanup.

Additional mixed-OS workers and adding Windows nodes to existing clusters remain future work.
Windows networking and restart reliability remain experimental.
