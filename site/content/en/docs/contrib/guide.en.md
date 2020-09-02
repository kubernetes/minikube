---
title: "Contributor Guide"
linkTitle: "Guide"
date: 2019-07-31
weight: 1
description: >
  How to become a minikube contributor
---

### Code of Conduct

 Be excellent to each another. Please refer to our [Kubernetes Community Code of Conduct](https://git.k8s.io/community/code-of-conduct.md).

### License Agreement

We'd love to accept your patches! Before we can take them, [please fill out either the individual or corporate Contributor License Agreement (CLA)](https://git.k8s.io/community/CLA.md)

### Finding issues to work on

* ["good first issue"](https://github.com/kubernetes/minikube/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22)  -  issues where there is a clear path to resolution
* ["help wanted"](https://github.com/kubernetes/minikube/issues?utf8=%E2%9C%93&q=is%3Aissue+is%3Aopen+label%3A%22help+wanted%22+) - issues where we've identified a need but not resources to work on them
"priority/important-soon" or "priority/important-longterm: - high impact issues that need to be addressed in the next couple of releases.

* Ask on the #minikube Slack if you aren't sure

Once you've discovered an issue to work on:

* Add a comment mentioning that you plan to work on the issue
* Send a PR out that mentions the issue
* Comment on the issue with `/assign` to assign it to yourself

### Contributing A Patch

1. Submit an issue describing your proposed change
2. A reviewer will respond to your issue promptly.
3. If your proposed change is accepted, and you haven't already done so, sign the [Contributor License Agreement (CLA)](https://git.k8s.io/community/CLA.md)
4. Fork the minikube repository, develop and test your code changes.
    * Before test, you may need to install some [prerequisites](https://minikube.sigs.k8s.io/docs/contrib/testing/#prerequisites).
5. Submit a pull request.

## Contributing larger changes

To get feedback on a larger, more ambitious changes, create a PR containing your idea using the [MEP (minikube enhancement proposal) template](https://github.com/kubernetes/minikube/tree/master/enhancements). This way other contributors can comment on design issues early on, though you are welcome to work on the code in parallel.

If you send out a large change without a MEP, prepare to be asked by other contributors for one to be included within the PR.

### Style Guides

For coding, refer to the [Kubernetes Coding Conventions](https://github.com/kubernetes/community/blob/master/contributors/guide/coding-conventions.md#code-conventions)

For documentation, refer to the [Kubernetes Documentation Style Guide](https://kubernetes.io/docs/contribute/style/style-guide/)
