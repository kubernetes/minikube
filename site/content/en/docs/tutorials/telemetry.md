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

To collect trace data with minikube and the Stackdriver exporter, run:

```shell
MINIKUBE_GCP_PROJECT_ID=<project ID> minikube start --output json --trace gcp
```

## Contributing

There are many exporters available via [OpenTelemetry community contributions](https://github.com/open-telemetry/opentelemetry-collector-contrib).

If you would like to see additional exporters, please create an [issue](https://github.com/kubernetes/minikube/issues) or refer to our [contribution][https://minikube.sigs.k8s.io/docs/contrib/] guidelines and submit a pull request. Thank you!
