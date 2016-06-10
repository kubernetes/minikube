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

## Tag the Release

Run a command like this to tag it locally: `git tag -a v0.2.0 -m "0.2.0 Release"`.

And run a command like this to push the tag: `git push upstream v0.2.0`.

## Build the Release

Run these commands:

```shell
GOOS=linux GOARCH=amd64 make out/minikube-linux-amd64
GOOS=darwin GOARCH=amd64 make out/minikube-darwin-amd64
```

## Upload to GCS:

```shell
gsutil cp out/minikube-linux-amd64  gs://minikube/releases/$RELEASE/
gsutil cp out/minikube-darwin-amd64  gs://minikube/releases/$RELEASE/
```

## Create a Release in Github

Create a new release based on your tag, like [this one](https://github.com/kubernetes/minikube/releases/tag/v0.2.0).

Upload the files, and calculate checksums.
