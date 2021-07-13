---
title: "ISO"
description: >
  How to release a new minikube ISO
---

Major releases always get a new ISO. Minor bugfixes may or may not require it: check for changes in the `deploy/iso` folder.
To check, run `git log -- deploy/iso` from the root directory and see if there has been a commit since the most recent release.

Note: you can build the ISO using the `hack/jenkins/build_iso.sh` script locally.

* Navigate to the minikube ISO jenkins job
* Ensure that you are logged in (top right)
* Click "▶️ Build with Parameters" (left)
* For `ISO_VERSION`, type in the intended release version (same as the minikube binary's version)
* For `ISO_BUCKET`, type in `minikube/iso`
* Click *Build*

The build will take roughly 50 minutes and will automatically create a PR with the changes.
