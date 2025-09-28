.PHONY: integration-kvm-prow

integration-kvm-prow:
	mkdir -p -m 0700 ~/.ssh
	cp -f "${GCE_SSH_PRIVATE_KEY_FILE}" ~/.ssh/google_compute_engine
	cp -f "${GCE_SSH_PUBLIC_KEY_FILE}" ~/.ssh/google_compute_engine.pub
	GOTOOLCHAIN=auto go build -C ./hack/prow/minitest -o $(PWD)/out/minitest .
	./out/minitest 
# 	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
# 	./hack/prow/linux_integration_kvm.sh
