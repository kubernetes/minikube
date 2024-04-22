# Windows Node Support for Multi-OS Cluster in Minikube 

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

Struggling with comprehensive testing for applications across Linux and Windows environments? We are proposing to enhance Minikube's capabilities by introducing Windows node support. Minikube start should have the option of setting up a multi-OS cluster with both Linux and Windows node. This is [one of the most requested features](https://github.com/kubernetes/minikube/issues/2015) among the issues opened in the minikube repository. Developers will experience a significant boost in testing capabilities, streamlined workflows, and faster deployment, contributing to a more user-friendly and developer-focused environment 

## Goals

*   Launching a cluster with both Windows and Linux nodes using the hyper-v driver. 

## Non-Goals

*   _A bulleted list of what is out of scope for this proposal_
*   _Listing non-goals helps to focus the discussion_

## Design Details

We have a working POC where we have automated joining a Windows node to a Minikube cluster using PowerShell scripts by just running one command.  

Our step-by-step process involves: 
* Setting up the Windows node – Create a Windows Virtual Machine and install Windows Server 2022 unattended via [Autounattend XML file](https://learn.microsoft.com/en-us/windows-hardware/manufacture/desktop/automate-windows-setup?view=windows-11). 
* Create and configure a Minikube cluster with Flannel as CNI - Remote into the Virtual Machine (Windows node) and install containerd, NSSM, kubelet and kubeadm respectively. 
* Join the Windows node into the Minikube cluster using kubeadm join command. 
* Configuring Flannel CNI and Kube-Proxy - configure the networking settings as any Kubernetes cluster. 

It is also worth noting that SIG Windows group exploring to bring Windows nodes on non-Windows host using KVM2 to create the nodes and as the minikube driver.  

 Users will be able to start multi-os clusters by specifying a windows flag as in the example below, 

minikube start –windows-22-node=1 or minikube start -windows-19-node=1 {windows-osversion-numberofwindowsnode} 

## Alternatives Considered

_Alternative ideas that you are leaning against._