ARG COMMIT_SHA
# using node image created by kind https://github.com/kubernetes-sigs/kind
# could be changed to ubuntu.
FROM kindest/node:v1.16.2
USER root
RUN apt-get update && apt-get install -y \
  sudo \
  dnsutils \
  && apt-get clean -y 
# Deleting all kind related stuff from the image.
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
# for minikube ssh. to match VM using docker username
RUN useradd -ms /bin/bash docker 
USER docker
RUN mkdir /home/docker/.ssh

