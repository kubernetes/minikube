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

We have a working POC where we have automated joining a Windows node to a minikube cluster using PowerShell scripts by just running one command.  

Our step-by-step process involves: 
* Setting up the Windows node – Create a Windows Virtual Machine and install Windows Server 2022 unattended via [Autounattend XML file](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/automate-windows-setup?view=windows-11). 
* Create and configure a minikube cluster with Flannel as CNI - Remote into the Virtual Machine (Windows node) and install containerd, NSSM, kubelet and kubeadm respectively. 
* Join the Windows node into the minikube cluster using kubeadm join command. 
* Configuring Flannel CNI and Kube-Proxy - configure the networking settings as any Kubernetes cluster. 

It is also worth noting that SIG Windows group exploring to bring Windows nodes on non-Windows host using KVM2 to create the nodes and as the minikube driver.  

 Users will be able to start multi-os clusters by specifying a windows flag as in the example below, 

minikube start –windows-node-version=2022 or minikube start -windows-node-version=2019 {windows-osversion}

Users will be able to add multiple versions of Windows node through separate `node add` commands, e.g.

minikube node add --os=windows –windows-node-version=2022 --nodes=2

## Alternatives Considered


* _[minikube on Windows and KVM driver](https://docs.google.com/document/d/1kCmZxvwAUmfc7SVqqigC_V8L3IaOki5TXX3fX7kEUgI/edit#heading=h.9r6udsnl0rp8) There are efforts by sig-windows to implement the same idea but on Linux environment using KVM driver_