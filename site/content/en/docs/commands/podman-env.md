---
title: "podman-env"
description: >
  Configure environment to use minikube's Podman service
---

## Requirements

- **Recent Podman version with Docker API compatibility is required.**
- **Docker client is required** - `podman-env` uses Docker's client to communicate with Podman's Docker-compatible API.
- The `podman-env` command configures Docker client environment variables to connect to minikube's Podman service via its Docker-compatible API.

{{% pageinfo color="info" %}}
**Note:** This command sets up standard Docker environment variables (`DOCKER_HOST`, `DOCKER_TLS_VERIFY`, `DOCKER_CERT_PATH`) to connect to Podman's Docker-compatible socket. Use the regular `docker` command-line tool to interact with minikube's Podman service.
{{% /pageinfo %}}

## minikube podman-env

Configure environment to use minikube's Podman service via Docker API compatibility

### Synopsis

Sets up Docker client env variables to use minikube's Podman Docker-compatible service.

```shell
minikube podman-env [flags]
```

### Usage

After running `minikube podman-env`, you can use the regular Docker client to interact with minikube's Podman service:

```shell
# Configure your shell
eval $(minikube podman-env)

# Now use docker commands as usual - they will connect to Podman
docker images
docker build -t myapp .
docker run myapp
```

This approach provides Docker API compatibility while using Podman as the container runtime inside minikube.

### Building Images for Local Development

You can build images directly in minikube and deploy them without a separate registry:

```shell
# Configure environment
eval $(minikube podman-env)

# Build image directly in minikube
docker build -t my-local-app .

# Deploy to Kubernetes without registry
kubectl run my-app --image=my-local-app --image-pull-policy=Never
```

### Options

```
      --shell string   Force environment to be configured for a specified shell: [fish, cmd, powershell, tcsh, bash, zsh], default is auto-detect
  -u, --unset          Unset variables instead of setting them
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

