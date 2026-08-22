---
title: "Dependency Versions"
linkTitle: "Dependency Versions"
date: 2026-08-21
weight: 6
description: >
  How minikube keeps third-party dependency versions up to date
---

Most third-party versions in this repository are **updated automatically**. Before
bumping a version by hand, check whether an updater already owns it.

## How it works

Each dependency has an updater under [`hack/update/`](https://github.com/kubernetes/minikube/tree/master/hack/update)
that declares a schema: a set of files, and a regular expression to replace in each one.
A matching `update-*.yml` GitHub workflow runs the updater on a weekly cron, and if
anything changed, opens a PR from `minikube-bot`.

To run an updater locally:

```shell
make update-golang-version
```

To see the version a dependency is currently pinned to:

```shell
DEP=golang make get-dependency-version
```

Because these files are rewritten by regular expression, hand-editing a managed
version is not durable: the next scheduled run will overwrite it.

## Go versions

Go is the one dependency where two different values are tracked on purpose, so it is
worth understanding before changing either.

| Location | Example | Meaning |
| -------- | ------- | ------- |
| `go.mod` `go` directive | `go 1.26.0` | The **minimum language version** minikube compiles against. The patch component is always `0`. |
| `GO_VERSION` in the `Makefile` and in `.github/workflows/*.yml` | `1.26.5` | The **exact toolchain patch** used to build and test minikube. |

[`hack/update/golang_version/golang_version.go`](https://github.com/kubernetes/minikube/blob/master/hack/update/golang_version/golang_version.go)
maintains both from a single upstream source, the [Kubernetes cross-build Go version](https://github.com/kubernetes/kubernetes/blob/master/build/build-image/cross/VERSION),
so minikube builds with the same Go release as Kubernetes itself. It rewrites the
`GO_VERSION` line in every file in `.github/workflows/`, discovered by listing the
directory, so a newly added workflow is picked up with no extra wiring. It also
updates the Makefile, the Jenkins and Prow Go installers, and the gvisor and
auto-pause Dockerfiles.

### Do not replace GO_VERSION with go-version-file

It is tempting to drop the `GO_VERSION` variable and point `actions/setup-go` at
`go.mod` instead:

```yaml
# Don't do this.
- uses: actions/setup-go@v5
  with:
    go-version-file: 'go.mod'
```

`setup-go` reads the `go` directive literally and coerces it with semver, so
`go 1.26.0` resolves to exactly Go 1.26.0 rather than the latest 1.26 patch. Since the
`go` directive intentionally pins the patch component to `0`, this **downgrades CI to
the compatibility floor** and drifts away from the `GO_VERSION` still used by local
builds, the Jenkins and Prow installers, and the images. Writing `go 1.26` would not
help, as semver coerces it to the same `1.26.0`.

The `GO_VERSION` variable is already generated, so it carries no manual maintenance
cost. Leave it in place.
