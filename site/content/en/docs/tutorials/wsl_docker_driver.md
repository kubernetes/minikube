---
title: "Setting Up WSL 2 for the Docker driver"
linkTitle: "Setting Up WSL 2"
weight: 1
date: 2022-01-12
---

## Overview

This guide shows how to set up [WSL 2](https://learn.microsoft.com/windows/wsl/) on Windows so you can run minikube with the Docker driver, and how to reach your workloads from a Windows browser.

There are two supported ways to provide Docker inside WSL 2:

- **Docker Engine inside your WSL distribution** — install Docker directly in the Linux distribution where you run `minikube`.
- **Docker Desktop with WSL integration** — install Docker Desktop for Windows and enable WSL integration for the distribution where you run `minikube`.

## Steps

1. [Install WSL](https://learn.microsoft.com/windows/wsl/install) and confirm you are using WSL 2.

2. Set up Docker using one of the two options above:

   - To install Docker Engine inside WSL, follow the [Docker Engine install guide](https://docs.docker.com/engine/install/) and the [Linux post-install steps](https://docs.docker.com/engine/install/linux-postinstall/#manage-docker-as-a-non-root-user) so you can run Docker as a non-root user.
   - To use Docker Desktop, install [Docker Desktop for Windows](https://docs.docker.com/desktop/windows/wsl/) and enable WSL integration for your distribution under **Settings > Resources > WSL integration**.

3. From inside WSL, confirm Docker is working:

   ```shell
   docker version
   docker ps
   ```

4. Start minikube with the Docker driver:

   ```shell
   minikube start --driver=docker
   ```

## Accessing your services from Windows

With the Docker driver on Windows and WSL 2, the minikube node runs inside a container, so the node IP printed by `minikube ip` is **not** reachable directly from a Windows browser. Use one of the following instead:

- **NodePort services** — run the command below and browse to the printed `http://127.0.0.1:<port>` address. Keep the terminal open while you use the URL:

  ```shell
  minikube service <service-name> --url
  ```

- **LoadBalancer services** — run `minikube tunnel` in a separate terminal and keep it running, then reach the service on `127.0.0.1`:

  ```shell
  minikube tunnel
  ```

For more ways to reach your applications, see the [Accessing apps]({{<ref "/docs/handbook/accessing">}}) handbook page.

## See also

- [Docker driver]({{<ref "/docs/drivers/docker">}})
- [Accessing apps]({{<ref "/docs/handbook/accessing">}})
