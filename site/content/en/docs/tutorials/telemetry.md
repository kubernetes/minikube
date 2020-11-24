---
title: "Telemetry"
linkTitle: "telemetry"
weight: 1
date: 2020-11-24
---

## Overview

minikube provides telemetry suppport via [OpenTelemetry tracing](https://opentelemetry.io/about/) to collect trace data for `minikube start`.

Currently, minikube supports the following exporters for tracing data:
- [Stackdriver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/master/exporter/stackdriverexporter)

Other exporters can be contributed by users as needed.

To collect trace data with minikube and the Stackdriver exporter, run:

```
MINIKUBE_GCP_PROJECT_ID minikube start --output json --trace gcp
```

