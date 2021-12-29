---
title: "Addons"
weight: 4
description: >
  How to develop minikube addons
---

## Creating a new addon

To create an addon, first fork the minikube repository, and check out your fork:

```shell
git clone git@github.com:<username>/minikube.git
```

Then go into the source directory:

```shell
cd minikube
```

Create a subdirectory:

```shell
mkdir deploy/addons/<addon name>
```

Add your manifest YAML's to the directory you have created:

```shell
cp *.yaml deploy/addons/<addon name>
```

Create an addon manifest file describing how to activate the addon, in the addon directory. Here is the manifest for the `dashboard` addon located at `deploy/addons/dashboard/dashboard.addon.yaml`:

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

The `maintainer` property is meant to inform users about the controlling party of an addon's images. In the case above, the maintainer is Kubernetes, since the registry addon uses images that Kubernetes controls. When creating a new addon, the source of the images should be contacted and requested whether they are willing to be the point of contact for this addon before being put. If the source does not accept the responsibility, leaving the maintainer field empty is acceptable.

Note: If the addon never needs authentication to GCP, then consider adding the following label to the pod's yaml:

```yaml
gcp-auth-skip-secret: "true"
```

If you need to customize how the addon is validated and activated, add it to `pkg/addons/config.go`. Here is the entry used by the `gvisor` addon, which has the `IsRuntimeContainerd` validation and `verifyAddonStatus` callback:

```go
	{
		name:        "gvisor",
		set:         SetBool,
		validations: []setFn{IsRuntimeContainerd},
		callbacks:   []setFn{EnableOrDisableAddon, verifyAddonStatus},
	},
```

Next, add all required files using `//go:embed` directives to a new embed.FS variable in `deploy/addons/assets.go`. Here is the entry used by the `csi-hostpath-driver` addon:

```go
	// CsiHostpathDriverAssets assets for csi-hostpath-driver addon
	//go:embed csi-hostpath-driver/deploy/*.tmpl csi-hostpath-driver/rbac/*.tmpl
	CsiHostpathDriverAssets embed.FS
```

At the bottom of the same file, add the plugin to the list of embedded plugins. Here is the entry used by the `csi-hostpath-driver` addon:

```go
var Embedded = map[string]embed.FS{
	// ...
	"addon.CsiHostpathDriverAssets":         CsiHostpathDriverAssets,
}
```

Then register the addon manifest path to the `deploy/addons/addon-registry.yaml` file. Here is the entry used by the `csi-hostpath-driver` addon:

```yaml
addons:
  // ...
  - embedfs://addon.CsiHostpathDriverAssets/csi-hostpath-driver/csi-hostpath-driver.addon.yaml
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

## Testing addon changes

Rebuild the minikube binary and apply the addon with extra logging enabled:

```shell
make && make test && ./out/minikube addons enable <addon name> --alsologtostderr
```

Please note that you must run `make` each time you change your YAML files. To disable the addon when new changes are made, run:

```shell
./out/minikube addons disable <addon name> --alsologtostderr
```

## Sending out your PR

Once you have tested your addon, click on [new pull request](https://github.com/kubernetes/minikube/compare) to send us your PR!
