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

```shell
minikube status [flags]
```

### Options

```
  -f, --format string         Go template format string for the status output.  The format for Go templates can be found here: https://pkg.go.dev/text/template
                              For the list accessible variables for the template, see the struct values here: https://pkg.go.dev/k8s.io/minikube/cmd/minikube/cmd#Status (default "{{.Name}}\ntype: Control Plane\nhost: {{.Host}}\nkubelet: {{.Kubelet}}\napiserver: {{.APIServer}}\nkubeconfig: {{.Kubeconfig}}\n{{- if .TimeToStop }}\ntimeToStop: {{.TimeToStop}}\n{{- end }}\n{{- if .DockerEnv }}\ndocker-env: {{.DockerEnv}}\n{{- end }}\n{{- if .PodManEnv }}\npodman-env: {{.PodManEnv}}\n{{- end }}\n\n")
  -l, --layout string         output layout (EXPERIMENTAL, JSON only): 'nodes' or 'cluster' (default "nodes")
  -n, --node string           The node to check status for. Defaults to control plane. Leave blank with default format for status on all nodes.
  -o, --output string         minikube status --output OUTPUT. json, text (default "text")
  -w, --watch duration[=1s]   Continuously listing/getting the status with optional interval duration. (default 1s)
```

### Options inherited from parent commands

```
      --add_dir_header                   If true, adds the file directory to the header of the log messages
      --alsologtostderr                  log to standard error as well as files (no effect when -logtostderr=true)
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
  -h, --help                             
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory (no effect when -logtostderr=true)
      --log_file string                  If non-empty, use this log file (no effect when -logtostderr=true)
      --log_file_max_size uint           Defines the maximum size a log file can grow to (no effect when -logtostderr=true). Unit is megabytes. If the value is 0, the maximum file size is unlimited. (default 1800)
      --logtostderr                      log to standard error instead of files (default true)
      --one_output                       If true, only write logs to their native severity level (vs also writing to each lower severity level; no effect when -logtostderr=true)
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --rootless                         Force to use rootless driver (docker and podman driver only)
      --skip-audit                       Skip recording the current command in the audit logs.
      --skip_headers                     If true, avoid header prefixes in the log messages
      --skip_log_headers                 If true, avoid headers when opening log files (no effect when -logtostderr=true)
      --stderrthreshold severity         logs at or above this threshold go to stderr when writing to files and stderr (no effect when -logtostderr=true or -alsologtostderr=true) (default 2)
      --user string                      Specifies the user executing the operation. Useful for auditing operations executed by 3rd party tools. Defaults to the operating system username.
  -v, --v Level                          number for the log level verbosity
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

