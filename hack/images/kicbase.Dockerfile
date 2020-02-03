ARG COMMIT_SHA
# for now using node image created by kind https://github.com/kubernetes-sigs/kind
# could be changed to slim ubuntu with systemd. 
FROM kindest/node:v1.16.2
USER root
RUN apt-get update && apt-get install -y \
  sudo \
  dnsutils \
  openssh-server \
  docker.io \
  && apt-get clean -y 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
# making SSH work for docker container 
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config
EXPOSE 22
# Deleting all "kind" related stuff from the image.
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
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
USER root
