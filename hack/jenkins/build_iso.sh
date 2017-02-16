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

# This script can take the following env variables
# ARGS: args to pass into the make rule
# 	ISO_BUCKET = the bucket location to upload the ISO (e.g. minikube-builds/PR_NUMBER)
# 	ISO_VERSION = the suffix for the iso (i.e. minikube-$(ISO_VERSION).iso)

set -e
${ARGS} make release-iso
