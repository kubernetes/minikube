---
title: "Troubleshooting"
weight: 20
description: >
  How to troubleshoot minikube issues
---

## Enabling debug logs

Pass `--alsologtostderr` to minikube commands to see detailed log output. To increase the log verbosity, you can use:

* `-v=1`: verbose messages
* `-v=2`: really verbose messages
* `-v=8`: more log messages than you can possibly handle.

Example:

`minikube start --alsologtostderr --v=2` will start minikube and output all the important debug logs to stderr.

## Post-mortem minikube debug logs

minikube stores post-mortem logs in the temporary directory of your system. One log file is created per subcommand and any subsequent invocations of the subcommand with the same args will append to the same file. If the log file has exceeded 1MB in size, a new log file is created. On macOS or Linux, it's easy to get a list of recent logs:

```shell
find $TMPDIR -mtime -1 -type f -name "*minikube*" -ls  2>/dev/null
```

For instance after running `minikube start`, the above comamnd will show:

`-rw-r--r-- 1 user  grp  718 Aug 18 12:40 /var/folders/n1/qxvd9kc/T//minikube_start_dc950831e1a232e0318a6d6ca82aaf4f4a8a048b_0.log`

These are plain text log files: you may rename them to "<filename>.log" and then drag/drop them into a GitHub issue for further analysis by the minikube team. You can quickly inspect the final lines of any of these logs via:
  
```shell
tail -n 10 <filename>
```

for example, this shows:

```
I0818 12:40:17.027317   63501 out.go:197] Setting ErrFile to fd 2...
I0818 12:40:17.027321   63501 out.go:231] isatty.IsTerminal(2) = true
I0818 12:40:17.027423   63501 root.go:272] Updating PATH: /Users/tstromberg/.minikube/bin
I0818 12:40:17.027715   63501 mustload.go:64] Loading cluster: minikube
```

## Gathering VM logs

To debug issues where Kubernetes failed to deploy, it is very useful to collect the Kubernetes pod and kernel logs:

```shell
minikube logs
```

## Viewing Pod Status

To view the deployment state of all Kubernetes pods, use:

```shell
kubectl get po -A
```

Example output:

```shell
NAMESPACE     NAME                        READY   STATUS    RESTARTS   AGE
kube-system   coredns-5c98db65d4-299md    1/1     Running   0          11m
kube-system   coredns-5c98db65d4-qlpkd    1/1     Running   0          11m
kube-system   etcd-minikube               1/1     Running   0          10m
kube-system   gvisor                      1/1     Running   0          11m
...
kube-system   storage-provisioner         1/1     Running   0          11m
```

To view more detailed information about a pod, use:

```shell
kubectl describe pod <name> -n <namespace>
```

## Debugging hung start-up

minikube will wait ~8 minutes before giving up on a Kubernetes deployment. If you want to see startup fails more immediately, consider using:

```shell
minikube logs --problems
```

This will attempt to surface known errors, such as invalid configuration flags. If nothing interesting shows up, try `minikube logs`.
