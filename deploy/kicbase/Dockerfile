ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind/blob/v0.8.1/images/base/Dockerfile
# which is an ubuntu 20.04 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
FROM kindest/base:v20200430-2c0eee40 as base
USER root
# specify version of everything explicitly using 'apt-cache policy'
RUN apt-get update && apt-get install -y --no-install-recommends \
    lz4 \
    gnupg \ 
    sudo \
    docker.io \
    openssh-server \
    dnsutils \
    runc \
    # libglib2.0-0 is required for conmon, which is required for podman
    libglib2.0-0 \
    # removing kind's crictl config
    && rm /etc/crictl.yaml

# Install cri-o/podman dependencies:
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \ 
    curl -LO https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_20.04/Release.key && \
    apt-key add - < Release.key && apt-get update && \
    apt-get install -y --no-install-recommends containers-common catatonit conmon containernetworking-plugins podman-plugins varlink

# install cri-o based on https://github.com/cri-o/cri-o/commit/96b0c34b31a9fc181e46d7d8e34fb8ee6c4dc4e1#diff-04c6e90faac2675aa89e2176d2eec7d8R128
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.18:/1.18.3/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \
    curl -LO https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.18:/1.18.3/xUbuntu_20.04/Release.key && \
    apt-key add - < Release.key && apt-get update && \
    apt-get install -y --no-install-recommends cri-o=1.18.3~2

# install podman
RUN sh -c "echo 'deb https://dl.bintray.com/afbjorklund/podman focal main' > /etc/apt/sources.list.d/podman.list" && \
    curl -L https://bintray.com/user/downloadSubjectPublicKey?username=afbjorklund -o afbjorklund-public.key.asc && \
    apt-key add - < afbjorklund-public.key.asc && apt-get update && \
    apt-get install -y --no-install-recommends podman=1.9.3~1

RUN mkdir -p /usr/lib/cri-o-runc/sbin && cp /usr/local/sbin/runc /usr/lib/cri-o-runc/sbin/runc

COPY entrypoint /usr/local/bin/entrypoint
# automount service
COPY automount/minikube-automount /usr/sbin/minikube-automount
COPY automount/minikube-automount.service /usr/lib/systemd/system/minikube-automount.service
RUN ln -fs /usr/lib/systemd/system/minikube-automount.service \
    /etc/systemd/system/multi-user.target.wants/minikube-automount.service

# disable non-docker runtimes by default
RUN systemctl disable containerd && systemctl disable crio && rm /etc/crictl.yaml
# enable docker which is default
RUN systemctl enable docker
# making SSH work for docker container 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config

EXPOSE 22
# create docker user for minikube ssh. to match VM using "docker" as username
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
USER root
# kind base-image entry-point expects a "kind" folder for product_name,product_uuid
# https://github.com/kubernetes-sigs/kind/blob/master/images/base/files/usr/local/bin/entrypoint
RUN mkdir -p /kind
# Deleting leftovers
RUN apt-get clean -y && rm -rf \
  /var/cache/debconf/* \
  /var/lib/apt/lists/* \
  /var/log/* \
  /tmp/* \
  /var/tmp/* \
  /usr/share/doc/* \
  /usr/share/man/* \
  /usr/share/local/* \
RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"
