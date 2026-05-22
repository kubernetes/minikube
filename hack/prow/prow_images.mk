# ----------------------------------------------------------------
# Below are the image push targets for prow
# ----------------------------------------------------------------	
PROW_IMAGE_PLATFORMS ?= linux/amd64,linux/arm64,linux/ppc64le,linux/s390x

.PHONY: push-kubernetes-bootcamp-image-prow
push-kubernetes-bootcamp-image-prow:
	docker buildx build --push --platform  $(PROW_IMAGE_PLATFORMS) \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kubernetes-bootcamp:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kubernetes-bootcamp:latest deploy/images/kubernetes-bootcamp


.PHONY: push-storage-provisioner-image-prow
push-storage-provisioner-image-prow: out/storage-provisioner-amd64 out/storage-provisioner-arm64 out/storage-provisioner-ppc64le out/storage-provisioner-s390x
	docker buildx build --push --platform  $(PROW_IMAGE_PLATFORMS) \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/storage-provisioner:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/storage-provisioner:latest -f deploy/storage-provisioner/Dockerfile .


.PHONY: push-gvisor-image-prow
push-gvisor-image-prow:
	docker buildx build --push --platform  $(PROW_IMAGE_PLATFORMS) \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/gvisor:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/gvisor:latest -f deploy/images/gvisor/Dockerfile .

.PHONY: push-kube-registry-proxy-image-prow
push-kube-registry-proxy-image-prow:
	docker buildx build --push --platform  $(PROW_IMAGE_PLATFORMS) \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kube-registry-proxy:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kube-registry-proxy:latest -f deploy/images/kube-registry-proxy/Dockerfile deploy/images/kube-registry-proxy

.PHONY: push-kicbase-image-prow
push-kicbase-image-prow:
	./hack/jenkins/build_changelog.sh deploy/kicbase/CHANGELOG || touch deploy/kicbase/CHANGELOG
	./hack/build_auto_pause.sh $(KICBASE_ARCH) $(CURDIR)/deploy/kicbase
	docker buildx build --push --platform  $(PROW_IMAGE_PLATFORMS) \
		-t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kicbase:$(_GIT_TAG) -t us-central1-docker.pkg.dev/k8s-staging-images/minikube/kicbase:latest -f deploy/kicbase/Dockerfile deploy/kicbase
