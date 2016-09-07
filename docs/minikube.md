## minikube

Minikube is a tool for managing local Kubernetes clusters.

### Synopsis


Minikube is a CLI tool that provisions and manages single-node Kubernetes clusters optimized for development workflows.

### Options

```
      --alsologtostderr[=false]: log to standard error as well as files
      --log-flush-frequency=5s: Maximum number of seconds between log flushes
      --log_backtrace_at=:0: when logging hits line file:N, emit a stack trace
      --log_dir="": If non-empty, write log files in this directory
      --logtostderr[=false]: log to standard error instead of files
      --show-libmachine-logs[=false]: Whether or not to show logs from libmachine.
      --stderrthreshold=2: logs at or above this threshold go to stderr
      --v=0: log level for V logs
      --vmodule=: comma-separated list of pattern=N settings for file-filtered logging
```

### SEE ALSO
* [minikube config](minikube_config.md)	 - Modify minikube config
* [minikube dashboard](minikube_dashboard.md)	 - Opens/displays the kubernetes dashboard URL for your local cluster
* [minikube delete](minikube_delete.md)	 - Deletes a local kubernetes cluster.
* [minikube docker-env](minikube_docker-env.md)	 - sets up docker env variables; similar to '$(docker-machine env)'
* [minikube get-k8s-versions](minikube_get-k8s-versions.md)	 - Gets the list of available kubernetes versions available for minikube.
* [minikube ip](minikube_ip.md)	 - Retrieve the IP address of the running cluster.
* [minikube logs](minikube_logs.md)	 - Gets the logs of the running localkube instance, used for debugging minikube, not user code.
* [minikube service](minikube_service.md)	 - Gets the kubernetes URL for the specified service in your local cluster
* [minikube ssh](minikube_ssh.md)	 - Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'
* [minikube start](minikube_start.md)	 - Starts a local kubernetes cluster.
* [minikube status](minikube_status.md)	 - Gets the status of a local kubernetes cluster.
* [minikube stop](minikube_stop.md)	 - Stops a running local kubernetes cluster.
* [minikube version](minikube_version.md)	 - Print the version of minikube.

