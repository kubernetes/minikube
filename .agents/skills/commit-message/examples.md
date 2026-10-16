# Commit message examples

## Bug fix

```text
kvm: avoid nil deref when VM IP is not yet assigned

During early start, GetIP can run before the lease exists. Callers treated
a missing address as a fatal error and crashed the start path.

Return ErrNotFound when the lease is absent and let the existing retry
loop wait for the address. Log at V(1) so quiet starts stay quiet.
```

## Docs

```text
docs: document WSL2 service access for the Docker driver

WSL2 users hit connection refused when curling NodePort services from
Windows and assumed minikube was broken. The missing piece is how Windows
reaches the WSL2 VM IP.

Add a short handbook section with the working patterns and known
limitations. No code changes.
```

## Refactor (still explain why)

```text
pkg/minikube/command: split SSH runner deadline handling

Deadline logic was inlined in three call paths and drifted: one path
ignored the parent context on retry. Tests could not cover the edge cases
without copying the same stubs.

Extract waitWithDeadline and use it from all three paths. Behavior should
be unchanged except retries now honor cancellation consistently.
```

## Performance (quantify)

```text
start: stream preload tarballs instead of buffering whole file

Buffering the full preload in memory OOMs mid-start on large images over
slow links.

Stream with a capped 32MiB buffer. Peak RSS returns to the pre-regression
range; LAN cold-start time is unchanged within noise.
```

## Rewrite these

| Weak | Why |
|------|-----|
| `fix bug` | No area, problem, or solution |
| `Updated styles` | No why |
| `fix(ui): improve UX` | Prefix without meaning; empty body |
| `Address review comments` | Useless months later in `git log` |
| `WIP` | Not mergeable as a final message |
| `Fixes: #123` / `fixes #123` in the commit | Triggers `do-not-merge/invalid-commit-message`; put `Fixes #123` in the PR description |
| `@someone` in the commit | Mentions notify on every PR update; avoid |
