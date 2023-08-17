# Allow importing of profile configurations

* First proposed: 2023-08-17
* Authors: Jeremias Weber (@jelemux)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

Often, projects want to store configurations for development clusters in version control.
Currently, this is not very easy to do. Minikube aims to be user-friendly and developer-focused.
As configuration-as-code is common practice nowadays, this use-case should be supported.

## Goals

*   Allow the import of profile configurations in the `minikube start` command

## Non-Goals

*   Merging the imported configuration with an already existing one.

## Design Details

An `import-profile` flag is added to the `minikube start` command.
It can be used to pass a path to a profile configuration file that should be imported.

Example:
```shell
minikube start --import-profile ./path/to/profile/config.json
```

If a profile with the given name already exists, the command will fail.
This behaviour can be overwritten with the `force` flag.

Configuration in flags must overwrite any imported config values.

Unit tests will verify this behaviour and ensure that any errors will be handled.

## Alternatives Considered

There are not many alternatives besides manually creating the profile in the `.minikube` directory.
Of course, the configuration could also be persisted as a script, but I think that a declarative configuration is much better.
