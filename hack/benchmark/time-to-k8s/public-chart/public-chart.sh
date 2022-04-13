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

DRIVER="$1"
# container-runtime (docker or containerd)
RUNTIME="$2"
BUCKET="s3://time-to-k8s"

install_minikube() {
        make
        sudo install ./out/minikube /usr/local/bin/minikube
}

run_benchmark() {
        ( cd ./hack/benchmark/time-to-k8s/time-to-k8s-repo/ &&
                git submodule update --init &&
                go run . --config "../public-chart/$DRIVER-$RUNTIME-benchmark.yaml" --iterations 10 --output ./output.csv )
}

generate_chart() {
        go run ./hack/benchmark/time-to-k8s/public-chart/generate-chart.go --csv ./hack/benchmark/time-to-k8s/time-to-k8s-repo/output.csv --daily-chart ./daily-chart.png --weekly-chart ./weekly-chart.png --past-runs ./runs.json
}

copy() {
	aws s3 cp "$1" "$2"
}

cleanup() {
	rm ./runs.json
	rm ./hack/benchmark/time-to-k8s/time-to-k8s-repo/output.csv
	rm ./daily-chart.png
	rm ./weekly-chart.png
}

copy "$BUCKET/$DRIVER-$RUNTIME-runs.json" ./runs.json

install_minikube

run_benchmark
generate_chart

copy ./runs.json "$BUCKET/$DRIVER-$RUNTIME-runs.json"
copy ./runs.json "$BUCKET/$(date +'%Y-%m-%d')-$DRIVER-$RUNTIME.json"
copy ./daily-chart.png "$BUCKET/$DRIVER-$RUNTIME-chart.png"
copy ./weekly-chart.png "$BUCKET/$DRIVER-$RUNTIME-weekly-chart.png"

cleanup
