.PHONY: integration-prow-kvm-docker-linux-x86-64 integration-prow-docker-docker-linux-x86-64  push-kubernetes-bootcamp

integration-prow-docker-docker-linux-x86-64:
# 	build first
#	container-runtime=docker driver=docker on linux/amd64
	./hack/prow/minikube_cross_build.sh $(GO_VERSION) linux amd64
	./hack/prow/util/integration_prow_wrapper.sh ./hack/prow/integration_docker_docker_linux_x86-64.sh

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

push-kubernetes-bootcamp:
	docker run --rm --privileged tonistiigi/binfmt:latest --install all
	docker buildx create --name multiarch --bootstrap
	docker buildx build --builder multiarch --push --platform  linux/amd64,linux/arm64 \
		-t gcr.io/minikube/kubernetes-bootcamp:$(_GIT_TAG) -t gcr.io/minikube/kubernetes-bootcamp:latest deploy/image/kubernetes-bootcamp
	docker buildx rm multiarch