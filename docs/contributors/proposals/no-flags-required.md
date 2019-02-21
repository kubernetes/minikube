---
title: KEP Template
authors:
  - "@janedoe"
owning-sig: sig-xxx
participating-sigs:
  - sig-aaa
  - sig-bbb
reviewers:
  - TBD
  - "@alicedoe"
approvers:
  - TBD
  - "@oscardoe"
editor: TBD
creation-date: yyyy-mm-dd
last-updated: yyyy-mm-dd
status: provisional|implementable|implemented|deferred|rejected|withdrawn|replaced
see-also:
  - "/keps/sig-aaa/20190101-we-heard-you-like-keps.md"
  - "/keps/sig-bbb/20190102-everyone-gets-a-kep.md"
replaces:
  - "/keps/sig-ccc/20181231-replaced-kep.md"
superseded-by:
  - "/keps/sig-xxx/20190104-superceding-kep.md"
---

# minikube: no flags required (automated VM configuration)


## Table of Contents

A table of contents is helpful for quickly jumping to sections of a KEP and for highlighting any additional information provided beyond the standard KEP template.
[Tools for generating][] a table of contents from markdown are available.

- [Title](#title)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
    - [User Stories [optional]](#user-stories-optional)
      - [Story 1](#story-1)
      - [Story 2](#story-2)
    - [Implementation Details/Notes/Constraints [optional]](#implementation-detailsnotesconstraints-optional)
    - [Risks and Mitigations](#risks-and-mitigations)
  - [Design Details](#design-details)
    - [Test Plan](#test-plan)
    - [Graduation Criteria](#graduation-criteria)
      - [Examples](#examples)
        - [Alpha -> Beta Graduation](#alpha---beta-graduation)
        - [Beta -> GA Graduation](#beta---ga-graduation)
        - [Removing a deprecated flag](#removing-a-deprecated-flag)
    - [Upgrade / Downgrade Strategy](#upgrade--downgrade-strategy)
    - [Version Skew Strategy](#version-skew-strategy)
  - [Implementation History](#implementation-history)
  - [Drawbacks [optional]](#drawbacks-optional)
  - [Alternatives [optional]](#alternatives-optional)
  - [Infrastructure Needed [optional]](#infrastructure-needed-optional)

[Tools for generating]: https://github.com/ekalinin/github-markdown-toc

## Summary

`minikube start` should work for 95% of users out of the box, without needing to specifically prepare a VM driver or download binaries.

## Motivation

New users to minikube constantly stumble on VM driver setup. It is by far the most complicated part of our documentation, and time consuming part of installation.

### Goals

If a hypervisor is correctly installed, `minikube start` should work without any flags required for the following combinations:

- Linux: KVM, VirtualBox
- Windows: HyperV, VirtualBox
- macOS: Hyperkit, VirtualBox, VMWare Fusion, Parallels

If there is not a fully installed hypervisor, `minikube start` should offer the user the choice to automatically download or fix issues relating to the following hypervisors

- Linux: KVM
- Windows: HyperV, VirtualBox
- macOS: Hyperkit

### Non-Goals

- A separate installer
- Installing hypervisors that are not open-source
- Support for the `none` driver

## Proposal

This is where we get down to the nitty gritty of what the proposal actually is.

### User Stories [optional]

#### Story 1: macOS

- User on a brand new macOS install downloads minikube
- User runs `minikube start`
- minikube outputs:

```
We were not able to locate a hypervisor compatible with minikube. Would it be OK if minikube sets up the hyperkit driver? [y/n]
```

If `y`, minikube would then fetch and install hyperkit as well as the hyperkit VM driver. To avoid version mismatches, it should do so in a location that does not conflict with Docker binaries.

If `n`, minikube would output:

```
OK. If you would like to install hyperkit in the future, run:

minikube fix --step=hyperkit
```

#### Story 2: Ubuntu


- User on a brand new Ubuntu 18.10 install downloads minikube
- User runs `minikube start`
- minikube outputs:

```
We were not able to locate a hypervisor compatible with minikube. Would it be OK if minikube sets up the Linux KVM driver? [y/n]
```

If `y`, minikube would install libvirt via apt, create the libvirt group, and install the kvm2 driver.

If `n`, minikube would output:

```
OK. If you would like to install kvm in the future, run:

minikube fix --step=kvm
```

#### Story 3: Windows

- User on a brand new Windows 10 install downloads minikube
- User runs `minikube start`
- minikube just works with HyperV out of the box.

### Implementation Details/Notes/Constraints [optional]

The suggested hypervisors will be chosen based on systems compatibility, and ranked against recent stability trends in our continuous integration system.


## Design Details

### Test Plan

**Note:** *Section not required until targeted at a release.*

Consider the following in developing a test plan for this enhancement:
- Will there be e2e and integration tests, in addition to unit tests?
- How will it be tested in isolation vs with other components?

No need to outline all of the test cases, just the general strategy.
Anything that would count as tricky in the implementation and anything particularly challenging to test should be called out.

All code is expected to have adequate tests (eventually with coverage expectations).
Please adhere to the [Kubernetes testing guidelines][testing-guidelines] when drafting this test plan.

[testing-guidelines]: https://git.k8s.io/community/contributors/devel/sig-testing/testing.md

### Version Skew Strategy

To avoid version skew issues with VM drivers, VM drivers will be stored in a version-specific directory of minikube. 

Hypervisors on the other hand will be shared across minikube versions.

## Drawbacks [optional]

Why should this KEP _not_ be implemented.

## Alternatives [optional]

A separate installer

Similar to the `Drawbacks` section the `Alternatives` section is used to highlight and record other possible approaches to delivering the value proposed by a KEP.

