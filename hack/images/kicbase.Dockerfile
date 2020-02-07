ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind
# which is an ubuntu 19.10 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
FROM kindest/base:v20200122-2dfe64b2 as base
USER root
RUN apt-get update && apt-get install -y \
  sudo \
  dnsutils \
  openssh-server \
  docker.io \
  && apt-get clean -y 
# disable containerd by default
RUN systemctl disable containerd
RUN rm /etc/crictl.yaml
# enable docker which is default
RUN systemctl enable docker
# making SSH work for docker container 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config
EXPOSE 22
# for minikube ssh. to match VM using "docker" as username
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
# Deleting leftovers
USER root
# kind base-image entry-point expects a "kind" folder for product_name,product_uuid
# https://github.com/kubernetes-sigs/kind/blob/master/images/base/files/usr/local/bin/entrypoint
RUN mkdir -p /kind
RUN rm -rf \
  /var/cache/debconf/* \
  /var/lib/apt/lists/* \
  /var/log/* \
  /tmp/* \
  /var/tmp/* \
  /usr/share/doc/* \
  /usr/share/man/* \
  /usr/share/local/* \
  RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"


FROM busybox
ARG KUBERNETES_VERSION
COPY out/preloaded-images-k8s-$KUBERNETES_VERSION.tar /preloaded-images.tar
RUN tar xvf /preloaded-images.tar -C /

FROM base
COPY --from=1 /var/lib/docker /var/lib/docker
