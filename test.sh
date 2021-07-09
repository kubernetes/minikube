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
CI="${CI:-false}" # if env variable not set don't run CI tests
exitcode=0

if [[ "$TESTSUITE" = "lint" ]] || [[ "$TESTSUITE" = "all" ]] || [[ "$TESTSUITE" = "lintall" ]]
then
    echo "= make lint ============================================================="
    make -s lint-ci && echo ok || ((exitcode += 4))
    echo "= go mod ================================================================"
    go mod download 2>&1 | grep -v "go: finding" || true
    if [[ "$CI" = "true" ]]
    then
        go mod tidy -v && git diff --quiet go.* && echo ok || (((exitcode += 2)) && echo ERROR: Please run go mod tidy)
        echo "= generate docs ========================================================="
        make generate-docs > /dev/null 2>&1 && git diff --quiet site && echo ok || (((exitcode += 3)) && echo ERROR: Please run make generate-docs)
    else
        go mod tidy -v && echo ok || ((exitcode += 2))
    fi
fi



if [[ "$TESTSUITE" = "boilerplate" ]] || [[ "$TESTSUITE" = "all" ]] || [[ "$TESTSUITE" = "lintall" ]]
then
    echo "= boilerplate ==========================================================="
    readonly ROOT_DIR=$(pwd)
    readonly BDIR="${ROOT_DIR}/hack/boilerplate"
    pushd . >/dev/null
    cd ${BDIR}
    missing="$(go run boilerplate.go -rootdir ${ROOT_DIR} -boilerplate-dir ${BDIR} | egrep -v '/assets.go|/translations.go|/site/themes/|/site/node_modules|\./out|/hugo/' || true)"
    if [[ -n "${missing}" ]]; then
        echo "boilerplate missing: $missing"
        echo "consider running: ${BDIR}/fix.sh"
        ((exitcode += 8))
    else
        echo "ok"
    fi
    popd >/dev/null
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
        -ldflags="$MINIKUBE_LDFLAGS" \
        -tags "container_image_ostree_stub containers_image_openpgp" \
        -covermode=count \
        -coverprofile="${cov_tmp}" \
        ${pkgs} \
        && echo ok || ((exitcode += 32))
    tail -n +2 "${cov_tmp}" >>"${COVERAGE_PATH}"
    rm ${cov_tmp}
fi

exit "${exitcode}"
