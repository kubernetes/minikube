---
title: "completion"
description: >
  Generate command completion for a shell
---


## minikube completion

Generate command completion for a shell

### Synopsis

Outputs minikube shell completion for the given shell (bash, zsh or fish)

	This depends on the bash-completion binary.  Example installation instructions:
	OS X:
		$ brew install bash-completion
		$ source $(brew --prefix)/etc/bash_completion
		$ minikube completion bash > ~/.minikube-completion  # for bash users
		$ minikube completion zsh > ~/.minikube-completion  # for zsh users
		$ source ~/.minikube-completion
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish # for fish users
	Ubuntu:
		$ apt-get install bash-completion
		$ source /etc/bash-completion
		$ source <(minikube completion bash) # for bash users
		$ source <(minikube completion zsh) # for zsh users
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish # for fish users

	Additionally, you may want to output the completion to a file and source in your .bashrc

	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
	Note for fish users: [2] please refer to this docs for more details https://fishshell.com/docs/current/#tab-completion


```
minikube completion SHELL [flags]
```

### Options

```
  -h, --help   help for completion
```

### Options inherited from parent commands

```
      --alsologtostderr                  log to standard error as well as files
  -b, --bootstrapper string              The name of the cluster bootstrapper that will set up the Kubernetes cluster. (default "kubeadm")
      --log_backtrace_at traceLocation   when logging hits line file:N, emit a stack trace (default :0)
      --log_dir string                   If non-empty, write log files in this directory
      --logtostderr                      log to standard error instead of files
  -p, --profile string                   The name of the minikube VM being used. This can be set to allow having multiple instances of minikube independently. (default "minikube")
      --stderrthreshold severity         logs at or above this threshold go to stderr (default 2)
  -v, --v Level                          log level for V logs
      --vmodule moduleSpec               comma-separated list of pattern=N settings for file-filtered logging
```

