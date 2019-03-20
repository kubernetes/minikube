# Adding a New Dependency

Minikube uses `dep` to manage vendored dependencies.

See the `dep` [documentation](https://golang.github.io/dep/docs/introduction.html) for installation and usage instructions.

If you are introducing a large dependency change, please commit the vendor/ directory changes separately.
This makes review easier in GitHub.
