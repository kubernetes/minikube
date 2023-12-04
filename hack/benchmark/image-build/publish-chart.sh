#!/bin/bash

# Copyright 2023 The Kubernetes Authors All rights reserved.
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

set -x

BUCKET="s3://time-to-k8s/image-benchmark"

install_minikube() {
        make
        sudo install ./out/minikube /usr/local/bin/minikube
}

run_benchmark() {
        ( cd ./hack/benchmark/image-build/minikube-image-benchmark &&
                git submodule update --init &&
                make                        &&
                ./out/benchmark --runs=4 --memory="1800m" --images="buildpacksFewLargeFiles" --iters="iterative" --bench-methods="image load docker,image build docker,docker-env docker,registry docker,image load containerd,image build containerd,registry containerd,docker-env containerd")
}

generate_chart() {
        go run ./hack/benchmark/image-build/generate-chart.go --csv hack/benchmark/image-build/minikube-image-benchmark/out/results.csv  --past-runs record.json
}

copy() {
	aws s3 cp "$1" "$2"
}

cleanup() {
	rm ./Iterative_buildpacksFewLargeFiles_containerd_chart.png
        rm ./Iterative_buildpacksFewLargeFiles_docker_chart.png
	rm hack/benchmark/image-build/minikube-image-benchmark/out/results.csv
}


install_minikube
copy "$BUCKET/record.json" ./record.json
set -e

run_benchmark
generate_chart

copy ./record.json "$BUCKET/record.json"
copy ./Iterative_buildpacksFewLargeFiles_containerd_chart.png "$BUCKET/Iterative_buildpacksFewLargeFiles_containerd_chart.png"
copy ./Iterative_buildpacksFewLargeFiles_docker_chart.png "$BUCKET/Iterative_buildpacksFewLargeFiles_docker_chart.png"
cleanup
