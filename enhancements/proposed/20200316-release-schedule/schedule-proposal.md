# Release Schedule

* First proposed: 2020-03-30
* Authors: Thomas Stromberg (@tstromberg)

## Reviewer Priorities

Please review this proposal with the following priorities:

* Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
* Are there other approaches to consider?
* Could the implementation be made simpler?
* Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Adding structure to the release process to encourage predictable stress-free releases with fewer regressions.

## Goals

* A decrease in release regressions
* Minimal disruption to development velocity
* Compatible with the upstream Kubernetes release schedule

## Non-Goals

* Maintaining release branches

## Design Details

minikube currently has 3 types of releases:

* Feature release (v1.9.0)
* Bugfix release (v1.9.1)
* Beta releases

This proposal maintains the pre-existing structure, but adds dates for when each step will occur:

* Day 0: Create milestones for the next regression & feature release
* Day 7: Regression release (optional)
* Day 14: Early Beta release (optional)
* Day 21: Beta release
* Day 24: Feature freeze and optional final beta
* Day 28: Feature release

To synchronize with Kubernetes release schedule (Tuesday afternoon PST), minikube releases should be Wednesday morning (PST). To select a final release date, consult [sig-release](https://github.com/kubernetes/sig-release/tree/master/releases) to see if there is an upcoming minor release of Kubernetes within the next 6 weeks. If so, schedule the minikube release to occur within 24 hours of it.

Even with this schedule, it is assumed that release dates may slip.

## Alternatives Considered

### Release branches

Rather than considering master to always be in a releasable state, we could maintain long-lived release branches. This adds a lot of overhead to the release manager, as they have to manage cherry-picks.

### Extending cycle by a week

As this process assumes a regression release at Day 7, it begs the question on whether or not a 5-week feature release cycle makes more sense:

* Day 0: Create milestones for the next regression & feature release
* Day 7: Regression release (optional)
* Day 21: Beta release
* Day 28: Beta 2 release
* Day 31: Feature freeze and optional final beta
* Day 35: Feature release

The downside is a slightly lower release velocity, the upside may be more a more stable final release.
