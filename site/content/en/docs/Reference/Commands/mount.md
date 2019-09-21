---
title: "mount"
linkTitle: "mount"
weight: 1
date: 2019-08-01
description: >
  Mounts the specified directory into minikube
---

### Usage

```
minikube mount [flags] <source directory>:<target directory>
```

### Options

```
      --9p-version string   Specify the 9p version that the mount should use (default "9p2000.L")
      --gid string          Default group id used for the mount (default "docker")
  -h, --help                help for mount
      --ip string           Specify the ip that the mount should be setup on
      --kill                Kill the mount process spawned by minikube start
      --mode uint           File permissions used for the mount (default 493)
      --msize int           The number of bytes to use for 9p packet payload (default 262144)
      --options strings     Additional mount options, such as cache=fscache
      --type string         Specify the mount filesystem type (supported types: 9p) (default "9p")
      --uid string          Default user id used for the mount (default "docker")
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

## Note for People on Windows -
If you are on Windows, you might see an error where the mounting command errors like below -
```
I0809 15:07:14.612443 13984 mount.go:64] mount err=Process exited with status 32, out=mount: /mount-dir: mount(2) system call failed: Connection timed out.
W0809 15:07:14.616451 13984 exit.go:99] mount failed: mount: /mount-dir: mount(2) system call failed: Connection timed out.
: Process exited with status 32
*
X mount failed: mount: /mount-dir: mount(2) system call failed: Connection timed out.
: Process exited with status 32
```

What this technically means is that 9P client inside minikube was not able to connect to the Host. This generally happens when your Firewall is blocking the incoming connection to the host for file sharing.
This has been observed generally on people with Windows Defender Firewall. Affects both Hyper-V & VirtualBox.

In order to resolve this issue, please follow the below steps (These need to be followed only once and require Administrator privileges.)-
1. Open up an Administrator PowerShell Prompt.
2. Type in the following query -
```
Get-NetFirewallRule | Where-Object { $_.Name -match "minikube" } | Remove-NetFirewallRule
```
3. Once the command has completed, run your mounting command again.
4. This time, Windows Defender Firewall will ask your for permissions.
5. Allow the minikube executable access to the Private & Public networks and click on "Allow".

If you install a new version, you *might* have to redo these steps.

If you are using some 3rd party Firewall, please follow the below steps -
1. Fire up your cluster.
2. Once your cluster has started, run `minikube ip`. Note down the IP address which is outputted.
3. For example, if the IP address you get is `192.168.215.103`, then you need to allow INBOUND communication to your system from the range - `192.168.215.0-192.168.215.255`