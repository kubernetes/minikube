# Periodically tell user about minikube features/tips and tricks

* First proposed: 2021-06-18
* Authors: Peixuan Ding (@dinever)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

minikube has lots of great features. We want to proactively remind users that
those features are available.

To achieve this, we can have a tips feature that randomly shows a tip
from a curated list whenever the user starts a new minikube profile.

For example:

![Screenshot from 2021-06-18 00-58-02](https://user-images.githubusercontent.com/1311594/122508665-53bd6380-cfd0-11eb-9e99-a6c5935514d5.png)

## Goals

* Store a list of tips in a static file
* Show a random minikube usage tip each time a user starts a minikube profile
* Have the tips synced to the Hugo docs website to make those available through docs
* Allow user to disable the Tips feature with minikube config

## Non-Goals

* Modify any existing functionalities or docs

## Design Details

First, we need a static file to store all the tips, we can have a YAML file at [pkg/generate/tips/tips.yaml](https://github.com/kubernetes/minikube/tree/master/pkg/generate):

```YAML
tips:
  - |
    You can specify any Kubernetes version you want. For example:

    ```
    minikube start --kubernetes-version=v1.19.0
    ```
  - |
    You can use minikube's built-in kubectl. For example:

    ```
    minikube kubectl -- get pods
    ```
  - |
    minikube has the built-in Kubernetes Dashboard UI. To access it:

    ```
    minikube dashboard
    ```
```

Use `goembed` to embed this file into the minikube binary.

The current `out.Boxed` has a hard-coded style (red). I propose to add another `out.BoxedWithConfig` method to allow
output with customized style:

```go
// BoxedWithConfig writes a templated message in a box with customized style config to stdout
func BoxedWithConfig(cfg box.Config, st style.Enum, title string, format string, a ...V) {
}
```

Whenever minikube successfully starts, we randomly choose a tip.

Before printing it out, we need to do some regex replacement to strip the markdown syntax
for better view experience in Terminal:

From this:

``````markdown
You can specify any Kubernetes version you want. For example:

```
minikube start --kubernetes-version=v1.19.0
```
``````

To this:

```markdown
You can specify any Kubernetes version you want. For example:

minikube start --kubernetes-version=v1.19.0
```

Then we can print out the tip:


```go
boxCfg := out.BoxConfig{
	Config: box.Config{
		Py: 1,
		Px: 5,
		TitlePos: "Top",
		Type: "Round",
		Color: tipBoxColor,
	},
	Title: tipTitle,
	Icon: style.Tip,
}

out.BoxedWithConfig(boxCfg, tips.Tips[chosen] + "\n\n" + tipSuffix)
```

![Screenshot from 2021-06-18 00-58-02](https://user-images.githubusercontent.com/1311594/122508665-53bd6380-cfd0-11eb-9e99-a6c5935514d5.png)

User can choose to disable this through `minikube config set disable-tips true`

We will have `make generate-docs` generating the docs site based on this YAML file as well.

We can have a `Nice to know` sub-page under `FAQ`?

![Screenshot from 2021-06-18 01-00-30](https://user-images.githubusercontent.com/1311594/122508827-a139d080-cfd0-11eb-98bb-f7c3c1c604c2.png)


### About the tip collection

I plan to start with the command lines and cover almost all CLI usages of minikube.

That includes but not limited to:
- addons
- cached images
- command line completion
- config
- file copy
- dashboard
- delete minikube cluster
- configure minikube's docker/podman env
- image build / load / ls / rm
- ip
- logging
- kubectl
- mount file directory
- multi-node
- pause/unpause to save resource
- multi-profile
- surface URL to a k8s service
- ssh into minikube
- status
- tunnel to connect to LB
- update-check to check versions
- update-context

### Implementation

I plan to open at least 4 PRs:

1. `out.Boxed` with custom style
2. random `tips` display with ability to disable through config, with an initial set of about 10 tips
3. `make generate-docs` to sync tips to docs
4. Add more tips

## Alternatives Considered

1. Is there a more preferred file format to YAML?

2. Maybe we just want to sync the tips to the `FAQ` page list instead of creating a new page?

3. Instead of the file format I proposed, maybe add a `question` field?

    ```yaml
    tips:
      - question: How to specify a different Kubernetes version?
        answer: |
          You can specify any Kubernetes version you want. For example:

          ```
          minikube start --kubernetes-version=v1.19.0
          ```
      - question: Do I have to install `kubectl` myself?
        answer: |
          You can use minikube's built-in kubectl. For example:

          ```
          minikube kubectl -- get pods
          ```
      - question: How do I access the Kubernetes Dashboard UI?
        answer: |   
          minikube has the built-in Kubernetes Dashboard UI. To access it:

          ```
          minikube dashboard
          ```
    ```

   On the docs side we can show both questions and answers. On the CLI side
   we can either show both questions and answers, or just show the answers
   to make it more compact.

   ![Screenshot from 2021-06-18 01-25-54](https://user-images.githubusercontent.com/1311594/122510785-2c689580-cfd4-11eb-9fd0-0a0ff344e3cc.png)

