#!/bin/bash
set -e
gsutil -m cp -r gs://minikube-builds/${MINIKUBE_LOCATION}/common.sh .
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/print-debug-info.sh . || true
gsutil cp gs://minikube-builds/${MINIKUBE_LOCATION}/linux_integration_tests_kvm.sh .

sudo gsutil cp gs://minikube-builds/kvm-driver/docker-machine-driver-kvm /usr/local/bin/docker-machine-driver-kvm
sudo chmod +x /usr/local/bin/docker-machine-driver-kvm

bash -x linux_integration_tests_kvm.sh
