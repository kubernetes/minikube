<!--[metadata]>
+++
draft=true
+++
<![end-metadata]-->

# Docker Machine Release Process

The Docker Machine release process is fairly straightforward and as many steps
have been taken as possible to make it automated, but there is a procedure and
several "checklist items" which should be documented.  This document is intended
to cover the current Docker Machine release process.  It is written for Docker
Machine core maintainers who might find themselves performing a release.

0.  The new version of `azure` driver released in 0.7.0 is not backwards compatible
    and therefore errors out with a message saying the new driver is unsupported with
    the new version. The commit 7b961604 should be undone prior to 0.8.0 release and
    this notice must be removed from `docs/RELEASE.md`.
1.  **Get a GITHUB_TOKEN** Check that you have a proper `GITHUB_TOKEN`. This
    token needs only to have the `repo` scope. The token can be created on github
    in the settings > Personal Access Token menu.
2.  **Run the release script** At the root of the project, run the following
    command `GITHUB_TOKEN=XXXX script/release.sh X.Y.Z` where `XXXX` is the
    value of the GITHUB_TOKEN generated, `X.Y.Z` the version to release
    ( Explicitly excluding the 'v' prefix, the script takes care of it.). As of
    now, this version number must match the content of `version/version.go`. The
    script has been built to be as resilient as possible, cleaning everything
    it does along its way if necessary. You can run it many times in a row,
    fixing the various bits along the way.
3.  **Update the changelog on github** -- The script generated a list of all
    commits since last release. You need to edit this manually, getting rid of
    non critical details, and putting emphasis to what need our users attention.
4.  **Update the CHANGELOG.md** -- Add the same notes from the previous step to
    the `CHANGELOG.md` file in the repository.
5.  **Update the Documentation** -- Ensure that the `docs` branch on GitHub
    (which the Docker docs team uses to deploy from) is up to date with the
    changes to be deployed from the release branch / master.
6.  **Verify the Installation** -- Copy and paste the suggested commands in the
    installation notes to ensure that they work properly.  Best of all, grab an
    (uninvolved) buddy and have them try it.  `docker-machine -v` should give
    them the released version once they have run the install commands.
7.  (Optional) **Drink a Glass of Wine** -- You've worked hard on this release.
    You deserve it.  For wine suggestions, please consult your friendly
    neighborhood sommelier.
