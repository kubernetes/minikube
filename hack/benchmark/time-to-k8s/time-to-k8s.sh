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

install_kind() {
        curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.11.0/kind-linux-amd64
	chmod +x ./kind
	sudo mv ./kind /usr/local
}

install_k3d() {
	curl -s https://raw.githubusercontent.com/rancher/k3d/main/install.sh | bash
}

install_minikube() {
	make
	sudo install ./out/minikube /usr/local/bin/minikube
}

run_benchmark() {
	( cd ./hack/benchmark/time-to-k8s/time-to-k8s/ &&
		git submodule update --init &&
		go run . --config local-kubernetes.yaml --iterations 5 --output output.csv )
}

generate_chart() {
	go run ./hack/benchmark/time-to-k8s/chart.go --csv ./hack/benchmark/time-to-k8s/time-to-k8s/output.csv --output ./site/static/images/benchmarks/timeToK8s/"$1".png
}

create_page() {
	printf -- "---\ntitle: \"%s Benchmark\"\nlinkTitle: \"%s Benchmark\"\nweight: 1\n---\n\n![time-to-k8s](/images/benchmarks/timeToK8s/%s.png)\n" "$1" "$1" "$1" > ./site/content/en/docs/benchmarks/timeToK8s/"$1".md
}

commit_chart() {
	git add ./site/static/images/benchmarks/timeToK8s/"$1".png ./site/content/en/docs/benchmarks/timeToK8s/"$1".md
	git commit -m 'update time-to-k8s chart'
}

install_kind
install_k3d
install_minikube
VERSION=$(minikube version --short)

run_benchmark
generate_chart "$VERSION"
create_page "$VERSION"
commit_chart "$VERSION"
