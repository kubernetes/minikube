#!/bin/bash
# shellcheck disable=SC1091,SC2034

# Configures environment for Bicep and Packer templates with required inputs
# for pipelines to create SIG, build VM image, create VM.
#
# Create env.local.sh with user-specific overrides of the defaults.

set -o allexport

### Azure environment defaults
# https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration#cli-configuration-values-and-environment-variables
[ -z "$AZURE_CORE_ONLY_SHOW_ERRORS" ] && AZURE_CORE_ONLY_SHOW_ERRORS=true # Disable warnings about preview features
[ -z "$AZURE_CORE_OUTPUT" ] && AZURE_CORE_OUTPUT=jsonc
# Not necessary as we expect pre-existing resource group, then its location is referenced as default.
[ -z "$AZURE_DEFAULTS_LOCATION" ] && AZURE_DEFAULTS_LOCATION="$(az config get defaults.location --only-show-errors --query value -o tsv)"
[ -z "$AZURE_DEFAULTS_LOCATION" ] && AZURE_DEFAULTS_LOCATION=southcentralus
# AZURE_DEFAULTS_GROUP not used to avoid implicit target group. Makefile commands are given explicit --resource-group name.

### Minikube environment defaults
[ -z "$MINIKUBE_AZ_SUBSCRIPTION_ID" ] && MINIKUBE_AZ_SUBSCRIPTION_ID="$(az account show --query 'id' --output tsv)"
[ -z "$MINIKUBE_AZ_RESOURCE_GROUP" ] && MINIKUBE_AZ_RESOURCE_GROUP="SIG-CLUSTER-LIFECYCLE-MINIKUBE"
[ -z "$MINIKUBE_AZ_SIG_NAME" ] && MINIKUBE_AZ_SIG_NAME="minikube"
[ -z "$MINIKUBE_AZ_IMAGE_NAME" ] && MINIKUBE_AZ_IMAGE_NAME="minikube-ci-windows-11"
[ -z "$MINIKUBE_AZ_IMAGE_VERSION" ] && MINIKUBE_AZ_IMAGE_VERSION="1.0.0"
[ -z "$MINIKUBE_AZ_VM_ADMIN_USERNAME" ] && MINIKUBE_AZ_VM_ADMIN_USERNAME="minikubeadmin"
[ -z "$MINIKUBE_AZ_VM_ADMIN_PASSWORD" ] && MINIKUBE_AZ_VM_ADMIN_PASSWORD="" # Must be overriden via env.local.sh or export
[ -z "$MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY" ] && MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY="" # Must be overriden via env.local.sh or export

### Load user-specific environment overrides, if present
[ -f ./env.local.sh ] && source ./env.local.sh

set +o allexport
