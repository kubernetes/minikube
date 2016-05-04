GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
BUILD_DIR ?= ./out
GOPATH := $(shell pwd)/.gopath
REPOPATH ?= k8s.io/minikube
GO15VENDOREXPERIMENT := 1

clean:
	rm -rf $(GOPATH)
	rm -rf $(BUILD_DIR)

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

minikube-$(GOOS)-$(GOARCH): gopath
	GOPATH=$(GOPATH) CGO_ENABLED=0 GOARCH=$(GOARCH) GOOS=$(GOOS) go build --installsuffix cgo -a -o $(BUILD_DIR)/minikube-$(GOOS)-$(GOARCH) ./cmd/minikube

localkube-$(GOOS)-$(GOARCH): gopath
	GOARCH=$(GOARCH) GOOS=$(GOOS) CGO_ENABLED=1 go build -o $(BUILD_DIR)/localkube-$(GOOS)-$(GOARCH) -i ./cmd/localkube

.PHONY: integration
integration: minikube
	go test -v ./integration --tags=integration

.PHONY: test
test: gopath
	./test.sh
