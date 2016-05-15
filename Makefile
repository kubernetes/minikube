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

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out

ifeq ($(IN_DOCKER),1)
	GOPATH := /go
else
	GOPATH := $(shell pwd)/.gopath
endif

REPOPATH ?= k8s.io/minikube
K8S_VENDOR_PATH := $(REPOPATH)/vendor/k8s.io/kubernetes
export GO15VENDOREXPERIMENT=1

# Update this to match the vendored version of k8s in godeps.json
K8S_VERSION := v1.3.0-alpha.3-838+ba170aa191f8c7
# Set the version information in kubernetes.
LD_FLAGS = "-s -w -X $(K8S_VENDOR_PATH)/pkg/version.gitVersion=$(K8S_VERSION) -X $(K8S_VENDOR_PATH)/pkg/version.gitCommit=$(shell git rev-parse HEAD)"

clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)
	rm pkg/minikube/cluster/localkubecontents.go

MKGOPATH := mkdir -p $(shell dirname $(GOPATH)/src/$(REPOPATH)) && ln -s -f $(shell pwd) $(GOPATH)/src/$(REPOPATH)
LOCALKUBEFILES := $(shell find pkg/localkube -name '*.go') $(shell find cmd/localkube -name '*.go') $(shell find vendor -name '*.go')
MINIKUBEFILES := $(shell find pkg/minikube -name '*.go') $(shell find cmd/minikube -name '*.go') $(shell find vendor -name '*.go')

out/minikube: out/minikube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minikube

out/localkube: $(LOCALKUBEFILES)
	$(MKGOPATH)
ifeq ($(GOOS),linux)
	CGO_ENABLED=1 go build -ldflags=$(LD_FLAGS) -o $(BUILD_DIR)/localkube ./cmd/localkube
else
	docker run -w /go/src/$(REPOPATH) -e IN_DOCKER=1 -v $(shell pwd):/go/src/$(REPOPATH) golang:1.6 make out/localkube
endif

out/minikube-$(GOOS)-$(GOARCH): $(MINIKUBEFILES) pkg/minikube/cluster/localkubecontents.go
	$(MKGOPATH)
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build --installsuffix cgo -a -o $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) ./cmd/minikube

.PHONY: integration
integration: out/minikube
	go test -v ./test/integration --tags=integration

.PHONY: test
test: pkg/minikube/cluster/localkubecontents.go
	$(MKGOPATH)
	./test.sh

pkg/minikube/cluster/localkubecontents.go: out/localkube $(GOPATH)/bin/go-bindata
	$(GOPATH)/bin/go-bindata -nomemcopy -o pkg/minikube/cluster/localkubecontents.go -pkg cluster ./out/localkube

$(GOPATH)/bin/go-bindata:
	$(MKGOPATH)
	go get github.com/jteeuwen/go-bindata/...
