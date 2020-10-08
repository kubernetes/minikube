---
title: "status"
description: >
  Gets the status of a local Kubernetes cluster
---


## minikube status

Gets the status of a local Kubernetes cluster

### Synopsis

Gets the status of a local Kubernetes cluster.
	Exit status contains the status of minikube's VM, cluster and Kubernetes encoded on it's bits in this order from right to left.
	Eg: 7 meaning: 1 (for minikube NOK) + 2 (for cluster NOK) + 4 (for Kubernetes NOK)

```
minikube status [flags]
```

### Options

```
  -f, --format string   Go template format string for the status output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
                        For the list accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#Status (default "{{.Name}}\ntype: Control Plane\nhost: {{.Host}}\nkubelet: {{.Kubelet}}\napiserver: {{.APIServer}}\nkubeconfig: {{.Kubeconfig}}\n\n")
  -h, --help            help for status
  -l, --layout string   output layout (EXPERIMENTAL, JSON only): 'nodes' or 'cluster' (default "nodes")
  -n, --node string     The node to check status for. Defaults to control plane. Leave blank with default format for status on all nodes.
  -o, --output string   minikube status --output OUTPUT. json, text (default "text")
```

### Options inherited from parent commands

```
      --add_dir_header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --log_file string                  If non-empty, use this log file
      --log_file_max_size uint           Defines the maximum size a log file can grow to. Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

