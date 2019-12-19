---
title: "Releasing"
date: 2019-07-31
weight: 9
description: >
  How to release minikube
---

## Preparation

* Announce release intent on #minikube
* Pause merge requests so that they are not accidentally left out of the ISO or release notes
* Two minikube repos checked out locally:
  * Your personal fork
  * Upstream  

## Build a new ISO

Major releases always get a new ISO. Minor bugfixes may or may not require it: check for changes in the `deploy/iso` folder.
To check, run `git log -- deploy/iso` from the root directory and see if there has been a commit since the most recent release.

Note: you can build the ISO using the `hack/jenkins/build_iso.sh` script locally.

* Navigate to the minikube ISO jenkins job
* Ensure that you are logged in (top right)
* Click "▶️ Build with Parameters" (left)
* For `ISO_VERSION`, type in the intended release version (same as the minikube binary's version)
* For `ISO_BUCKET`, type in `minikube/iso`
* Click *Build*

The build will take roughly 50 minutes.

## Update Makefile

Edit the minikube `Makefile`, updating the version number values at the top:

* `VERSION_MAJOR`, `VERSION_MINOR`, `VERSION_BUILD` as necessary
* `ISO_VERSION` - defaults to MAJOR.MINOR.0 - update if point release requires a new ISO to be built.

Make sure the integration tests run against this PR, once the new ISO is built. 

You can merge this change at any time before the release, but often the Makefile change is merged the next step: Release notes.

## Update Release Notes

Run the following script from your local upstream repo copy to generate updated release notes:

```shell
hack/release_notes.sh
```

Paste the output into CHANGELOG.md. See [PR#3175](https://github.com/kubernetes/minikube/pull/3175) as an example. 

You'll need to massage the output in a few key ways:

- The changelog should only contain user facing change. This means removing PR's for:
  - Documentation
  - Low-risk refactors
  - Test-only changes 
- Sort the changes so that the ones users will want to know about the most appear first
- Remove bots from the contributor list
- Remove duplicated similar names from the contributor list

Merge the output into CHANGELOG.md. See [PR#3175](https://github.com/kubernetes/minikube/pull/3175) as an example. 

## Tag the Release

```shell
sh hack/tag_release.sh 1.<minor>.<patch>
```

## Build the Release

This step uses the git tag to publish new binaries to GCS and create a github release:

* Navigate to the minikube "Release" jenkins job
* Ensure that you are logged in (top right)
* Click "▶️ Build with Parameters" (left)
* `VERSION_MAJOR`, `VERSION_MINOR`, and `VERSION_BUILD` should reflect the values in your Makefile
* For `ISO_SHA256`, run: `gsutil cat gs://minikube/iso/minikube-v<version>.iso.sha256`
* Click *Build*

## Check the release logs

After job completion, click "Console Output" to verify that the release completed without errors. This is typically where one will see brew automation fail, for instance.

**Note: If you are releasing a beta, you are done when you get here.**

## Check releases.json

This file is used for auto-update notifications, but is not active until releases.json is copied to GCS.

minikube-bot will send out a PR to update the release checksums at the top of `deploy/minikube/releases.json`. You should merge this PR.

## Update documentation link

Update `latest_release` in `site/config.toml`

example: https://github.com/kubernetes/minikube/pull/5413

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

## Announce

Please mention the new release https://github.com/kubernetes/minikube/blob/master/README.md

Other places:

- #minikube on Slack
- minikube-dev, minikube-users mailing list
- Twitter (optional!)
