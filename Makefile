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
GOPATH ?= $(shell pwd)/.gopath
REPOPATH ?= k8s.io/minikube
export GO15VENDOREXPERIMENT=1

clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)

.gopath:
	mkdir -p $(shell dirname $(GOPATH)/src/$(REPOPATH))
	ln -s -f $(shell pwd) $(GOPATH)/src/$(REPOPATH)

out/minikube: out/minikube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minikube

out/localkube: .gopath
ifeq ($(GOOS),linux)
	CGO_ENABLED=1 go build -v -x -ldflags="-s"-o $(BUILD_DIR)/localkube ./cmd/localkube
else
	docker run -w /go/src/k8s.io/minikube -e GOPATH=/go -v $(shell pwd):/go/src/k8s.io/minikube golang:1.6 make out/localkube
endif

out/minikube-$(GOOS)-$(GOARCH): .gopath pkg/minikube/cluster/localkubecontents.go
	CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build --installsuffix cgo -a -o $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) ./cmd/minikube

.PHONY: integration
integration: out/minikube
	go test -v ./test/integration --tags=integration

.PHONY: test
test: .gopath
	./test.sh

pkg/minikube/cluster/localkubecontents.go: out/localkube
	go-bindata -nomemcopy -o pkg/minikube/cluster/localkubecontents.go -pkg cluster ./out/localkube
