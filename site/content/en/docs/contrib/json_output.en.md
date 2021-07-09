---
title: "Minikube JSON Output"
date: 2019-07-31
weight: 4
description: >
  How to add logs to facilitate JSON output
---

This document is written for minikube contributors who need to add logs to the minikube log registry for successful JSON output.
You may need to add logs to the registry if the `TestJSONOutput` integration test is failing on your PR.

### Background

minikube provides JSON output for `minikube start`, accesible via the `--output` flag:

```shell
minikube start --output json
```

This converts regular output:

```
$ minikube start

😄  minikube v1.12.1 on Darwin 10.14.6
✨  Automatically selected the hyperkit driver
👍  Starting control plane node minikube in cluster minikube
🔥  Creating hyperkit VM (CPUs=2, Memory=6000MB, Disk=20000MB) ...
```

into a [Cloud Events](https://cloudevents.io/) compatible JSON output:

```
$ minikube start --output json

{"data":{"currentstep":"0","message":"minikube v1.12.1 on Darwin 10.14.6\n","name":"Initial Minikube Setup","totalsteps":"10"},"datacontenttype":"application/json","id":"68ff70ae-202b-4b13-8351-e9f060e8c56e","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}
{"data":{"currentstep":"1","message":"Automatically selected the hyperkit driver\n","name":"Selecting Driver","totalsteps":"10"},"datacontenttype":"application/json","id":"39bed8e9-3c1a-444e-997c-2ec19bdb1ca1","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}
{"data":{"currentstep":"3","message":"Starting control plane node minikube in cluster minikube\n","name":"Starting Node","totalsteps":"10"},"datacontenttype":"application/json","id":"7c80bc53-3ac4-4a42-a493-92e269cc56c9","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}
{"data":{"currentstep":"6","message":"Creating hyperkit VM (CPUs=2, Memory=6000MB, Disk=20000MB) ...\n","name":"Creating VM","totalsteps":"10"},"datacontenttype":"application/json","id":"7f5f23a4-9a09-4954-8abc-d29bda2cc569","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.step"}
```

There are a few key points to note in the above output:

1. Each log of type `io.k8s.sigs.minikube.step` indicates a distinct step in the `minikube start` process
1. Each step has a `currentstep` field which allows clients to track `minikube start` progress
1. Each `currentstep` is distinct and increasing in order

To achieve this output, minikube maintains a registry of logs.
This way, minikube knows how many expected `totalsteps` there are at the beginning of the process, and what the current step is.

If you change logs, or add a new log, you need to update the minikube registry to pass integration tests.

### Adding a Log to the Registry

There are three steps to adding a log to the registry, which exists in [register.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/out/register/register.go).

You will need to add your new log in two places:

1. As a constant of type `RegStep` [here](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/out/register/register.go#L24)
1. In the register itself in the `init()` function, [here](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/out/register/register.go#L52)

**Note: It's important that the order of steps matches the expected order they will be called in. So, if you add a step that is supposed to be called after "Preparing Kubernetes", the new step should be place after "Preparing Kubernetes".

Finally, set your new step in the cod by placing this line before you call `out.T`:

```go
register.Reg.SetStep(register.MyNewStep)
```

You can see an example of setting the registry step in the code in [config.go](https://github.com/kubernetes/minikube/blob/master/pkg/minikube/node/config.go):

```go
 register.Reg.SetStep(register.PreparingKubernetes)
 out.Step(cr.Style(), "Preparing Kubernetes {{.k8sVersion}} on {{.runtime}} {{.runtimeVersion}} ...", out.V{"k8sVersion": k8sVersion, "runtime": cr.Name(), "runtimeVersion": version})
```
