.PHONY: integration-kvm-prow

integration-kvm-prow:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/linux_integration_kvm.sh
