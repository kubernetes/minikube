---
title: "CPU Usage Benchmarks (Linux)"
linkTitle: "CPU Usage Benchmarks (Linux)"
weight: 1
---

## CPU% Busy Overhead - Average first 5 minutes only

This chart shows each tool's CPU busy overhead percentage.
After each tool's starting, we measured each tool's idle for 5 minutes.
This chart was measured only after the start without deploying any pods.

  1. start each local kubernetes tool
  2. measure its cpu usage with [cstat](https://github.com/tstromberg/cstat)

![idleOnly](/images/benchmarks/cpuUsage/idleOnly/linux.png)

NOTE: the benchmark environment uses GCE with nested virtualization. This may affect virtual machine's overhead.
https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances

## CPU% Busy Overhead - With Auto Pause vs. Non Auto Pause

This chart shows each tool's CPU busy overhead percentage with auto-pause addon.
The auto-pause is mechanism which reduce CPU busy usage by pausing kube-apiserver.
We compare CPU usage after deploying sample application (nginx deployment) to all tools (including minikube and other tools).
This chart was measured with the following steps.
By these steps, we compare CPU usage with auto-pause vs. non-auto-pause.

 1. start each local kubernetes tool
 2. deploy sample application (nginx deployment) to each tool
 3. wait 1 minute without anything
 4. measure No.3 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)
 5. if tool is minikube, enable auto-pause addon which pause control plane
 6. if tool is minikube, wait 1 minute so that control plane will become Paused status (It takes 1 minute to become Pause status from Stopped status)
 7. if tool is minikube, verify if minikube control plane is paused
 8. if tool is minikube, wait 3 minute without anything
 9. if tool is minikube, measure No.8 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)

No.1-4: Initial start CPU usage with sample (nginx) deployment
No.5-9: Auto Paused CPU usage with sample (nginx) deployment

![autopause](/images/benchmarks/cpuUsage/autoPause/linux.png)

NOTE: the benchmark environment uses GCE with nested virtualization. This may affect virtual machine's overhead.
https://cloud.google.com/compute/docs/instances/enable-nested-virtualization-vm-instances
