GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

.PHONY: minikube-e2e-fast
minikube-e2e-fast:
	./tests/e2e/fast.sh

minikube-e2e-conformance:
	./tests/e2e/conformance.sh

integration-kvm-prow: #temp for prow
	./tests/e2e/fast.sh
