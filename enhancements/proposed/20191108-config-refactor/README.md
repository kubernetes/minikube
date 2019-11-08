# Config types refactor

* First proposed: 2019-11-08
* Authors: Thomas Stromberg (@tstromberg)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

A refactor of `config.Config` to address multi-node and VM-free plans and improve readability. At the moment, `config.Config` has two items:

* `MachineConfig`
* `KubernetesConfig`

The two primary shortcomings is that this layout does not allow for multiple nodes, and `MachineConfig` does not contain a Kubelet version to deploy.

## Goals

* Prepare for multi-node
* Prepare for VM-free, which needs to know what Kubernetes version to deploy
* Idiomatic, in particular, addressing naming stutter
* Reflects terminology used by [Kubernetes Standardized Glossary](https://kubernetes.io/docs/reference/glossary/?fundamental=true)
* Shallow structure: no more than 3 levels deep

## Non-Goals

* Handling use-cases that are not yet included on a roadmap
* Perfection

## Terminology

### Kubernetes

* *Cluster*: A set of machines, called nodes, that run containerized applications managed by Kubernetes. A cluster has at least one worker node and at least one master node
* *Node*: A node is a worker machine in Kubernetes, running a Kubelet
* *Machine*: The environment in which a Kubelet will be deployed. Often a guest virtual machine, but may be a physical machine or container.

### minikube

* *Host*: The environment where the driver is running. Typically the same machine that minikube is running on. Not to be confused with libmachine, which uses `host.Host` to refer to a Virtual Machine.

## Design Details

`config.Config` is renamed to `config.Options`, which will contain:

* `Name`: the name of this configuration set (profile)
* `Host`: options which only affect the host running minikube
* `Cluster`: options which apply to the entire Kubernetes cluster
* `Nodes`: options which apply individually to each node

## Host

| *Old* | *New* |
| Config.KubernetesConfig.ShouldLoadCachedImages | LoadCachedImages |
| Config.MachineConfig.DisableDriverMounts | EnableMounts | 
| Config.MachineConfig.DNSProxy | EnableDNSProxy |
| Config.MachineConfig.EmbedCerts | EmbedCerts |
| Config.MachineConfig.HostDNSResolver | EnableClusterDNS |
| Config.MachineConfig.HyperkitVpnKitSock | HyperKitVPNKitSock |
| Config.MachineConfig.HyperkitVSockPorts | HyperKitVSockPorts |
| Config.MachineConfig.HypervVirtualSwitch | Switch |
| Config.MachineConfig.KeepContext | KeepContext | 
| Config.MachineConfig.KVMNetwork | Switch |
| Config.MachineConfig.KVMGPU | EnableGPU | 
| Config.MachineConfig.KVMHidden | HideHypervisor | 
| Config.MachineConfig.KVMQemuURI | ConnectionURI |
| Config.MachineConfig.NFSShares | Mounts | 
| Config.MachineConfig.NoVTXCheck | VTXCheck |
| Config.MachineConfig.VMDriver | Driver |

## Cluster

| *Old* | *New* |
| Config.KubernetesConfig.APIServerIPs | APIServer.IPs |
| Config.KubernetesConfig.APIServerName | APIServer.Names[0] |
| Config.KubernetesConfig.APIServerNames | APIServer.Names |
| Config.KubernetesConfig.DNSDomain | DNSDomain |
| Config.KubernetesConfig.FeatureGates | FeatureGates |
| Config.KubernetesConfig.ImageRepository | ImageRepository |
| Config.KubernetesConfig.NetworkPlugin | NetworkPlugin |
| Config.KubernetesConfig.NodePort | APIServer.Port |
| Config.KubernetesConfig.ServiceCIDR | ServiceCIDR |
| Config.MachineConfig.HostOnlyCIDR | CIDR |

## Node (multiple)

| *Old* | *New* |
| N/A | Role | 
| Config.KubernetesConfig.ContainerRuntime | ContainerRuntime |
| Config.KubernetesConfig.CRISocket | ContainerRuntimeSocket |
| Config.KubernetesConfig.EnableDefaultCNI | EnableDefaultCNI |
| Config.KubernetesConfig.ExtraOptions | KubeletOptions | 
| Config.KubernetesConfig.KubernetesVersion | KubeletVersion |
| Config.KubernetesConfig.NodeIP | IP |
| Config.KubernetesConfig.NodeName | Name |
| Config.MachineConfig.ContainerRuntime | ContainerRuntime 
| Config.MachineConfig.CPUs | CPUCount | 
| Config.MachineConfig.DiskSize | DiskSize |
| Config.MachineConfig.DockerEnv | DockerEnv |
| Config.MachineConfig.DockerOpt | DockerOpt |
| Config.MachineConfig.InsecureRegistry | InsecureRegistry |
| Config.MachineConfig.Memory | Memory |
| Config.MachineConfig.MinikubeISO | BootImage |
| Config.MachineConfig.NFSShares | MountRoot | 
| Config.MachineConfig.RegistryMirror | RegistryMirror |
| Config.MachineConfig.UUID | UUID |

## Alternatives Considered

A more deeply nested structure. While this is ideal from an ideal of separation of duties (Law of Demeter), it doesn't appear to assist in readability.