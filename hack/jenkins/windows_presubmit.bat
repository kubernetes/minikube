:: Copyright 2019 The Kubernetes Authors All rights reserved.
::
:: Licensed under the Apache License, Version 2.0 (the "License");
:: you may not use this file except in compliance with the License.
:: You may obtain a copy of the License at
::
::     http://www.apache.org/licenses/LICENSE-2.0
::
:: Unless required by applicable law or agreed to in writing, software
:: distributed under the License is distributed on an "AS IS" BASIS,
:: WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
:: See the License for the specific language governing permissions and
:: limitations under the License.

set MINIKUBE_LOCATION=%KOKORO_GITHUB_PULL_REQUEST_NUMBER%
set COMMIT=%KOKORO_GITHUB_PULL_REQUEST_COMMIT%

cd github/minikube
mkdir -p out
gsutil -m cp gs://minikube-builds/%MINIKUBE_LOCATION%/minikube-windows-amd64.exe out/
gsutil -m cp gs://minikube-builds/%MINIKUBE_LOCATION%/e2e-windows-amd64.exe out/
gsutil -m cp -r gs://minikube-builds/%MINIKUBE_LOCATION%/testdata .

out/minikube-windows-amd64.exe delete --all

out/e2e-windows-amd64.exe -minikube-start-args="--driver=docker" -binary=out/minikube-windows-amd64.exe -test.v -test.timeout=65m -test.run=TestFunctional
