---
title: "Custom Addons"
weight: 4
description: >
  How to create custom addons
---

## Create a local registry

Create a directory for your registry:

```shell
mkdir -p $HOME/<registry name>
```

Add your registry to the minikube configuration, `.minikube/config/addons.json`:

```json
    "customRegistries": [
        {
            "path": "$HOME/<registry name>",
            "enabled": true
        }
    ]
```

## Creating a new addon

Create a directory for your registry:

```shell
mkdir -p $HOME/<registry name>/<addon name>
```

Add your manifest YAML's to the directory you have created. Then create an addon manifest file describing how to activate the addon, in the addon directory. For example `~/<registry name>/<addon name>/<addon name>.yaml`. Here is the manifest for the `dashboard`:

```yaml
name: dashboard
maintainer: kubernetes
templates:
  - source: dashboard-dp.yaml.tmpl
    target: /etc/kubernetes/addons/dashboard-dp.yaml
assets:
  - source: dashboard-ns.yaml
    target: /etc/kubernetes/addons/dashboard-ns.yaml
  - source: dashboard-clusterrole.yaml
    target: /etc/kubernetes/addons/dashboard-clusterrole.yaml
  # ...
images:
  Dashboard:
    image: kubernetesui/dashboard:v2.3.1@sha256:ec27f462cf1946220f5a9ace416a84a57c18f98c777876a8054405d1428cc92e
  MetricsScraper:
    image: kubernetesui/metrics-scraper:v1.0.7@sha256:36d5b3f60e1a144cc5ada820910535074bdf5cf73fb70d1ff1681537eef4e172
```

The entries under `templates` and `assets` have the following properties:

* `source`: the source path
* `target`: destination path (typically `/etc/kubernetes/addons/<filename>`)
* `permissions`: the file permissions, `0640` by default

Note: If the addon never needs authentication to GCP, then consider adding the following label to the pod's yaml:

```yaml
gcp-auth-skip-secret: "true"
```

Then register the addon manifest path to the `~/<registry name>/addon-registry.yaml` file. For example:

```yaml
addons:
  // ...
  - <addon name>/<addon name>.addon.yaml
}
```

To see other examples, see the [addons commit history](https://github.com/kubernetes/minikube/commits/master/deploy/addons) for other recent examples.

## "addons open" support

If your addon contains a NodePort Service, please add the `kubernetes.io/minikube-addons-endpoint: <addon name>` label, which is used by the  `minikube addons open` command:

```yaml
apiVersion: v1
kind: Service
metadata:
 labels:
    kubernetes.io/minikube-addons-endpoint: <addon name>
```

NOTE: `minikube addons open` currently only works for the `kube-system` namespace: [#8089](https://github.com/kubernetes/minikube/issues/8089).

