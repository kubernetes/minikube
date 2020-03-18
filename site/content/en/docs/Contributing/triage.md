---
linkTitle: "Triage"
title: "Triaging Minikube Issues"
date: 2020-03-17
weight: 10
description: >
  How to triage issues in the minikube repo
---

Triage is an important part of maintaining the health of the minikube repo.
A well organized repo allows maintainers to prioritize feature requests, fix bugs, and respond to users facing difficulty with the tool as quickly as possible.

Triage includes:
- Labeling issues
- Responding to issues
- Closing issues (under certain circumstances!)

If you're interested in helping out with minikube triage, this doc covers the basics of doing so.

Additionally, if you'd be interested in participating in our weekly triage meeting, please fill out this [form](https://forms.gle/vNtWZSWXqeYaaNbU9) to express interest. Thank you! 

# Labeling Issues





# Responding to Issues

Many issues in the minikube repo fall into one of the following categories:
- Needs more information from the author to be actionable
- Duplicate Issue


## Closing with Care

Issues typically need to be closed for the following reasons:

- The issue has been addressed
- The issue is a duplicate of an existing issue
- There has been a lack of information over a long period of time

In any of these situations, we aim to be kind when closing the issue, and offer the author action items should they need to reopen their issue or still require a solution.

Samples responses for these situations include:

### Issue has been addressed

@author: I believe this issue is now addressed by minikube v1.4, as it <reason>. If you still see this issue with minikube v1.4 or higher, please reopen this issue by commenting with `/reopen`

Thank you for reporting this issue!

### Duplicate Issue


This issue appears to be a duplicate of #X, do you mind if we move the conversation there?

This way we can centralize the content relating to the issue. If you feel that this issue is not in fact a duplicate, please re-open it using `/reopen`. If you have additional information to share, please add it to the new issue.

Thank you for reporting this!

### Lack of Information

Hey @author -- hopefully it's OK if I close this - there wasn't enough information to make it actionable, and some time has already passed. If you are able to provide additional details, you may reopen it at any point by adding /reopen to your comment.

Here is additional information that may be helpful to us:

* Whether the issue occurs with the latest minikube release
*  The exact `minikube start` command line used
*  The full output of the `minikube start` command, preferably with `--alsologtostderr -v=3` for extra logging.
 * The full output of `minikube logs`

Thank you for sharing your experience!
