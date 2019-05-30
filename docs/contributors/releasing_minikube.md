# Steps to Release Minikube

## Preparation

* Announce release intent on #minikube
* Pause merge requests so that they are not accidentally left out of the ISO or release notes

## Build a new ISO

Major releases always get a new ISO. Minor bugfixes may or may not require it: check for changes in the `deploy/iso` folder.
To check, run `git log -- deploy/iso` from the root directory and see if there has been a commit since the most recent release.

Note: you can build the ISO using the `hack/jenkins/build_iso.sh` script locally.

* navigate to the minikube ISO jenkins job
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

## Ad-Hoc testing of other platforms

If there are supported platforms which do not have functioning Jenkins workers (Windows), you may use the following to build a sanity check:

```shell
env BUILD_IN_DOCKER=y make cross checksum
```

## Send out Makefile PR

Once submitted, HEAD will use the new ISO. Please pay attention to test failures, as this is our integration test across platforms. If there are known acceptable failures, please add a PR comment linking to the appropriate issue.

## Update Release Notes

Run the following script to update the release notes:

```shell
hack/release_notes.sh
```

Merge the output into CHANGELOG.md. See [PR#3175](https://github.com/kubernetes/minikube/pull/3175) as an example. Then get the PR submitted.

## Tag the Release

```shell
sh hack/tag_release.sh 1.<minor>.<patch>
```

## Build the Release

This step uses the git tag to publish new binaries to GCS and create a github release:

* navigate to the minikube "Release" jenkins job
* Ensure that you are logged in (top right)
* Click "▶️ Build with Parameters" (left)
* `VERSION_MAJOR`, `VERSION_MINOR`, and `VERSION_BUILD` should reflect the values in your Makefile
* For `ISO_SHA256`, run: `gsutil cat gs://minikube/iso/minikube-v<version>.iso.sha256`
* Click *Build*

## Check the release logs

After job completion, click "Console Output" to verify that the release completed without errors. This is typically where one  will see brew automation fail, for instance.

## Check releases.json

This file is used for auto-update notifications, but is not active until releases.json is copied to GCS.

minikube-bot will send out a PR to update the release checksums at the top of `deploy/minikube/releases.json`. You should merge this PR.

## Package managers which include minikube

These are downstream packages that are being maintained by others and how to upgrade them to make sure they have the latest versions

| Package Manager | URL | TODO |
| --- | --- | --- |
| Arch Linux AUR | <https://aur.archlinux.org/packages/minikube-bin/> | "Flag as package out-of-date"
| Brew Cask | <https://github.com/Homebrew/homebrew-cask/blob/master/Casks/minikube.rb> | The release job creates a new PR in [Homebrew/homebrew-cask](https://github.com/Homebrew/homebrew-cask) with an updated version and SHA256, double check that it's created.

WARNING: The Brew cask automation is error-prone. please ensure that a PR was created.

## Verification

Verify release checksums by running`make check-release`

## Update docs

If there are major changes, please send a PR to update <https://kubernetes.io/docs/setup/minikube/>

## Announce

Please mention the new release https://github.com/kubernetes/minikube/blob/master/README.md

Other places:

- #minikube on Slack
- minikube-dev, minikube-users mailing list
- Twitter (optional!)
