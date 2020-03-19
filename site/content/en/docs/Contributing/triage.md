---
linkTitle: "Triage"
title: "Triaging Minikube Issues"
date: 2020-03-17
weight: 10
description: >
  How to triage issues in the minikube repo
---

Triage is an important part of maintaining the health of the minikube repo.
A well organized repo allows maintainers to prioritize feature requests, fix bugs, and respond to users facing difficulty with the tool as quickly as possible.

Triage includes:
- Labeling issues
- Responding to issues
- Closing issues (under certain circumstances!)

If you're interested in helping out with minikube triage, this doc covers the basics of doing so.

Additionally, if you'd be interested in participating in our weekly triage meeting, please fill out this [form](https://forms.gle/vNtWZSWXqeYaaNbU9) to express interest. Thank you! 

# Daily Triage
Daily triage has two goals:

1. Responsiveness for new issues
1. Responsiveness when explicitly requested information was provided

The list of outstanding items are at https://teaparty-tts3vkcpgq-uc.a.run.app/s/daily-triage - it covers:

1. Issues without a `kind/` or `triage/` label
1. Issues without a `priority/` label
1. `triage/needs-information` issues which the user has followed up on

## Categorization

The most important level of categorizing the issue is defining what type it is.
We typically want at least one of the following labels on every issue, and some issues may fall into multiple categories:

- `triage/support`   - The default for most incoming issues
- `kind/bug` - When it’s a bug or we aren’t delivering the best user experience

Other possibilities: 
- `kind/feature`- Identify new feature requests
- `kind/flake` - Used for flaky integration or unit tests
- `kind/cleanup` - Cleaning up/refactoring the codebase
- `kind/documentation` - Updates or additions to minikube documentation
- `kind/ux` - Issues that involve improving user experience
- `kind/security` - When there's a security vulnerability in minikube

If the issue is specific to an operating system, hypervisor, container, addon, or Kubernetes component:

os/<operating system> - When the issue appears specific to an operating system
  - `os/linux`
  - `os/macos`
  - `os/windows`
co/<driver> - When the issue appears specific to a driver
  - `co/hyperkit`
  - `co/hyperv`
  - `co/kvm2`
  - `co/none-driver`
  - `co/docker-driver`
  - `co/virtualbox`
co/<kubernetes component> - When the issue appears specific to a k8s component
  - `co/apiserver`
  - `co/etcd`
  - `co/coredns`
  - `co/dashboard`
  - `co/kube-proxy`
  - `co/kubeadm`
  - `co/kubelet`
  - `co/kubeconfig`
 

Other useful tags:

Did an **Event** occur that we can dedup similar issues against?
- `ev/CrashLoopBackoff`
- `ev/Panic`
- `ev/Pending`
- `ev/kubeadm-exit-1`
Suspected **Root cause**:
- `cause/vm-environment`
- `cause/invalid-kubelet-options`

**Help wanted?**
`Good First Issue` - bug has a proposed solution, can be implemented w/o further discussion.
`Help wanted` - if the bug could use help from a contributor



# Responding to Issues

Many issues in the minikube repo fall into one of the following categories:
- Needs more information from the author to be actionable
- Duplicate Issue


## Closing with Care

Issues typically need to be closed for the following reasons:

- The issue has been addressed
- The issue is a duplicate of an existing issue
- There has been a lack of information over a long period of time

In any of these situations, we aim to be kind when closing the issue, and offer the author action items should they need to reopen their issue or still require a solution.

Samples responses for these situations include:

### Issue has been addressed

@author: I believe this issue is now addressed by minikube v1.4, as it <reason>. If you still see this issue with minikube v1.4 or higher, please reopen this issue by commenting with `/reopen`

Thank you for reporting this issue!

### Duplicate Issue


This issue appears to be a duplicate of #X, do you mind if we move the conversation there?

This way we can centralize the content relating to the issue. If you feel that this issue is not in fact a duplicate, please re-open it using `/reopen`. If you have additional information to share, please add it to the new issue.

Thank you for reporting this!

### Lack of Information

Hey @author -- hopefully it's OK if I close this - there wasn't enough information to make it actionable, and some time has already passed. If you are able to provide additional details, you may reopen it at any point by adding /reopen to your comment.

Here is additional information that may be helpful to us:

* Whether the issue occurs with the latest minikube release
*  The exact `minikube start` command line used
*  The full output of the `minikube start` command, preferably with `--alsologtostderr -v=3` for extra logging.
 * The full output of `minikube logs`

Thank you for sharing your experience!
