---
title: "logs"
description: >
  Returns logs to debug a local Kubernetes cluster
---


## minikube logs

Returns logs to debug a local Kubernetes cluster

### Synopsis

Gets the logs of the running instance, used for debugging minikube, not user code.

```shell
minikube logs [flags]
```

### Options

```
      --audit             Show only the audit logs
      --file string       If present, writes to the provided file instead of stdout.
  -f, --follow            Show only the most recent journal entries, and continuously print new entries as they are appended to the journal.
      --last-start-only   Show only the last start logs.
  -n, --length int        Number of lines back to go within the log (default 60)
      --node string       The node to get logs from. Defaults to the primary control plane.
      --problems          Show only log entries which point to known problems
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

