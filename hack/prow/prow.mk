.PHONY: integration-prow-kvm-docker-linux-x86-64


integration-prow-docker-docker-linux-x86-64:
# 	build first
#	container-runtime=docker driver=docker on linux/amd64
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_docker_linux_x86-64.sh

integration-prow-none-docker-linux-x86-64:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_none_docker_linux_x86-64.sh

integration-prow-kvm-docker-linux-x86-64:
# 	build first
#	container-runtime=docker driver=kvm on linux/amd64
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
# set up ssh keys for gcloud cli. These env vars are set by test/infra
	mkdir -p -m 0700 ~/.ssh
	cp -f "${GCE_SSH_PRIVATE_KEY_FILE}" ~/.ssh/google_compute_engine
	cp -f "${GCE_SSH_PUBLIC_KEY_FILE}" ~/.ssh/google_compute_engine.pub
	GOTOOLCHAIN=auto go build -C ./hack/prow/minitest -o $(PWD)/out/minitest .
	./out/minitest  --deployer boskos --tester kvm-docker-linux-amd64-integration --config hack/prow/kvm.json