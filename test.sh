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

TESTSUITE="${TESTSUITE:-all}" # if env variable not set run all the tests
exitcode=0

if [[ "$TESTSUITE" = "lint" ]] || [[ "$TESTSUITE" = "all" ]]
then
    echo "= make lint ============================================================="
    make -s lint-ci && echo ok || ((exitcode += 4))
    echo "= go mod ================================================================"
    go mod download 2>&1 | grep -v "go: finding" || true
    go mod tidy -v && echo ok || ((exitcode += 2))
fi



if [[ "$TESTSUITE" = "boilerplate" ]] || [[ "$TESTSUITE" = "all" ]]
then
    echo "= boilerplate ==========================================================="
    # readonly GO=$(type -P go || echo docker run --rm -it -v $(pwd):/minikube -w /minikube go go)
    readonly GO=$(type -P go)
    readonly BDIR="./hack/boilerplate"
    missing="$($GO run ${BDIR}/boilerplate.go --rootdir . --boilerplate-dir ${BDIR} | egrep -v 'deploy.sh|assets.go|translations.go|/site/themes/|/site/node_modules|\./out|/hugo/' || true)"
    if [[ -n "${missing}" ]]; then
        echo "boilerplate missing: $missing"
        echo "consider running: ${BDIR}/fix.sh"
        ((exitcode += 8))
    else
        echo "ok"
    fi
fi


if [[ "$TESTSUITE" = "unittest" ]] || [[ "$TESTSUITE" = "all" ]]
then 
    echo "= schema_check =========================================================="
    go run deploy/minikube/schema_check.go >/dev/null && echo ok || ((exitcode += 16))

    echo "= go test ==============================================================="
    cov_tmp="$(mktemp)"
    readonly COVERAGE_PATH=./out/coverage.txt
    echo "mode: count" >"${COVERAGE_PATH}"
    pkgs=$(go list -f '{{ if .TestGoFiles }}{{.ImportPath}}{{end}}' ./cmd/... ./pkg/... | xargs)
    go test \
        -tags "container_image_ostree_stub containers_image_openpgp" \
        -covermode=count \
        -coverprofile="${cov_tmp}" \
        ${pkgs} && echo ok || ((exitcode += 32))
    tail -n +2 "${cov_tmp}" >>"${COVERAGE_PATH}"
fi

exit "${exitcode}"
