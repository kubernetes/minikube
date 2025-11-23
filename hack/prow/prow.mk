.PHONY: integration-prow-kvm-docker-linux-x86-64
integration-prow-docker-docker-linux-x86-64:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_docker_linux_x86-64.sh

.PHONY: integration-prow-none-docker-linux-x86-64
integration-prow-none-docker-linux-x86-64: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester none-docker-linux-amd64-integration --config hack/prow/bosksos-nested.json

.PHONY: integration-prow-kvm-docker-linux-x86-64
integration-prow-kvm-docker-linux-x86-64: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester kvm-docker-linux-amd64-integration --config hack/prow/bosksos-nested.json

.PHONY: build-mini-test
build-mini-test: # build minitest binary
	GOTOOLCHAIN=auto go build -C ./hack/prow/minitest -o $(PWD)/out/minitest .

.PHONY: setup-prow-gcp-ssh-keys
setup-prow-gcp-ssh-keys: # set up ssh keys for gcloud cli. These env vars are set by test/infra
	mkdir -p -m 0700 ~/.ssh
	cp -f "${GCE_SSH_PRIVATE_KEY_FILE}" ~/.ssh/google_compute_engine
	cp -f "${GCE_SSH_PUBLIC_KEY_FILE}" ~/.ssh/google_compute_engine.pub