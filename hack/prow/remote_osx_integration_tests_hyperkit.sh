#!/bin/bash
set -e
gsutil -m cp -r gs://minikube-builds/${MINIKUBE_LOCATION}/common.sh .
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/print-debug-info.sh . || true
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/osx_integration_tests_hyperkit.sh .
bash -x osx_integration_tests_hyperkit.sh

