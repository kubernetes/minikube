# Uses node image built by kind as base https://github.com/kubernetes-sigs/kind
# becuase already tested with kind, could be changed to ubuntu
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
