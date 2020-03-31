---
title: "ssh"
linkTitle: "ssh"
weight: 1
date: 2019-08-01
description: >
  Log into or run a command on a machine with SSH; similar to 'docker-machine ssh'
---


### Usage

```
minikube ssh [flags]
```

### Options

```
  -h, --help         help for ssh
      --native-ssh   Use native Golang SSH client (default true). Set to 'false' to use the command line 'ssh' command when accessing the docker machine. Useful for the machine drivers when they will not start with 'Waiting for SSH'. (default true)
```
