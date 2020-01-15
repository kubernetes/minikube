# Thanks to kind (https://github.com/kubernetes-sigs/kind) the base image uses node image built by kind  
# which is an ubuntu image minimized with basic tools to run systemd. 
# in next iterations could be changed to a smaller basee image, for now manually deleting kind-related files from the node iamge.
# more info https://kind.sigs.k8s.io/docs/design/node-image/
ARG COMMIT_SHA
FROM kindest/node:v1.16.2
USER root
RUN apt-get update && apt-get install -y \
  sudo \
  dnsutils \
  && apt-get clean -y 
# Remove kind related binaries and images
RUN rm -rf \
    /var/cache/debconf/* \
    /var/lib/apt/lists/* \
    /var/log/* \
    /tmp/* \
    /var/tmp/* \
    /usr/share/doc/* \
    /usr/share/man/* \
    /usr/share/local/* \
    /kind/bin/kubeadm /kind/bin/kubelet /kind/systemd /kind/images /kind/manifests
RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"
