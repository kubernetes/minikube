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

Note: If the addon never needs authentication to GCP, then consider adding the following label to the pod's yaml:

```yaml
gcp-auth-skip-secret: "true"
```

To make the addon appear in `minikube addons list`, add it to `pkg/addons/config.go`. Here is the entry used by the `registry` addon, which will work for any addon which does not require custom code:

```go
  {
    name:      "registry",
    set:       SetBool,
    callbacks: []setFn{EnableOrDisableAddon},
  },
```

Then, add into `pkg/minikube/assets/addons.go` the list of files to copy into the cluster, including manifests. Here is the entry used by the `registry` addon:

```go
  "registry": NewAddon([]*BinAsset{
    MustBinAsset(
      "deploy/addons/registry/registry-rc.yaml.tmpl",
      vmpath.GuestAddonsDir,
      "registry-rc.yaml",
      "0640",
      false),
    MustBinAsset(
      "deploy/addons/registry/registry-svc.yaml.tmpl",
      vmpath.GuestAddonsDir,
      "registry-svc.yaml",
      "0640",
      false),
    MustBinAsset(
      "deploy/addons/registry/registry-proxy.yaml.tmpl",
      vmpath.GuestAddonsDir,
      "registry-proxy.yaml",
      "0640",
      false),
  }, false, "registry"),
```

The `MustBinAsset` arguments are:

* source filename
* destination directory (typically `vmpath.GuestAddonsDir`)
* destination filename
* permissions (typically `0640`)
* boolean value representing if template substitution is required (often `false`)

The boolean value on the last line is whether the addon should be enabled by default. This should always be `false`.

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
