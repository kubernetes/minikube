#!/bin/bash

# Copyright 2021 The Kubernetes Authors All rights reserved.
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

set -e

# container-runtime (docker or containerd)
RUNTIME="$1"

install_minikube() {
        make
        sudo install ./out/minikube /usr/local/bin/minikube
}

run_benchmark() {
        ( cd ./hack/benchmark/time-to-k8s/time-to-k8s-repo/ &&
                git submodule update --init &&
                go run . --config "../public-chart/$RUNTIME-benchmark.yaml" --iterations 10 --output ./output.csv )
}

generate_chart() {
        go run ./hack/benchmark/time-to-k8s/public-chart/generate-chart.go --csv ./hack/benchmark/time-to-k8s/time-to-k8s-repo/output.csv --output ./chart.png --past-runs ./runs.json
}

cleanup() {
	rm ./runs.json
	rm ./hack/benchmark/time-to-k8s/time-to-k8s-repo/output.csv
	rm ./chart.png
}

gsutil -m cp "gs://minikube-time-to-k8s/$RUNTIME-runs.json" ./runs.json

install_minikube

run_benchmark
generate_chart

gsutil -m cp ./runs.json "gs://minikube-time-to-k8s/$RUNTIME-runs.json"
gsutil -m cp ./runs.json "gs://minikube-time-to-k8s/$(date +'%Y-%m-%d')-$RUNTIME.json"
gsutil -m cp ./chart.png "gs://minikube-time-to-k8s/$RUNTIME-chart.png"

cleanup
