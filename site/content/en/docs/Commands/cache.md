---
title: "cache"
linkTitle: "cache"
weight: 1
date: 2019-08-01
description: >
  Add or delete an image from the local cache.
---


## minikube cache add

Add an image to local cache.

```
minikube cache add [flags]
```

## minikube cache delete

Delete an image from the local cache.

```
minikube cache delete [flags]
```

## minikube cache list

List all available images from the local cache.

```
minikube cache list [flags]
```

### Options

```
      --format string   Go template format string for the cache list output.  The format for Go templates can be found here: https://golang.org/pkg/text/template/
                        For the list of accessible variables for the template, see the struct values here: https://godoc.org/k8s.io/minikube/cmd/minikube/cmd#CacheListTemplate (default "{{.CacheImage}}\n")
  -h, --help            help for list
```

## minikube cache reload

reloads images previously added using the 'cache add' subcommand

```
minikube cache reload [flags]
```