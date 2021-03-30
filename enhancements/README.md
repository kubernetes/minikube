# Enhancements

The *minikube enhancement process (MEP)* is a way to propose, communicate, and coordinate on new efforts for the minikube project. You can read the full details in the [MEP proposal](implemented/20190925-minikube-enhancement-process/README.md)

MEP is based on a simplification of the [Kubernetes Enhancement Process](https://github.com/kubernetes/enhancements/blob/master/keps/sig-architecture/0000-kep-process/README.md).

## Workflow

1. Copy `template.md` to `proposed/<date>-title/README.md`. Include supporting documents in the `proposed/<date>-title/` directory.
1. Send PR out for review, titled: `MEP: <title>`
1. Proposal will be discussed at the bi-weekly minikube office hours
1. After a 2-week review window, the proposal can be merged once there are 3 approving maintainers or reviewers. To keep proposals neutral, each reviewer must be independent and/or represent a different company.
1. In your PR that implements the enhancement, move the proposal to the `implemented/` folder.
