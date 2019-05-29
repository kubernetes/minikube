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

set -eu -o pipefail

exitcode=0

echo "= go mod ================================================================"
go mod download 2>&1 | grep -v "go: finding" || true
go mod tidy -v && echo ok || ((exitcode+=2))

echo "= make lint ============================================================="
make -s lint && echo ok || ((exitcode+=4))

echo "= boilerplate ==========================================================="
readonly PYTHON=$(type -P python || echo docker run --rm -it -v $(pwd):/minikube -w /minikube python python)
readonly BDIR="./hack/boilerplate"
missing="$($PYTHON ${BDIR}/boilerplate.py --rootdir . --boilerplate-dir ${BDIR} | grep -v \/assets.go || true)"
if [[ -n "${missing}" ]]; then
    echo "boilerplate missing: $missing"
    echo "consider running: ${BDIR}/fix.sh"
    ((exitcode+=4))
else
    echo "ok"
fi

echo "= schema_check =========================================================="
go run deploy/minikube/schema_check.go >/dev/null && echo ok || ((exitcode+=8))

echo "= go test ==============================================================="
cov_tmp="$(mktemp)"
readonly COVERAGE_PATH=./out/coverage.txt
echo "mode: count" > "${COVERAGE_PATH}"
pkgs=$(go list -f '{{ if .TestGoFiles }}{{.ImportPath}}{{end}}' ./cmd/... ./pkg/... | xargs)
go test \
    -tags "container_image_ostree_stub containers_image_openpgp" \
    -covermode=count \
    -coverprofile="${cov_tmp}" \
    ${pkgs} && echo ok || ((exitcode+=16))
tail -n +2 "${cov_tmp}" >> "${COVERAGE_PATH}"

exit "${exitcode}"
