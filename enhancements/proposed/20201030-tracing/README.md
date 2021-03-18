# Tracing minikube

* First proposed: Oct 30 2020
* Authors: Priya Wadhwa (priyawadhwa@)

## Reviewer Priorities

Please review this proposal with the following priorities:

*   Does this fit with minikube's [principles](https://minikube.sigs.k8s.io/docs/concepts/principles/)?
*   Are there other approaches to consider?
*   Could the implementation be made simpler?
*   Are there usability, reliability, or technical debt concerns?

Please leave the above text in your proposal as instructions to the reader.

## Summary

This proposal covers using the [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) API to provide tracing data for minikube.
This data would be useful for maintainers to identify areas for performance improvements.
This data would also be used to create a dashboard of current performance and would allow us to catch performance regressions more quickly.

## Goals

*   Trace data is can be collected and exported for `minikube start`
*   `minikube start` can either create a new Trace or can read from a file and append data to an existing Trace
*   It is easy for users to add their own exporters if they wish to export data to their own service
*   We are able to create dashboards around `minikube start` performance that will alert maintainers if regressions happen

## Non-Goals

*   Collecting trace data for the minikube cluster while it is running

## Design Details

There are two pieces to the design: collecting the data and exporting the data.

### Collecting Data
Luckily, we already have a lot of this infrastructure set up for JSON output.
We know when a new substep of `minikube start` has started, because we call it explicitly via `register.SetStep`.
We also know that substep has ended when a new substep begins.

We can start new spans whenever `register.SetStep` is called, and thus collect tracing data.

### Exporting Data
OpenTelemetry supports a variety of [user-contributed exporters](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/master/instrumentation).
It would be a lot of work to implement all of them ourselves.

Instead, I propose writing a simple `GetExporter` function that would return whatever exporter is requested via a `--trace` flag.

So, something like this would tell minikube to use the `stackdriver` exporter:

```
minikube start --trace=stackdriver
```

Users can then contribute to minikube if they need to use an exporter that isn't currently provided.

Exporters also will require additional information to make sure data is sent to the correct place.
This could include things like, but not limited to:
* project ID
* zone

Since it could get messy passing in these things as flags to `minikube start`, I propose that these values are set via environment variable.
All environment variables will be of the form:

```
MINIKUBE_TRACE_PROJECT_ID
```
and the user-contributed code is responsible for parsing the environment variables correctly and returning the exporter.

### Testing Plan

I will set up a dashboard and alerting system in the minikube GCP project.
If we are collecting data at a consistent rate, and the dashboard is populated, we will know that this has worked.


## Alternatives Considered

### Building a Wrapper Binary
A wrapper binary could run `minikube start --output json` and collect the same data, and then export it to whatever service we need.

A large advantage of this is that the minikube code doesn't have to be changed at all for this to work.

However, I decided against this in case other tools that consume minikube or users want to collect this data as well -- it is much easier to pass in a flag to minikube than to download another binary.
