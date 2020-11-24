This script runs `minikube start` in a loop and measures how long it takes.
It exports this data to Stackdriver via the OpenTelemetry API.

This script is used to track minikube performance and prevent regressions.
