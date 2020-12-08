This script runs `minikube start` in a loop and measures how long it takes.
It exports this data to Stackdriver via the OpenTelemetry API.

To run this script, run:

```
MINIKUBE_GCP_PROJECT_ID=<GCP Project ID> go run hack/metrics/*.go
```

This script is used to track minikube performance and prevent regressions.

_Note: this script will export data to both Cloud Monitoring and Cloud Trace in the provided GCP project_
