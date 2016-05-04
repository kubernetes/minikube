GO ?= go
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
GOPATH := $(shell pwd)/.gopath
REPOPATH ?= k8s.io/minikube

.PHONY: clean
clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)

.PHONY: gopath
gopath: clean
	rm -rf $(GOPATH)
	mkdir -p $(shell dirname $(GOPATH)/src/$(REPOPATH))
	ln -s $(shell pwd) $(GOPATH)/src/$(REPOPATH)

.PHONY: minikube
minikube: minikube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/minikube

.PHONY: localkube
localkube: localkube-$(GOOS)-$(GOARCH)
	cp $(BUILD_DIR)/localkube-$(GOOS)-$(GOARCH) $(BUILD_DIR)/localkube

.PHONY: minikube-$(GOOS)-$(GOARCH)
minikube-$(GOOS)-$(GOARCH): gopath
	GOPATH=$(GOPATH) CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO) build --installsuffix cgo -a -o $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) ./cmd/minikube

.PHONY: localkube-$(GOOS)-$(GOARCH)
localkube-$(GOOS)-$(GOARCH): gopath
	GOARCH=$(GOARCH) GOOS=$(GOOS) $(GO) build -o $(BUILD_DIR)/localkube-$(GOOS)-$(GOARCH) -i --ldflags '-extldflags "-static" --s -w' ./cmd/localkube

integration: minikube
	$(GO) test -v ./integration --tags=integration

.PHONY: test
test: gopath
	for TEST in $$(find -name "*.go" | grep -v vendor | grep -v integration | grep _test.go | cut -d/ -f2- | sed 's|/\w*.go||g' | uniq); do \
		echo $$TEST; \
		$(GO) test -v $(REPOPATH)/$${TEST}; \
	done
