---
title: "CPU Usage Benchmarks(Linux)"
linkTitle: "CPU Usage Benchmarks(Linux)"
weight: 1
---

## CPU% Busy Overhead - Avarage first 5 minutes only

This chart shows each tool's CPU busy overhead percentage.   
After each tool's starting, we measured each tool's idle for 5 minutes.  
This chart was measured only after the start without deploying any pods.

![idleOnly](/images/benchmarks/cpuUsage/linux.png)

## CPU% Busy Overhead - With Auto Pause vs. Non Auto Pause

This chart shows each tool's CPU busy overhead percentage with auto-pause addon.   
The auto-pause is mechanism which reduce CPU busy usage by pausing kube-apiserver.  
This chart was measured with the following steps.
By these steps, we compare CPU usage with auto-pause vs. non-auto-pause.  

 1. start each local kubernetes tool
 2. deploy sample application(nginx deployment)
 3. wait 1 minute without anything
 4. measure No.3 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)
 5. enable auto-pause addons(only if tool is minikube)
 6. wait 3 minute without anything
 7. measure No.6 idle CPU usage with [cstat](https://github.com/tstromberg/cstat)
 
![autopause](/images/benchmarks/cpuUsage/autoPause/linux.png)