---
title: "completion"
description: >
  Generate command completion for a shell
---


## minikube completion

Generate command completion for a shell

### Synopsis

Outputs minikube shell completion for the given shell (bash, zsh, fish or powershell)

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
		$ source /etc/bash_completion
		$ source <(minikube completion bash) # for bash users
		$ source <(minikube completion zsh) # for zsh users
		$ minikube completion fish > ~/.config/fish/completions/minikube.fish # for fish users

	Additionally, you may want to output the completion to a file and source in your .bashrc

	Windows:
		## Save completion code to a script and execute in the profile
		PS> minikube completion powershell > $HOME\.minikube-completion.ps1
		PS> Add-Content $PROFILE '. $HOME\.minikube-completion.ps1'

		## Execute completion code in the profile
		PS> Add-Content $PROFILE 'if (Get-Command minikube -ErrorAction SilentlyContinue) {
		        minikube completion powershell | Out-String | Invoke-Expression
		    }'

	Note for zsh users: [1] zsh completions are only supported in versions of zsh >= 5.2
	Note for fish users: [2] please refer to this docs for more details https://fishshell.com/docs/current/#tab-completion


```shell
minikube completion SHELL [flags]
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

## minikube completion bash

bash completion.

### Synopsis

Generate command completion for bash.

```shell
minikube completion bash [flags]
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

## minikube completion fish

fish completion.

### Synopsis

Generate command completion for fish .

```shell
minikube completion fish [flags]
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

## minikube completion help

Help about any command

### Synopsis

Help provides help for any command in the application.
Simply type completion help [path to command] for full details.

```shell
minikube completion help [command] [flags]
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

## minikube completion powershell

powershell completion.

### Synopsis

Generate command completion for PowerShell.

```shell
minikube completion powershell [flags]
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

## minikube completion zsh

zsh completion.

### Synopsis

Generate command completion for zsh.

```shell
minikube completion zsh [flags]
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

