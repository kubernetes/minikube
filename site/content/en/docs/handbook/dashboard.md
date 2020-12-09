---
title: "Dashboard"
weight: 4
description: >
  Dashboard
aliases:
  - /docs/tasks/dashboard/
---

minikube has integrated support for the [Kubernetes Dashboard UI](https://github.com/kubernetes/dashboard).

## Overview

The Dashboard is a web-based Kubernetes user interface. You can use it to:

- deploy containerized applications to a Kubernetes cluster
- troubleshoot your containerized application
- manage the cluster resources
- get an overview of applications running on your cluster
- creating or modifying individual Kubernetes resources (such as Deployments, Jobs, DaemonSets, etc)

For example, you can scale a Deployment, initiate a rolling update, restart a pod or deploy new applications using a deploy wizard.

## Basic usage

To access the dashboard:

```shell
minikube dashboard
```

This will enable the dashboard add-on, and open the proxy in the default web browser.

It's worth noting that web browsers generally do not run properly as the root user, so if you are
in an environment where you are running as root, see the URL-only option.

To stop the proxy (leaves the dashboard running), abort the started process (`Ctrl+C`).

## Getting just the dashboard URL

If you don't want to open a web browser, the dashboard command can also simply emit a URL:

```shell
minikube dashboard --url
```

## Reference

For additional information, see [the official Dashboard documentation](https://kubernetes.io/docs/tasks/access-application-cluster/web-ui-dashboard/).
