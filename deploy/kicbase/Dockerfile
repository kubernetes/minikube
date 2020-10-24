# Copyright 2018 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# kind node base image
#
# For systemd + docker configuration used below, see the following references:
# https://systemd.io/CONTAINER_INTERFACE/

# start from ubuntu 20.04, this image is reasonably small as a starting point
# for a kubernetes node image, it doesn't contain much we don't need
FROM ubuntu:focal-20200925

# copy in static files (configs, scripts)
COPY 10-network-security.conf /etc/sysctl.d/10-network-security.conf
COPY 11-tcp-mtu-probing.conf /etc/sysctl.d/11-tcp-mtu-probing.conf
COPY clean-install /usr/local/bin/clean-install
COPY entrypoint /usr/local/bin/entrypoint

# Install dependencies, first from apt, then from release tarballs.
# NOTE: we use one RUN to minimize layers.
#
# First we must ensure that our util scripts are executable.
#
# The base image already has: ssh, apt, snapd, but we need to install more packages.
# Packages installed are broken down into (each on a line):
# - packages needed to run services (systemd)
# - packages needed for kubernetes components
# - packages needed by the container runtime
# - misc packages kind uses itself
# - packages that provide semi-core kubernetes functionality
# After installing packages we cleanup by:
# - removing unwanted systemd services
# - disabling kmsg in journald (these log entries would be confusing)
#
# Next we ensure the /etc/kubernetes/manifests directory exists. Normally
# a kubeadm debain / rpm package would ensure that this exists but we install
# freshly built binaries directly when we build the node image.
#
# Finally we adjust tempfiles cleanup to be 1 minute after "boot" instead of 15m
# This is plenty after we've done initial setup for a node, but before we are
# likely to try to export logs etc.
RUN echo "Ensuring scripts are executable ..." \
    && chmod +x /usr/local/bin/clean-install /usr/local/bin/entrypoint \
 && echo "Installing Packages ..." \
    && DEBIAN_FRONTEND=noninteractive clean-install \
      systemd \
      conntrack iptables iproute2 ethtool socat util-linux mount ebtables udev kmod \
      libseccomp2 pigz \
      bash ca-certificates curl rsync \
      nfs-common \
    && find /lib/systemd/system/sysinit.target.wants/ -name "systemd-tmpfiles-setup.service" -delete \
    && rm -f /lib/systemd/system/multi-user.target.wants/* \
    && rm -f /etc/systemd/system/*.wants/* \
    && rm -f /lib/systemd/system/local-fs.target.wants/* \
    && rm -f /lib/systemd/system/sockets.target.wants/*udev* \
    && rm -f /lib/systemd/system/sockets.target.wants/*initctl* \
    && rm -f /lib/systemd/system/basic.target.wants/* \
    && echo "ReadKMsg=no" >> /etc/systemd/journald.conf \
    && ln -s "$(which systemd)" /sbin/init \
 && echo "Ensuring /etc/kubernetes/manifests" \
    && mkdir -p /etc/kubernetes/manifests \
 && echo "Adjusting systemd-tmpfiles timer" \
    && sed -i /usr/lib/systemd/system/systemd-tmpfiles-clean.timer -e 's#OnBootSec=.*#OnBootSec=1min#' \
 && echo "Modifying /etc/nsswitch.conf to prefer hosts" \
    && sed -i /etc/nsswitch.conf -re 's#^(hosts:\s*).*#\1dns files#'

# tell systemd that it is in docker (it will check for the container env)
# https://systemd.io/CONTAINER_INTERFACE/
ENV container docker
# systemd exits on SIGRTMIN+3, not SIGTERM (which re-executes it)
# https://bugzilla.redhat.com/show_bug.cgi?id=1201657
STOPSIGNAL SIGRTMIN+3
# NOTE: this is *only* for documentation, the entrypoint is overridden later
ENTRYPOINT [ "/usr/local/bin/entrypoint", "/sbin/init" ]

ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind/blob/2c0eee40/images/base/Dockerfile
# which is an ubuntu 20.04 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
USER root

# install system requirements from the regular distro repositories
RUN clean-install \
    lz4 \
    gnupg \ 
    sudo \
    openssh-server \
    dnsutils \
    # libglib2.0-0 is required for conmon, which is required for podman
    libglib2.0-0

# install docker
RUN sh -c "echo 'deb https://download.docker.com/linux/ubuntu focal stable' > /etc/apt/sources.list.d/docker.list" && \
    curl -L https://download.docker.com/linux/ubuntu/gpg -o docker.key && \
    apt-key add - < docker.key && \
    clean-install docker-ce docker-ce-cli containerd.io

# Install cri-o/podman dependencies:
RUN sh -c "echo 'deb https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \
    curl -LO https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_20.04/Release.key && \
    apt-key add - < Release.key && \
    clean-install containers-common catatonit conmon containernetworking-plugins cri-tools podman-plugins

# install cri-o based on https://github.com/cri-o/cri-o/blob/release-1.18/README.md#installing-cri-o
RUN sh -c "echo 'deb https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.18/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable:cri-o:1.18.list" && \
    curl -LO https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable:/cri-o:/1.18/xUbuntu_20.04/Release.key && \
    apt-key add - < Release.key && \
    clean-install cri-o cri-o-runc

# install podman
RUN sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_20.04/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \
    curl -LO https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_20.04/Release.key && \
    apt-key add - < Release.key && \
    clean-install podman && \
    addgroup --system podman && \
    mkdir -p /etc/systemd/system/podman.socket.d && \
    printf "[Socket]\nSocketMode=0660\nSocketUser=root\nSocketGroup=podman\n" \
           > /etc/systemd/system/podman.socket.d/override.conf && \
    mkdir -p /etc/tmpfiles.d && \
    echo "d /run/podman 0770 root podman" > /etc/tmpfiles.d/podman.conf && \
    systemd-tmpfiles --create

# automount service
COPY automount/minikube-automount /usr/sbin/minikube-automount
COPY automount/minikube-automount.service /usr/lib/systemd/system/minikube-automount.service
RUN ln -fs /usr/lib/systemd/system/minikube-automount.service \
    /etc/systemd/system/multi-user.target.wants/minikube-automount.service

# disable non-docker runtimes by default
RUN systemctl disable containerd && systemctl disable crio && rm /etc/crictl.yaml
# enable docker which is default
RUN systemctl enable docker.service && systemctl enable podman.socket
# making SSH work for docker container 
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
RUN mkdir /var/run/sshd
RUN echo 'root:root' |chpasswd
RUN sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config

# minikube relies on /etc/hosts for control-plane discovery. This prevents nefarious DNS servers from breaking it.
RUN sed -ri 's/dns files/files dns/g' /etc/nsswitch.conf

EXPOSE 22
# create docker user for minikube ssh. to match VM using "docker" as username
RUN adduser --ingroup docker --disabled-password --gecos '' docker 
RUN adduser docker sudo
RUN adduser docker podman
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers
USER docker
RUN mkdir /home/docker/.ssh
USER root
# kind base-image entry-point expects a "kind" folder for product_name,product_uuid
# https://github.com/kubernetes-sigs/kind/blob/master/images/base/files/usr/local/bin/entrypoint
RUN mkdir -p /kind
# Deleting leftovers
RUN rm -rf \
  /usr/share/doc/* \
  /usr/share/man/* \
  /usr/share/local/* \
RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"
