---
name: commit-message
description: >-
  Write or improve git commit messages for this repository. Use when the user
  asks for a commit message, is creating a commit, amending a message,
  reviewing whether a commit message is good enough, or drafting an example
  message for a pull request contributor.
license: Apache-2.0
metadata:
  audience: contributors
  style: kubernetes-community
---

# Commit messages

Follow the Kubernetes
[Commit Message Guidelines](https://www.kubernetes.dev/docs/guide/pull-requests/#commit-message-guidelines).
The rules below summarize that guide for agents; prefer this skill text over
fetching the web page.

Commits are the permanent record of a change. `git log` must explain **why**
the change is needed (the problem) and **what** you did about it (the
solution), in that order. AI agents may draft messages; the text must still
match the diff — do not invent intent or files.

## When to use

1. Read the change (`git diff`, staged files, or PR commits). Do not guess.
2. Draft a message using the format below.
3. If asked for a *contributor example*, keep it copy-paste ready.
4. Do not run `git commit` unless the user explicitly asked to commit.

## Format

```text
area: Imperative summary in ~50 characters (max 72)

Explain the problem first. Who is hurt, when it shows up, why the current
behavior is wrong. User-visible impact beats implementation trivia.

Then describe what the change does about it in plain English — enough that
a reviewer can verify the diff matches the intent. Mention important
non-obvious trade-offs or limits (what this does *not* change).
```

### Subject

From the Kubernetes guide:

- Prefer ≤**50** characters; never exceed **72**.
- **Imperative mood** (a command): “Fix …”, “Add …”, “Remove …”.
  Not “Fixed”, “Adding”, “This commit”, or “I …”.
  Test: “If applied, this commit will <subject>”.
- **No trailing period**.
- Capitalize the first word **unless** the subject starts with a lowercase
  area/kind prefix (`site:`, `etcd:`, …).
- Optional `area:` / `kind:` prefix for context, e.g. `site:`, `addons:`,
  `cmd/minikube:`, `pkg/drivers:`, `deploy/iso:`.
- Must make sense alone in `git log --oneline`.

### Body

- One blank line after the subject.
- Wrap at **72** characters.
- Explain the **problem**, then the **solution** (why, then what). Self-contained:
  readers should not need the PR thread.
- One logical change per commit. If the body needs an unrelated “also…”, split.
- Do not narrate the diff line-by-line; describe behavior and rationale.
- Quantify performance claims; call out trade-offs when relevant.
- You may link related discussion with a plain URL. Do **not** use GitHub
  closing keywords with an issue number (see below).

### Do not use GitHub keywords or @mentions in commits

Kubernetes applies `do-not-merge/invalid-commit-message` when a commit message
uses [GitHub closing keywords](https://docs.github.com/en/get-started/writing-on-github/working-with-advanced-formatting/using-keywords-in-issues-and-pull-requests) followed by an `#` issue number.

**Do not use in commit messages** (with an issue reference):

`close`, `closes`, `closed`, `fix`, `fixes`, `fixed`, `resolve`, `resolves`,
`resolved`

Also avoid `@mentions` in commit messages (they notify on every PR update).

Put issue linkage in the **pull request description** instead, e.g. `Fixes #123`.

Do not add `Signed-off-by:` unless a reviewer or process asks for it (minikube
uses the Kubernetes CLA).

## Quality bar

- [ ] `git log --oneline` explains the change at a glance
- [ ] Body states the **problem** / why the change is needed
- [ ] Body states the **solution** / what the change does
- [ ] Message matches the actual diff
- [ ] No GitHub closing keywords + `#N`, and no `@mentions`
- [ ] Residual limitations called out when relevant

Reject or rewrite: title-only messages with no why, empty bodies under noisy
prefixes (`fix(ui): …`), or “Address review comments”.

## Examples

**Too weak**

```text
bugfix(ui):copy button and code block overlap
```

**Better**

```text
site: fix copy button overlapping code in docs

On narrow viewports, long lines in Prism code blocks scroll under the
floating Copy button. Prism styles that button with a translucent
background, so the code shows through the muted "Copy" label and both
become hard to read.

Override the button background in the project SCSS to Prism's solid
fallback color (#f5f2f0) so the label stays readable when text sits
underneath. Use the same selector as Prism's toolbar CSS so the override
targets the copy control clearly.

This keeps Prism's toolbar behavior (show on hover, Copied! label); it
only disables the translucent fill.
```

More examples: [examples.md](examples.md)

## For reviewers

When a PR commit message is weak:

1. Say what is wrong (title-only, missing what/why, keyword+#N, mismatches diff).
2. Paste an example the author can adapt.
3. Point them at this skill and the Kubernetes guide linked above.
4. AI drafting is fine if the result meets the quality bar.
