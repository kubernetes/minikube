#!/bin/bash
set -e
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/common.sh .
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/print-debug-info.sh . || true
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/osx_integration_tests_virtualbox.sh .
bash -x osx_integration_tests_virtualbox.sh
