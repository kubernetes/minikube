# Steps to Release Minikube

## Create a Release Notes PR

Assemble all the meaningful changes since the last release into the CHANGELOG.md file.
See [this PR](https://github.com/kubernetes/minikube/pull/164) for an example.

## Build and Release a New ISO

This step isn't always required. Check if there were changes in the deploy directory.
If you do this, bump the ISO URL to point to the new ISO, and send a PR.
To do this, build the new iso by running:
```shell
deploy/iso/build.sh
```
This will generate a new iso at 'deploy/iso/minikube.iso'.  Then upload the iso and shasum using the following command:
```shell
gsutil cp deploy/iso/minikube.iso gs://minikube/minikube-<increment.version>.iso
gsutil cp deploy/iso/minikube.iso.sha256 gs://minikube/minikube-<increment.version>.iso.sha256
```

## Run integration tests

Run this command:
```shell
make integration
```
Investigate and fix any failures.

## Bump the version in the Makefile and Update Docs to reflect this

See [this PR](https://github.com/kubernetes/minikube/pull/165) for an example.

##Send an initial commit with the Makefile change:

Send a PR for the Makefile change and wait until it is merged.  Once the commit is merged, continue.

## Build the Release

Run this command:

```shell
BUILD_IN_DOCKER=y make cross checksum
```

## Add the version to the releases.json file

Add an entry **to the top** of deploy/minikube/releases.json with the **version** and **checksums**.
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
gsutil cp out/minikube-windows-amd64.exe gs://minikube/releases/$RELEASE/
gsutil cp out/minikube-windows-amd64.exe.sha256 gs://minikube/releases/$RELEASE/
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
gsutil cp -r gs://minikube/releases/$RELEASE/* gs://minikube/releases/latest/
```

## Package managers which include minikube

These are downstream packages that are being maintained by others and how to upgrade them to make sure they have the latest versions

| Package Manager | URL | TODO |
| --- | --- | --- |
| Arch Linux AUR | https://aur.archlinux.org/packages/minikube/ | "Flag as package out-of-date"
| Brew Cask | https://github.com/caskroom/homebrew-cask/blob/master/Casks/minikube.rb | Create a new PR in [caskroom/homebrew-cask](https://github.com/caskroom/homebrew-cask) with an updated version and appcast checkpoint

#### [How to Generate an Appcast Checkpoint for Homebrew](https://github.com/caskroom/homebrew-cask/blob/master/doc/cask_language_reference/stanzas/appcast.md)
`curl --compressed --location --user-agent 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/33.0.1750.152 Safari/537.36' "https://github.com/kubernetes/minikube/releases.atom" | sed 's|<pubDate>[^<]*</pubDate>||g' | shasum --algorithm 256`

#### Updating the arch linux package
The Arch Linux AUR is maintained at https://aur.archlinux.org/packages/minikube/.  The installer PKGBUILD is hosted in its own repository.  The public read-only repository is hosted here `https://aur.archlinux.org/minikube.git` and the private read-write repository is hosted here `ssh://aur@aur.archlinux.org/minikube.git`

The repository is tracked in this repo under a submodule `installers/linux/arch_linux`.  Currently, its configured to point at the public readonly repository so if you want to push you should run this command to overwrite

`git config submodule.archlinux.url ssh://aur@aur.archlinux.org/minikube.git `

To actually update the package, you should bump the version and update the sha512 checksum.  You should also run `makepkg --printsrcinfo > .SRCINFO` to update the srcinfo file.  You can edit this manually if you don't have `makepkg` on your machine.

## Release Verification

After you've finished the release, run this command from the release commit to verify the release was done correctly:
`make check-release`.

## Update kubernetes.io docs

If there are major changes, please send a PR upstream for this file https://github.com/kubernetes/kubernetes.github.io/blob/master/docs/getting-started-guides/minikube.md in order to keep the getting started guide up to date.
