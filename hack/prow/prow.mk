# ----------------------------------------------------------------
# Below Integration tests run in a prow container (no external cloud vm)
# ----------------------------------------------------------------
.PHONY: integration-prow-docker-docker-linux-x86
integration-prow-docker-docker-linux-x86:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_docker_linux_x86.sh

.PHONY: integration-prow-docker-docker-linux-arm
integration-prow-docker-docker-linux-arm:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux arm64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_docker_linux_arm.sh

# TODO: rename integration-prow-docker-containerd-linux-arm in upstream test-infra
.PHONY: integration-prow-docker-containerd-arm
integration-prow-docker-containerd-arm:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux arm64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_containerd_linux_arm.sh

.PHONY: integration-prow-docker-containerd-linux-x86
integration-prow-docker-containerd-linux-x86:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_containerd_linux_x86.sh

.PHONY: integration-prow-docker-crio-linux-x86
integration-prow-docker-crio-linux-x86:
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_crio_linux_x86.sh

# ----------------------------------------------------------------
# Below Integration tests run in cloud VM using boskos
# ----------------------------------------------------------------

.PHONY: integration-prow-none-docker-linux-x86
integration-prow-none-docker-linux-x86: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester none-docker-linux-amd64-integration --config hack/prow/boskos-cfg-x86.json

.PHONY: integration-prow-none-containerd-linux-x86
integration-prow-none-containerd-linux-x86: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester none-containerd-linux-amd64-integration --config hack/prow/boskos-cfg-x86.json

.PHONY: integration-prow-kvm-docker-linux-x86
integration-prow-kvm-docker-linux-x86: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester kvm-docker-linux-amd64-integration --config hack/prow/boskos-cfg-x86.json

.PHONY: integration-prow-kvm-containerd-linux-x86
integration-prow-kvm-containerd-linux-x86: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester kvm-containerd-linux-amd64-integration --config hack/prow/boskos-cfg-x86.json

.PHONY: integration-prow-kvm-crio-linux-x86
integration-prow-kvm-crio-linux-x86: setup-prow-gcp-ssh-keys build-mini-test
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./out/minitest  --deployer boskos --tester kvm-crio-linux-amd64-integration --config hack/prow/boskos-cfg-x86.json

.PHONY: build-mini-test
build-mini-test: # build minitest binary
	GOTOOLCHAIN=auto go build -C ./hack/prow/minitest -o $(PWD)/out/minitest .

.PHONY: setup-prow-gcp-ssh-keys
setup-prow-gcp-ssh-keys: # set up ssh keys for gcloud cli. These env vars are set by test/infra
	mkdir -p -m 0700 ~/.ssh
	cp -f "${GCE_SSH_PRIVATE_KEY_FILE}" ~/.ssh/google_compute_engine
	cp -f "${GCE_SSH_PUBLIC_KEY_FILE}" ~/.ssh/google_compute_engine.pub
	
.PHONY: push-kubernetes-bootcamp
push-kubernetes-bootcamp:
	docker run --rm --privileged tonistiigi/binfmt:latest --install all
	docker buildx create --name multiarch --bootstrap
	docker buildx build --builder multiarch --push --platform  linux/amd64,linux/arm64 \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kubernetes-bootcamp:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kubernetes-bootcamp:latest deploy/image/kubernetes-bootcamp
	docker buildx rm multiarch