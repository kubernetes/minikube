# mkcmp

mkcmp (“minikube compare”) is a binary to compare the performance of two minikbue binaries.
It takes in two minikube binaries, runs `minikube start` 3 times with each, and then outputs a comparison table in Markdown.

mkcmp takes in references to minikube binaries in two ways:
1. Direct path to a minikube binary
1. Reference to a PR number via `pr://<PR number>`, which will use the binary built at that PR

Sample usage:

```shell
make out/mkcmp
# Compare local minikube binary with binary built at PR 400
./out/mkcmp ./out/minikube pr://400
```

mkcmp is primarily used for our pr-bot, which comments mkcmp output on valid PRs [example](https://github.com/kubernetes/minikube/pull/10430#issuecomment-776311409).
To make changes to the pr-bot output, submitting a PR to change mkcmp code should be sufficient.

Note: STDOUT from mkcmp is *exactly* what is commented on github, so we want it to be in Markdown.

