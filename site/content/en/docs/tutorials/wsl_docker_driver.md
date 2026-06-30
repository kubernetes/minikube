---
title: "Setting Up WSL 2 for the Docker driver"
linkTitle: "Setting Up WSL 2"
weight: 1
date: 2022-01-12
---

## Overview

This guide shows how to prepare WSL 2 so you can run minikube with the Docker driver.

If you install Docker inside your WSL distribution, follow the standard Linux Docker installation guide and complete the post-install steps for non-root usage.

If you use Docker Desktop, make sure WSL integration is enabled for the distribution where you run `minikube`.

## Steps

1. [Install WSL](https://learn.microsoft.com/windows/wsl/install).
2. Choose one Docker setup:
   - Install Docker inside your WSL distribution by following the [Docker Engine installation guide](https://docs.docker.com/engine/install/) and the [Linux post-install steps](https://docs.docker.com/engine/install/linux-postinstall/).
   - Or install and configure [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/wsl/#download), then enable WSL integration for your distribution.
3. Verify that Docker works from inside WSL:

   ```shell
   docker version
   docker ps
   ```

4. Start minikube with the Docker driver:

   ```shell
   minikube start --driver=docker
   ```

## Accessing services on WSL

When you use the Docker driver on Windows or WSL, the node IP is not reachable directly. If `minikube ip` prints an address such as `127.0.0.1` or `192.168.49.2`, do not rely on that alone to access your workloads from a browser.

- For `NodePort` services, use `minikube service <service-name> --url` and keep that terminal open while you access the forwarded URL.
- For `LoadBalancer` services or ingress, run `minikube tunnel` in a separate terminal and keep it running while you access the service.
- For ingress on WSL with the Docker driver, use `127.0.0.1` or `localhost` together with `minikube tunnel` rather than the node IP.

See also:

- [Accessing apps]({{< ref "/docs/handbook/accessing/" >}})
- [Docker driver]({{< ref "/docs/drivers/docker/" >}})
