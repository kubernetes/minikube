# Copyright 2016 The Kubernetes Authors All rights reserved.
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

FROM debian:jessie

ARG CNI_VERSION=v0.5.2

ENV ARCH amd64

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y \
    && DEBIAN_FRONTEND=noninteractive apt-get -yy -q install \
    iptables \
    ebtables \
    ethtool \
    ca-certificates \
    conntrack \
    util-linux \
    socat \
    git \
    nfs-common \
    glusterfs-client \
    cifs-utils \
    apt-transport-https \
    ca-certificates \
    curl \
    gnupg2 \
    software-properties-common \
    bridge-utils \
    ipcalc \
    && DEBIAN_FRONTEND=noninteractive apt-get upgrade -y \
    && DEBIAN_FRONTEND=noninteractive apt-get autoremove -y \
    && DEBIAN_FRONTEND=noninteractive apt-get clean && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN mkdir -p /etc/cni/net.d && \
    mkdir -p /opt/cni/bin && \
    curl -sSL --retry 5 https://github.com/containernetworking/cni/releases/download/${CNI_VERSION}/cni-${ARCH}-${CNI_VERSION}.tgz | \
    tar -C /opt/cni/bin -xz

# Install docker
RUN curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add -
RUN add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/debian \
   $(lsb_release -cs) \
   stable"
RUN apt-get update && apt-get install -yy -q docker-ce

# Install sshd
RUN apt-get update && apt-get install -y openssh-server
RUN mkdir /var/run/sshd

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd
# Change SSH Port to not conflict with Devshell
# TODO(aaron-prindle) make this configurable, perhaps including sshd at all
RUN sed -i 's/Port 22/Port 26/g' /etc/ssh/sshd_config
RUN sed -i 's/#AuthorizedKeysFile/AuthorizedKeysFile/g' /etc/ssh/sshd_config
RUN sed -i 's|authorized_keys|authorized_keys /etc/localkube-ssh/localkube_rsa.pub|g' /etc/ssh/sshd_config

# Copy over important files
COPY out/localkube /localkube
COPY deploy/docker/localkube-dind/start.sh /start.sh
COPY deploy/docker/localkube-dind/dindnet /dindnet
COPY deploy/addons/addon-manager.yaml /etc/kubernetes/manifests/addon-manager.yaml
COPY deploy/addons/dashboard /etc/kubernetes/addons
COPY deploy/addons/kube-dns /etc/kubernetes/addons

RUN chmod +x /localkube
RUN chmod +x /start.sh
RUN chmod +x /dindnet

CMD /usr/sbin/sshd -D
