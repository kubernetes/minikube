# Steps to Release Minikube

## Create a Release Notes PR

Assemble all the meaningful changes since the last release into the CHANGELOG.md file.
See [this PR](https://github.com/kubernetes/minikube/pull/164) for an example.

## Build and Release a New ISO

This step isn't always required. Check if there were changes in the deploy directory.
If you do this, bump the ISO URL to point to the new ISO, and send a PR.

## Bump the version in the Makefile

See [this PR](https://github.com/kubernetes/minikube/pull/165) for an example.

## Run integration tests

Run this command:
```shell
make integration
```
Investigate and fix any failures.

## Build the Release

Run this command:

```shell
make cross checksum
```

## Add the version to the releases.json file

Add an entry **to the top** of deploy/minikube/releases.json with the version and checksums.
Send a PR.
This file controls the auto update notifications in minikube.
Only add entries to this file that should be released to all users (no pre-release, alpha or beta releases).
The file must be uploaded to GCS before notifications will go out. That step comes at the end.

The schema for this file can be found in deploy/minikube/schema.json.

An automated test to verify the schema runs in Travis before each submit.

## Upload to GCS:

```shell
gsutil cp out/minikube-linux-amd64 gs://minikube/releases/$RELEASE/
gsutil cp out/minikube-linux-amd64.sha256 gs://minikube/releases/$RELEASE/
gsutil cp out/minikube-darwin-amd64 gs://minikube/releases/$RELEASE/
gsutil cp out/minikube-darwin-amd64.sha256 gs://minikube/releases/$RELEASE/
```

## Tag the Release

Run a command like this to tag it locally: `git tag -a v0.2.0 -m "0.2.0 Release"`.

And run a command like this to push the tag: `git push upstream v0.2.0`.

## Create a Release in Github

Create a new release based on your tag, like [this one](https://github.com/kubernetes/minikube/releases/tag/v0.2.0).

Upload the files, and calculated checksums.

## Upload the releases.json file to GCS

This step makes the new release trigger update notifications in old versions of Minikube.
Use this command from a clean git repo:

```shell
gsutil cp deploy/minikube/releases.json gs://minikube/releases.json
```

## Mark the release as `latest` in GCS:

```shell
gsutil cp -r gs://minikube/releases/$RELEASE gs://minikube/releases/latest
```
