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
GOPATH := $(shell pwd)/.gopath
REPOPATH ?= k8s.io/minikube
export GO15VENDOREXPERIMENT=1

clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)

gopath:
	mkdir -p $(shell dirname $(GOPATH)/src/$(REPOPATH))
	ln -s -f $(shell pwd) $(GOPATH)/src/$(REPOPATH)

.PHONY: minikube
minikube: minikube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minikube

.PHONY: localkube
localkube: localkube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/localkube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/localkube

minikube-$(GOOS)-$(GOARCH): gopath
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build --installsuffix cgo -a -o $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) ./cmd/minikube

localkube-$(GOOS)-$(GOARCH): gopath
	GOARCH=$(GOARCH) GOOS=$(GOOS) go build -o $(BUILD_DIR)/localkube-$(GOOS)-$(GOARCH) ./cmd/localkube

.PHONY: integration
integration: minikube
	go test -v ./test/integration --tags=integration

localkube-incremental:
	GOPATH=/go CGO_ENABLED=1 GOBIN=$(shell pwd)/$(BUILD_DIR) go install ./cmd/localkube

docker/localkube:
	docker run -w /go/src/k8s.io/minikube -v $(shell pwd):/go/src/k8s.io/minikube golang:1.6 make localkube-incremental

.PHONY: test
test: gopath
	./test.sh
