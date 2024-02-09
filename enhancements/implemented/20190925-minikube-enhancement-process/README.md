# minikube enhancement process

First proposed: 2019-09-25
Authors: tstromberg

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's principles?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?

Please leave the above text in your proposal as instructions to the reader.

## Summary

A design review process for non-trivial enhancements to minikube.

## Goals

* Facilitate communication about the "how" and "why" of an enhancement before code is written
* Lightweight enough to not deter casual contributions
* A process applicable to any roadmap-worthy enhancement

## Non-Goals

* Coverage for smaller enhancements that would not be represented within the minikube roadmap.
* Reduced development velocity

## Design Details

The *minikube enhancement process (MEP)* is a way to propose, communicate, and coordinate on new efforts for the minikube project. MEP is based on a simplification of the [Kubernetes Enhancement Process](https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/0000-kep-process/README.md).

### Proposal Workflow

1. Copy `template.md` to `proposed/<date>-title.md`
1. Send PR out for review, titled: `Proposal: <title>`
1. Proposal will be discussed at the bi-weekly minikube office hours
1. After a 2-week review window, the proposal can be merged once there are 3 approving maintainers or reviewers. To keep proposals neutral, each reviewer must be independent and/or represent a different company.

### Implementation Workflow

1. In your PR that implements the enhancement, move the proposal to the `implemented/` folder.

## Alternatives Considered

### Kubernetes Enhancement Process

KEP's are a well-understood, but lengthier process geared toward making changes where multiple Kubernetes SIG's are affected. 

#### Pro's

* Easily facilitate input from multiple SIG's
* Clear, well understood process within Kubernetes, shared by multiple projects

#### Con's

* Invisible to casual contributors to a project, as these proposals do not show up within the GitHub project page
* Lengthy template (1870+ words) that prompts for information that is not relevant to minikube
* Time commitment deters casual contribution

### Google Docs Proposal Template

Rather than maintaining Markdown documents in the minikube repository, we could use a Google Docs template, and then a Google Sheet to track proposal status.

### Pro's

* Easier editing for trivial proposals

### Con's

* Authors may waste unnecessary time styling output
* Styling may be inconsistent between proposals
* Invisible to casual contributors to a project, as these proposals do not show up within the GitHub project page



