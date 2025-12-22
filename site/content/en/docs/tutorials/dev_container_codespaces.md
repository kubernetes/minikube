---
title: "Run minikube in a Dev Container or in a Github Codespace"
linkTitle: "minikube in a Dev Container or in a Github Codespace"
weight: 1
date: 2025-12-12
description: >
  Running minikube and the minikube Dashboard inside a Dev Container or in a Github Codespace
---

This tutorial shows how to run minikube in a Github Codespace or locally in a Dev Container.

## Prerequisites

### Github Codespace

- A Github account.

[![Open in GitHub Codespaces](https://img.shields.io/badge/Open%20in-GitHub%20Codespaces-blue?logo=github)](https://codespaces.new/kubernetes/minikube?quickstart=1)

### Locally in a Dev Container

- Docker installed locally. [Docker](https://www.docker.com/get-started/)
- Visual Studio Code installed locally. [VSCode](https://code.visualstudio.com/)
- Install the Dev Containers extension for VSCode. [Dev Containers extension](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers)

## Instructions

- Open the project in VSCode or Github Codespace
- Once the the Workspace is open, you can launch minikube as usual 
  
  If you have an existing minikube instance, you may need to delete it if it was built before installing the AMD drivers.
  ```shell
  minikube start
  ```
  
  ![Start minikube](/images/docs_tutorials_dev_container_codespaces_minikube_start.png)

- minikube dashboard:
  ```shell
  minikube dashboard
  ```

  Once minikube dashboard is running you can open the URL locally by clicking on the link displayed or in Github Codespace in the ports tab by opening the URL in a broser window and navigating to the dashboard URL

  ![Start minikube Dashboard](/images/docs_tutorials_dev_container_codespaces_minikube_dashboard.png)

  ![Open minikube Dashboard](/images/docs_tutorials_dev_container_codespaces_minikube_dashboard_browser_window.png)

  On Github Codespaces the minikube Dashboard URL needs to be modified to include the Codespace URL and add the Dashboard path /api/v1/namespaces/kubernetes-dashboard/services/http:kubernetes-dashboard:/proxy/

  ![Open minikube Dashboard](/images/docs_tutorials_dev_container_codespaces_minikube_dashboard_browser_window_codespace.png)

  ![Open minikube Dashboard](/images/docs_tutorials_dev_container_codespaces_minikube_dashboard_browser_without_path_codespace.png)

  ![Open minikube Dashboard](/images/docs_tutorials_dev_container_codespaces_minikube_dashboard_codespace.png)

## Where can I learn more about Dev Containers?

See the excellent documentation at
<https://containers.dev/>

## Where can I learn more about Github Codespaces?

See the excellent documentation at
<https://github.com/features/codespaces>
