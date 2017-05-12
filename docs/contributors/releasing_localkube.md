# Steps to Release Localkube

## Configure The Correct Kubernetes Version to Build for Localkube
When building localkube for a specific Kubernetes version, follow the steps at [Updating Kubernetes](https://github.com/kubernetes/minikube/blob/master/docs/contributors/updating_kubernetes.md).  After you have setup a new folder and GOPATH with the desired version of Kubernetes (from the directions above), you go on to build localkube.

## Build the Localkube Release
```shell
make out/localkube
```

## Upload to GCS:

```shell
gsutil cp out/localkube  gs://minikube/k8sReleases/$K8S_RELEASE/localkube-linux-amd64
```

## Add the version to the k8s_releases.json file

Add an entry **in the appropriate version location** to deploy/minikube/k8s_releases.json with the version, and send a PR.
This file lists the available versions of localkube for minikube.
Only add entries to this file that should be released to all users (no pre-release, alpha or beta releases).

The schema for this file can be found in deploy/minikube/k8s_schema.json.

An automated test to verify the schema runs in Travis before each submit.

## Upload the releases.json file to GCS

This step makes the new release trigger update notifications in old versions of Minikube.
Use this command from a clean git repo:

```shell
gsutil cp deploy/minikube/k8s_releases.json gs://minikube/k8s_releases.json
```
