#!/bin/bash

# Copyright 2016 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# This script runs the integration tests on an OSX machine for the xhyve Driver

# The script expects the following env variables:
# K8S_VERSION: GIT_COMMIT from upstream build.
# BUCKET: The GCP bucket the build files should be uploaded to.

gsutil mv gs://${BUCKET}/k8sReleases/${K8S_VERSION}+testing/localkube-linux-amd64 gs://${BUCKET}/k8sReleases/${K8S_VERSION}/localkube-linux-amd64
gsutil mv gs://${BUCKET}/k8sReleases/${K8S_VERSION}+testing/localkube-linux-amd64.sha256 gs://${BUCKET}/k8sReleases/${K8S_VERSION}/localkube-linux-amd64.sha256
