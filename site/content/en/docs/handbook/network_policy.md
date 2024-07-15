---
title: "Network Policy"
linkTitle: "Network Policy"
weight: 10
date: 2022-01-31
description: >
  Controlling traffic flowing through the cluster
aliases:
  - /docs/reference/network_policy
---

minikube allows users to create and test network policies in the local Kubernetes cluster. This is useful since it allows the network policies to be, considered, built, and evaluated during the application development, as an integral part of the process rather than "bolted on" at the end of development.

[Kubernetes NetworkPolicies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) allow the control of pod network traffic passing through the cluster, at the IP address or port level (OSI layer 3 or 4). The linked page provides much more information about the functionality and implementation.

However, the [prerequisites](https://kubernetes.io/docs/concepts/services-networking/network-policies/#prerequisites) note that Network policies are implemented by the Container Network Interface (CNI) network plugin. Therefore to use or test network policies in any Kubernetes cluster, you must be using a networking solution which supports NetworkPolicy. Creating a NetworkPolicy resource without a controller that implements it will have no effect.

A vanilla minikube installation (`minikube start`) does not support any NetworkPolicies, since the default CNI, [Kindnet](https://github.com/aojea/kindnet), does not support Network Policies, [by design](https://github.com/kubernetes-sigs/kind/issues/842#issuecomment-528824670).

However, minikube can support [NetworkPolicies](https://kubernetes.io/docs/concepts/services-networking/network-policies/) if a supported CNI, such as [Calico](https://projectcalico.docs.tigera.io/about/about-calico), is installed. In addition, in this scenario both [Kubernetes network policy](https://projectcalico.docs.tigera.io/security/kubernetes-network-policy) and [Calico network policy](https://projectcalico.docs.tigera.io/security/calico-network-policy) are supported.

**Calico network policy** provides a richer set of policy capabilities than **Kubernetes network policy** including:
* policy ordering/priority
* deny rules
* more flexible match rules

## Enabling Calico on a minikube cluster

It is possible to replace the CNI on a running minikube cluster, but it is significantly easier to simply append the `--cni calico` flag to the `minikube start` command when following the instructions on the [Get Started!]({{<ref "/docs/start/" >}}) page to build the minikube cluster with Calico installed from the outset.

## Kubernetes Network Policy example

The [Kubernetes documentation on declaring network policy](https://kubernetes.io/docs/tasks/administer-cluster/declare-network-policy/) is a good place to start to understand the possibilities. In addition, the tutorials in [Further reading]({{< ref "#further-reading" >}}) below give much more guidance. 

The YAML below from the [Kubernetes NetworkPolicies](https://kubernetes.io/docs/concepts/services-networking/network-policies/#default-deny-all-ingress-traffic) documentation shows a very simple default ingress isolation policy on a namespace by creating a NetworkPolicy that selects all pods but does not allow any ingress traffic to those pods.

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
spec:
  podSelector: {}
  policyTypes:
  - Ingress
```

## Calico Network Policy example

The [Calico network policy documentation](https://projectcalico.docs.tigera.io/security/calico-policy) is the best place to learn about the extended feature set of Calico network policy and how it coexists with Kubernetes network policy.

The YAML below from the [Calico policy tutorial](https://projectcalico.docs.tigera.io/security/tutorials/calico-policy) shows a very simple default deny **Global** Calico Network Policy (not available with vanilla Kubernetes network policy) that is often used as a starting point for an effective zero-trust network model. Note that **Global** Calico Network Policies are not namespaced and affect all pods that match the policy selector. In contrast, Kubernetes Network Policies are namespaced, so you would need to create a default deny policy per namespace to achieve the same effect. In this example pods in the kube-system namespace are excluded to keep Kubernetes itself running smoothly.

```yaml
---
apiVersion: projectcalico.org/v3
kind: GlobalNetworkPolicy
metadata:
  name: default-deny
spec:
  selector: projectcalico.org/namespace != "kube-system"
  types:
  - Ingress
  - Egress
```

## Further reading

This [Advanced Kubernetes policy tutorial](https://docs.tigera.io/calico/latest/network-policy/get-started/kubernetes-policy/kubernetes-policy-advanced) gives an example of what can be achieved with Kubernetes network policy. It walks through using Kubernetes NetworkPolicy to define more complex network policies.

This [Calico policy tutorial](https://projectcalico.docs.tigera.io/security/tutorials/calico-policy) demonstrates the extended functionalities Calico network policy offers over and above vanilla Kubernetes network policies. To demonstrate this, this tutorial follows a similar approach to the tutorial above, but instead uses Calico network policies and highlights differences between the two policy types, making use of features that are not available in Kubernetes network policies. 
