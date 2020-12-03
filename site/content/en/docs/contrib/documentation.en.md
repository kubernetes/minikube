---
linkTitle: "Documentation"
title: "Contributing to minikube documentation"
date: 2019-07-31
weight: 2
aliases:
  - /docs/contribution-guidelines/
---

minikube's documentation is in [Markdown](https://www.markdownguide.org/cheat-sheet/), and generated using the following tools:

* [Hugo](https://gohugo.io)
* [Docsy](https://www.docsy.dev)

In production, the minikube website is served using [Netlify](https://netlify.com/)

## Small or cosmetic contributions

Use Github's repositories and markdown editor as described by [Kubernetes's general guideline for documentation contributing](https://kubernetes.io/docs/contribute/start/#submit-a-pull-request)

## Local documentation website

To serve documentation pages locally, clone the `minikube` repository and run:

```shell
make site
```

Notes :

* On GNU/Linux, golang package shipped with the distribution may not be recent enough. Use the latest version.
* On Windows, our site currently causes Hugo to `panic`.

## Lint

We recommend installing [markdownlint](https://github.com/markdownlint/markdownlint) to find issues with your markdown file. Once installed, you can use this handy target:

```shell
make mdlint
```

## Style Guidelines

We follow the [Kubernetes Documentation Style Guide](https://kubernetes.io/docs/contribute/style/style-guide/)

## Linking between documents

For compile-time checking of links, use one of the following forms to link between documentation pages:


```go-html-template
{{</* ref "document.md" */>}}
{{</* ref "#anchor" */>}}
{{</* ref "document.md#anchor" */>}}
{{</* ref "/blog/my-post" */>}}
{{</* ref "/blog/my-post.md" */>}}
{{</* relref "document.md" */>}}
{{</* relref "#anchor" */>}}
{{</* relref "document.md#anchor" */>}}
```

For more information, please see [Hugo: Links and Cross References](https://gohugo.io/content-management/cross-references/)

## Pull Request Previews

When reviewing documentation PR's, look for the test that says:

**✓ deploy/netlify** Deploy preview ready! *Details*

The `Details` link will point to a site preview URL in the form of:

<https://deploy-preview-PR#--kubernetes-sigs-minikube.netlify.com>
