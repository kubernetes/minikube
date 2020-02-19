ARG KIC_IMAGE_VERSION
FROM gcr.io/k8s-minikube/kicbase:$KIC_IMAGE_VERSION AS base

FROM busybox AS files
ARG KUBERNETES_VERSION
COPY out/preloaded-images-k8s-$KUBERNETES_VERSION.tar /preloaded-images.tar
RUN tar xvf /preloaded-images.tar -C /

FROM base
COPY --from=files /var/lib/docker /var/lib/docker
COPY --from=files /var/lib/minikube/binaries /var/lib/minikube/binaries
