---
title: "Binaries"
weight: 9
description: >
  How to release minikube binaries
---

## Preparation

* Announce release intent on #minikube
* Pause merge requests so that they are not accidentally left out of the ISO or release notes
* Two minikube repos checked out locally:
  * Your personal fork
  * Upstream  
  
## Update the Kubernetes version

* Run `make update-kubernetes-version` from your local upstream repo copy
* If any files are updated, create and merge a PR before moving forward  

## Build a new ISO

* All non-patch releases require a new ISO to be built.
* Patch releases (vx.x.1+) require a new ISO if the `deploy/iso` directory has seen changes since the previous release.

See [ISO release instructions]({{<ref "iso.md">}})

## Release new kicbase image

Run the `kic-release` job in Jenkins, which will automatically create a PR which must be merged (make sure to enter the correct version and repos).

## Update Release Notes

Run the following script from your local upstream repo copy to generate updated release notes:

```shell
make release-notes
```

Paste the output into CHANGELOG.md, sorting changes by importance to an end-user. If there are >8 changes, split them into *Improvements* and *Bug fixes*

- The changelog should only contain user facing change. This means removing PR's for:
  - Documentation
  - Low-risk refactors
  - Test-only changes
- Remove bots from the contributor list
- Remove duplicated similar names from the contributor list

You may merge this PR at any time, or combine it with a `Makefile` update PR.

## Update Makefile

Update the version numbers in  `Makefile`:

* `VERSION_MAJOR`, `VERSION_MINOR`, `VERSION_BUILD`

{{% alert title="Warning" color="warning" %}}
Merge this PR only if all non-experimental integration tests pass!
{{% /alert %}}

## Tag the Release

```shell
sh hack/tag_release.sh 1.<minor>.<patch>
```

## Build the Release

This step uses the git tag to publish new binaries to GCS and create a GitHub release:

* Navigate to the minikube "Release" jenkins job
* Ensure that you are logged in (top right)
* Click "▶️ Build with Parameters" (left)
* `VERSION_MAJOR`, `VERSION_MINOR`, and `VERSION_BUILD` should reflect the values in your Makefile
* For `ISO_SHA256_AMD64`, run: `gsutil cat gs://minikube/iso/minikube-v<version>-amd64.iso.sha256`
* For `ISO_SHA256_ARM64`, run: `gsutil cat gs://minikube/iso/minikube-v<version>-arm64.iso.sha256`
* Click *Build*

## Check the release logs

After job completion, click "Console Output" to verify that the release completed without errors. This is typically where one will see brew automation fail, for instance.

**Note: If you are releasing a beta, you are done when you get here.**

## Merge the releases.json change

The release script updates https://storage.googleapis.com/minikube/releases.json - which is used by the minikube binary to check for updates, and is live immediately.

minikube-bot will also send out a PR to merge this into the tree. Please merge this PR to keep GCS and GitHub in sync.

## Package managers which include minikube

These are downstream packages that are being maintained by others and how to upgrade them to make sure they have the latest versions

| Package Manager | URL                                                                       | TODO                                                                                                                                                                        |
| --------------- | ------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Arch Linux AUR  | <https://aur.archlinux.org/packages/minikube-bin/>                        | "Flag as package out-of-date"                                                                                                                                               |
| Brew Cask       | <https://github.com/Homebrew/homebrew-cask/tree/master/Casks> | The release job creates a new PR in [Homebrew/homebrew-cask](https://github.com/Homebrew/homebrew-cask) with an updated version and SHA256, double check that it's created. |

WARNING: The Brew cask automation is error-prone. please ensure that a PR was created.

## Verification

Verify release checksums by running `make check-release`

## Update docs

If there are major changes, please send a PR to update <https://kubernetes.io/docs/setup/learning-environment/minikube/>

## Update SECURITY-INSIGHTS.yml
Make appropriate changes to [SECURITY-INSIGHTS.yml](https://github.com/kubernetes/minikube/SECURITY-INSIGHTS.yml). Check [OPENSSF Security Insights Specification](https://github.com/ossf/security-insights-spec/blob/main/specification.md) for reference.

## Announce

Please mention the new release https://github.com/kubernetes/minikube/blob/master/README.md
Please notify the LWKD team of the new release at slack channel - #sig-contribex-comms

Other places:

- #minikube on Slack
- minikube-dev, minikube-users mailing list
- Twitter (this is now automated with the @minikube_dev account)
