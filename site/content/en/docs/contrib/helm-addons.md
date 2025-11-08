---
title: "Helm Based Addons"
weight: 5
description: >
  How to develop minikube addons using Helm charts.
---

## Overview

Minikube supports creating addons that are deployed via Helm charts. This allows for more complex applications to be managed as addons. This guide will walk you through creating a Helm-based addon.

For a general overview of creating addons, please see the [Creating a new addon](https://minikube.sigs.k8s.io/docs/contrib/addons/) guide first. This guide focuses on the specifics of Helm-based addons.

## Creating a Helm-based addon

Creating a Helm-based addon is very similar to creating a standard addon, with a few key differences in how the addon is defined.

### 1. Define the Addon in `pkg/minikube/assets/addons.go`

The core of a Helm-based addon is the `HelmChart` struct within your `Addon` definition. You will define your addon in `pkg/minikube/assets/addons.go`.

Here is an example of what a Helm-based addon definition looks like:

```go
"my-helm-addon": NewAddon(
    []*BinAsset{}, // Usually empty for pure Helm addons
    false,
    "my-helm-addon",
    "Your Name",
    "",
    "path/to/your/addon/docs.md",
    map[string]string{
        // Optional: Define images for caching if not in the chart
    },
    map[string]string{
        // Optional: Define registries for images
    },
    &HelmChart{
        Name:      "my-helm-addon-release",
        Repo:      "oci://my-repo/my-chart",
        Namespace: "my-addon-namespace",
        Values: []string{
            "key1=value1",
            "key2=value2",
        },
        ValueFiles: []string{
            // Paths to values files inside the minikube VM
        },
    },
),
```

#### `HelmChart` struct fields:

*   `Name`: The release name for the Helm installation (`helm install <release-name>`).
*   `Repo`: The Helm chart repository URL (e.g., `stable/chart-name` or `oci://my-repo/my-chart`).
*   `Namespace`: The Kubernetes namespace to install the chart into. The `--create-namespace` flag is always used.
*   `Values`: A slice of strings for setting individual values via `--set` (e.g., `key=value`).
*   `ValueFiles`: A slice of strings pointing to paths of YAML value files inside the minikube VM. These are passed to Helm with the `--values` flag.

When the addon is enabled, minikube will automatically ensure the `helm` binary is installed within the cluster and then run `helm upgrade --install` with the parameters you have defined. When disabled, it will run `helm uninstall`.

### 2. Add the Addon to `pkg/addons/config.go`

To make your addon visible to `minikube addons list` and enable it to be managed, you need to add an entry for it in `pkg/addons/config.go`.

```go
{
    name:      "my-helm-addon",
    set:       SetBool,
    callbacks: []setFn{EnableOrDisableAddon},
},
```

### 3. Testing your Helm Addon

To test your new addon, rebuild minikube and enable it:

```shell
make && ./out/minikube addons enable my-helm-addon
```

You can then verify that the Helm chart was deployed correctly using `kubectl`.

```bash
kubectl -n my-addon-namespace get pods
```