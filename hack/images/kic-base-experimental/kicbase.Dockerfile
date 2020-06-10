# Copyright 2019 The Kubernetes Authors All rights reserved.
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

FROM ubuntu:focal-20200319 as base

COPY files/ /usr/local/bin/

ARG COMMIT_SHA

RUN echo "set ENV variables ..." \
 && export CRIO_VERSION="1.17.4~1" \
    PODMAN_VERSION="1.9.3~1" \
    CNI_VERSION="v0.8.5" \
    DOCKER_VERSION="19.03.8-0ubuntu1" \
    CRICTL_VERSION="v1.17.0" \
    CONTAINERD_VERSION="1.3.3-61-g60bc1282" \
 && echo "Ensuring scripts are executable ..." \
    && chmod +x /usr/local/bin/entrypoint \
 && echo "Installing Packages ..." \
    && DEBIAN_FRONTEND=noninteractive apt-get update && \
      apt-get install -y --no-install-recommends \
      systemd conntrack iptables iproute2 ethtool \
      socat util-linux mount ebtables udev kmod \
      gnupg libglib2.0-0 libseccomp2 ca-certificates \
      curl rsync \
      && rm -rf /var/lib/apt/lists/* \
    && sh -c "echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/xUbuntu_19.10/ /' > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list" && \
        curl -LO https://download.opensuse.org/repositories/devel:kubic:libcontainers:stable/xUbuntu_19.10/Release.key && \
        apt-key add - < Release.key && apt-get update && \
        apt-get install -y --no-install-recommends cri-o-1.17=${CRIO_VERSION} podman=${PODMAN_VERSION} lz4 sudo \
        docker.io=${DOCKER_VERSION} dnsutils \
    && find /lib/systemd/system/sysinit.target.wants/ -name "systemd-tmpfiles-setup.service" -delete \
    && rm -f /lib/systemd/system/multi-user.target.wants/* \
    && rm -f /etc/systemd/system/*.wants/* \
    && rm -f /lib/systemd/system/local-fs.target.wants/* \
    && rm -f /lib/systemd/system/sockets.target.wants/*udev* \
    && rm -f /lib/systemd/system/sockets.target.wants/*initctl* \
    && rm -f /lib/systemd/system/basic.target.wants/* \
    && echo "ReadKMsg=no" >> /etc/systemd/journald.conf \
    && ln -s "$(which systemd)" /sbin/init \
 && echo "Installing containerd ..." \
    && export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm/') \
    && export CONTAINERD_BASE_URL="https://github.com/kind-ci/containerd-nightlies/releases/download/containerd-${CONTAINERD_VERSION}" \
    && curl -sSL --retry 5 --output /tmp/containerd.tgz "${CONTAINERD_BASE_URL}/containerd-${CONTAINERD_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/local -xzvf /tmp/containerd.tgz \
    && rm -rf /tmp/containerd.tgz \
    && rm -f /usr/local/bin/containerd-stress /usr/local/bin/containerd-shim-runc-v1 \
    && curl -sSL --retry 5 --output /usr/local/sbin/runc "${CONTAINERD_BASE_URL}/runc.${ARCH}" \
    && chmod 755 /usr/local/sbin/runc \
    && containerd --version \
 && echo "Installing crictl ..." \
    && curl -fSL "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" | tar xzC /usr/local/bin \
    && rm -rf crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz \
 && echo "Installing CNI binaries ..." \
    && export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm/') \
    && export CNI_TARBALL="${CNI_VERSION}/cni-plugins-linux-${ARCH}-${CNI_VERSION}.tgz" \
    && export CNI_URL="https://github.com/containernetworking/plugins/releases/download/${CNI_TARBALL}" \
    && curl -sSL --retry 5 --output /tmp/cni.tgz "${CNI_URL}" \
    && mkdir -p /opt/cni/bin \
    && tar -C /opt/cni/bin -xzf /tmp/cni.tgz \
    && rm -rf /tmp/cni.tgz \
    && find /opt/cni/bin -type f -not \( \
         -iname host-local \
         -o -iname ptp \
         -o -iname portmap \
         -o -iname loopback \
      \) \
      -delete \
 && echo "Ensuring /etc/kubernetes/manifests" \
    && mkdir -p /etc/kubernetes/manifests \
 && echo "Adjusting systemd-tmpfiles timer" \
    && sed -i /usr/lib/systemd/user/systemd-tmpfiles-clean.timer -e 's#OnBootSec=.*#OnBootSec=1min#'

# tell systemd that it is in docker (it will check for the container env)
# https://www.freedesktop.org/wiki/Software/systemd/ContainerInterface/
ENV container docker
# systemd exits on SIGRTMIN+3, not SIGTERM (which re-executes it)
# https://bugzilla.redhat.com/show_bug.cgi?id=1201657
STOPSIGNAL SIGRTMIN+3
# NOTE: this is *only* for documentation, the entrypoint is overridden later
ENTRYPOINT [ "/usr/local/bin/entrypoint", "/sbin/init" ]

# In this step we First disable non-docker runtimes by default
# then enable docker which is default
# next making SSH work for docker container
# based on https://github.com/rastasheep/ubuntu-sshd/blob/master/18.04/Dockerfile
# next create docker user for minikube ssh. to match VM using "docker" as username
# finally deleting leftovers
RUN echo "disable non-docker runtimes ..." \
    && systemctl disable containerd && systemctl disable crio \
    && systemctl enable docker \
 && echo "making SSH work for docker ..." \
    && apt-get install -y --no-install-recommends openssh-server \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir /var/run/sshd \
    && echo 'root:root' |chpasswd \
    && sed -ri 's/^#?PermitRootLogin\s+.*/PermitRootLogin yes/' /etc/ssh/sshd_config \
    && sed -ri 's/UsePAM yes/#UsePAM yes/g' /etc/ssh/sshd_config \
 && echo "create docker user for minikube ssh ..." \
    && adduser --ingroup docker --disabled-password --gecos '' docker \
    && adduser docker sudo \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
 && echo "creating docker folders && delete leftovers ..." \
 && mkdir /home/docker/.ssh \
    && mkdir -p /kind \
 && apt-get clean -y && rm -rf \
    /var/cache/debconf/* \
    /var/lib/apt/lists/* \
    /var/log/* \
    /tmp/* \
    /var/tmp/* \
    /usr/share/doc/* \
    /usr/share/man/* \
    /usr/share/local/* \
 && echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"