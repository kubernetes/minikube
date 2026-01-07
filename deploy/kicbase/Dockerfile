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


# this ARG needs to be global to use it in `FROM` & is updated for new versions of debian:bookworm-slim-*
ARG KICBASE_IMAGE="debian:bookworm-20251208-slim"
# start from debin 12, this image is reasonably small as a starting point
# for a kubernetes node image, it doesn't contain much we don't need
FROM ${KICBASE_IMAGE} AS kicbase

ARG BUILDKIT_VERSION="v0.26.3"
ARG CRIO_VERSION="v1.35"
ARG CRICTL_VERSION="v1.35.0"
ARG CRI_DOCKERD_VERSION="v0.4.1"
ARG CRI_DOCKERD_COMMIT="55d6e1a1d6f2ee58949e13a0c66afe7d779ac942"
ARG CNI_PLUGINS_VERSION="v1.9.0"
ARG TARGETARCH
ARG NERDCTL_VERSION="2.2.1"
ARG NERDCTLD_VERSION="0.7.0"

# copy in static files (configs, scripts)
COPY 10-network-security.conf /etc/sysctl.d/10-network-security.conf
COPY 11-tcp-mtu-probing.conf /etc/sysctl.d/11-tcp-mtu-probing.conf
COPY 02-crio.conf /etc/crio/crio.conf.d/02-crio.conf
COPY containerd.toml /etc/containerd/config.toml
COPY containerd_docker_io_hosts.toml /etc/containerd/certs.d/docker.io/hosts.toml
COPY clean-install /usr/local/bin/clean-install
COPY entrypoint /usr/local/bin/entrypoint
COPY CHANGELOG ./CHANGELOG
COPY nerdctld/nerdctld.socket  /etc/systemd/system/nerdctld.socket
COPY nerdctld/nerdctld.service  /etc/systemd/system/nerdctld.service
COPY auto-pause-${TARGETARCH} /bin/auto-pause

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
# a kubeadm debian / rpm package would ensure that this exists but we install
# freshly built binaries directly when we build the node image.
#
# Finally we adjust tempfiles cleanup to be 1 minute after "boot" instead of 15m
# This is plenty after we've done initial setup for a node, but before we are
# likely to try to export logs etc.
RUN echo "Ensuring scripts are executable ..." \
    && chmod +x /usr/local/bin/clean-install /usr/local/bin/entrypoint \
 && echo "Installing Packages ..." \
    && DEBIAN_FRONTEND=noninteractive clean-install \
      systemd dbus \
      conntrack iptables iproute2 ethtool socat util-linux mount ebtables udev kmod \
      libseccomp2 pigz \
      bash ca-certificates curl rsync \
      nfs-common \
      iputils-ping netcat-openbsd vim-tiny \
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
 && echo "Disabling udev" \
    && systemctl disable udev.service \
 && echo "Modifying /etc/nsswitch.conf to prefer hosts" \
    && sed -i /etc/nsswitch.conf -re 's#^(hosts:\s*).*#\1dns files#'

# delete this file due to https://github.com/kubernetes/minikube/issues/17700
RUN rm -f /usr/lib/binfmt.d/python3.10.conf

ARG COMMIT_SHA
# using base image created by kind https://github.com/kubernetes-sigs/kind/blob/b6bc1125/images/base/Dockerfile
# available as a docker image: docker.io/kindest/base:v20210402-3d9112b0
# which is an ubuntu 20.10 with an entry-point that helps running systemd
# could be changed to any debian that can run systemd
USER root

# Install cri-dockerd from pre-compiled binaries stored in GCS, this is way faster than building from source in multi-arch
RUN echo "Installing cri-dockerd ${CRI_DOCKERD_VERSION}" && \
	curl -L "https://storage.googleapis.com/kicbase-artifacts/cri-dockerd/${CRI_DOCKERD_COMMIT}/${TARGETARCH}/cri-dockerd" -o /usr/bin/cri-dockerd && chmod +x /usr/bin/cri-dockerd && \
	curl -L "https://storage.googleapis.com/kicbase-artifacts/cri-dockerd/${CRI_DOCKERD_COMMIT}/cri-docker.socket" -o /usr/lib/systemd/system/cri-docker.socket && \
	curl -L "https://storage.googleapis.com/kicbase-artifacts/cri-dockerd/${CRI_DOCKERD_COMMIT}/cri-docker.service" -o /usr/lib/systemd/system/cri-docker.service

# install system requirements from the regular distro repositories
RUN clean-install \
    lz4 \
    gnupg \
    sudo \
    openssh-server \
    dnsutils

# Add support for rsa1 in sshd
# modern debian-based OSs dont support rsa1 by default, so we need to enable it to support older ssh clients
# TODO: remove after https://github.com/kubernetes/minikube/issues/21543 is solved
RUN cat <<EOF >> /etc/ssh/sshd_config
PubkeyAcceptedAlgorithms +ssh-rsa
HostkeyAlgorithms +ssh-rsa
PubkeyAuthentication yes
PasswordAuthentication no
EOF

# Install nerdctl and nerdctld
RUN export ARCH=$(dpkg --print-architecture) \
    && if [ "$ARCH" = 'amd64' ] || [ "$ARCH" = 'arm64' ]; then \
        echo "Installing nerdctl and nerdctld ..." && \
        curl -L --retry 5 --output /tmp/nerdctl.tgz "https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-$ARCH.tar.gz" &&\
        tar -C /usr/local/bin -xzvf /tmp/nerdctl.tgz &&\
        curl -L --retry 5 --output /tmp/nerdctld.tgz "https://github.com/afbjorklund/nerdctld/releases/download/v${NERDCTLD_VERSION}/nerdctld-${NERDCTLD_VERSION}-linux-$ARCH.tar.gz" &&\
        tar -C /usr/local/bin -xzvf /tmp/nerdctld.tgz &&\
        chmod 777  /usr/local/bin/nerdctl &&\
        chmod 777 /usr/local/bin/nerdctld; \
    fi

# install docker
## ref: https://docs.docker.com/engine/install/debian/#install-using-the-repository
RUN install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc \
    && chmod a+r /etc/apt/keyrings/docker.asc \
    && echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null \
    && clean-install docker-ce docker-ce-cli containerd.io docker-buildx-plugin

# Set ulimit for containerd to match docker's default-ulimit setting
# This ensures consistent file descriptor limits (1048576) across container runtimes
RUN sed -i '/^\[Service\]/a LimitNOFILE=1048576' /lib/systemd/system/containerd.service

# install buildkit
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm-v7/') \
 && echo "Installing buildkit ..." \
    && addgroup --system buildkit \
    && export BUILDKIT_BASE_URL="https://github.com/moby/buildkit/releases/download/${BUILDKIT_VERSION}" \
    && curl -sSL --retry 5 --output /tmp/buildkit.tgz "${BUILDKIT_BASE_URL}/buildkit-${BUILDKIT_VERSION}.linux-${ARCH}.tar.gz" \
    && tar -C /usr/local -xzvf /tmp/buildkit.tgz \
    && rm -rf /tmp/buildkit.tgz \
    && mkdir -p /usr/local/lib/systemd/system \
    && curl -L --retry 5 --output /usr/local/lib/systemd/system/buildkit.service "https://raw.githubusercontent.com/moby/buildkit/${BUILDKIT_VERSION}/examples/systemd/system/buildkit.service" \
    && curl -L --retry 5 --output /usr/local/lib/systemd/system/buildkit.socket "https://raw.githubusercontent.com/moby/buildkit/${BUILDKIT_VERSION}/examples/systemd/system/buildkit.socket" \
    && mkdir -p /etc/buildkit \
    && echo "[worker.oci]\n  enabled = false\n[worker.containerd]\n  enabled = true\n  namespace = \"k8s.io\"" > /etc/buildkit/buildkitd.toml \
    && chmod 755 /usr/local/bin/buildctl \
    && chmod 755 /usr/local/bin/buildkit-runc \
    && chmod 755 /usr/local/bin/buildkit-qemu-* \
    && chmod 755 /usr/local/bin/buildkitd \
    && systemctl enable buildkit.socket

# install podman
RUN clean-install podman catatonit crun && \
    addgroup --system podman && \
    mkdir -p /etc/systemd/system/podman.socket.d && \
    printf "[Socket]\nSocketMode=0660\nSocketUser=root\nSocketGroup=podman\n" \
           > /etc/systemd/system/podman.socket.d/override.conf && \
    mkdir -p /etc/tmpfiles.d && \
    echo "d /run/podman 0770 root podman" > /etc/tmpfiles.d/podman.conf && \
    systemd-tmpfiles --create

 # install crictl
RUN export ARCH=$(dpkg --print-architecture) && \
    case "$ARCH" in \
      amd64) CRICTL_ARCH="amd64" ;; \
      arm64) CRICTL_ARCH="arm64" ;; \
      armhf) CRICTL_ARCH="armhf" ;; \
      ppc64el) CRICTL_ARCH="ppc64le" ;; \
      s390x) CRICTL_ARCH="s390x" ;; \
      *) echo "Unsupported architecture for crictl: $ARCH" && exit 1 ;; \
    esac && \
    if [ "$CRICTL_ARCH" != "armhf" ]; then \
      curl -fsSL --retry 5 --output /tmp/crictl.tgz "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${CRICTL_ARCH}.tar.gz" && \
      tar -C /usr/local/bin -xzvf /tmp/crictl.tgz && \
      rm -f /tmp/crictl.tgz && \
      chmod +x /usr/local/bin/crictl && \
      if [ -f /usr/local/bin/critest ]; then chmod +x /usr/local/bin/critest; fi; \
    else \
      echo "Skipping crictl install for armhf: No asset available in v1.34.0"; \
    fi   
# install containernetworking-plugins
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm/') && \
    curl -LO "https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-$ARCH-${CNI_PLUGINS_VERSION}.tgz" && \
    mkdir -p /opt/cni/bin && \
    tar -xf "cni-plugins-linux-$ARCH-${CNI_PLUGINS_VERSION}.tgz" -C /opt/cni/bin && \
    rm "cni-plugins-linux-$ARCH-${CNI_PLUGINS_VERSION}.tgz"

# install cri-o from the OBS repositories
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm-v7/') && \
    if [ "$ARCH" != "ppc64le" ] && [ "$ARCH" != "arm-v7" ]; then \
        mkdir -p /etc/apt/sources.list.d /etc/apt/trusted.gpg.d && \
        echo 'deb http://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/Debian_12/ /' \
            > /etc/apt/sources.list.d/devel:kubic:libcontainers:stable.list && \
        curl -fsSL https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/Debian_12/Release.key \
            | gpg --dearmor --yes -o /etc/apt/trusted.gpg.d/devel_kubic_libcontainers_stable.gpg && \
        echo "deb http://download.opensuse.org/repositories/isv:/cri-o:/stable:/${CRIO_VERSION}:/build/deb/ /" \
            > "/etc/apt/sources.list.d/isv:cri-o:stable:${CRIO_VERSION}:build.list" && \
        curl -fsSL "https://download.opensuse.org/repositories/isv:cri-o:stable:${CRIO_VERSION}:build/deb/Release.key" \
            | gpg --dearmor --yes -o "/etc/apt/trusted.gpg.d/isv_cri-o_stable_${CRIO_VERSION}_build.gpg" && \
        clean-install cri-o; \
    fi

# install NVIDIA container toolkit
RUN export ARCH=$(dpkg --print-architecture) && \
    if [ "$ARCH" = 'amd64' ] || [ "$ARCH" = 'arm64' ] || [ "$ARCH" = 'ppc64el' ]; then \
    curl -fsSL https://nvidia.github.io/libnvidia-container/gpgkey | sudo gpg --dearmor -o /usr/share/keyrings/nvidia-container-toolkit-keyring.gpg && \
    curl -s -L https://nvidia.github.io/libnvidia-container/stable/deb/nvidia-container-toolkit.list | \
    sed 's#deb https://#deb [signed-by=/usr/share/keyrings/nvidia-container-toolkit-keyring.gpg] https://#g' | \
    sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list && \
    clean-install nvidia-container-toolkit; fi

# install version.json
ARG VERSION_JSON
RUN echo "${VERSION_JSON}" > /version.json

# automount service
COPY automount/minikube-automount /usr/sbin/minikube-automount
COPY automount/minikube-automount.service /usr/lib/systemd/system/minikube-automount.service
RUN ln -fs /usr/lib/systemd/system/minikube-automount.service \
    /etc/systemd/system/multi-user.target.wants/minikube-automount.service

# scheduled stop service
COPY scheduled-stop/minikube-scheduled-stop /var/lib/minikube/scheduled-stop/minikube-scheduled-stop
COPY scheduled-stop/minikube-scheduled-stop.service /usr/lib/systemd/system/minikube-scheduled-stop.service
RUN  chmod +x /var/lib/minikube/scheduled-stop/minikube-scheduled-stop

# disable non-docker runtimes by default (since cri-dockerd is default in minikube)
RUN systemctl disable containerd
# disable crio for archs that support it
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/' | sed 's/armhf/arm-v7/') && \
    if [ "$ARCH" != "ppc64le" ] && [ "$ARCH" != "arm-v7" ]; then systemctl disable crio && rm -f /etc/crictl.yaml; fi
# enable podman socket on archs that support it
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/') && if [ "$ARCH" != "ppc64le" ]; then systemctl enable podman.socket; fi
# enable docker which is default
RUN systemctl enable docker.service
# minikube relies on /etc/hosts for control-plane discovery. This prevents nefarious DNS servers from breaking it.
RUN sed -ri 's/dns files/files dns/g' /etc/nsswitch.conf

EXPOSE 22
# create docker user for minikube ssh. to match VM using "docker" as username
# Ensure 'docker' group exists; create a 'docker' user with a shell
## ref: https://docs.docker.com/engine/install/linux-postinstall/
RUN groupadd -f docker \
    && useradd -m -s /bin/bash -g docker -G sudo docker
# Ensure the account is usable for key-only SSH - unlock the account, and disable password
RUN usermod -p '*' docker && passwd -u docker && passwd -d docker
RUN echo '%sudo ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/20-passwordless-sudo \
    && chmod 0440 /etc/sudoers.d/20-passwordless-sudo
RUN export ARCH=$(dpkg --print-architecture | sed 's/ppc64el/ppc64le/') && if [ "$ARCH" != "ppc64le" ]; then adduser docker podman; fi
RUN adduser docker buildkit
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
  /usr/share/local/*
RUN echo "kic! Build: ${COMMIT_SHA} Time :$(date)" > "/kic.txt"

# squash all layers into one
FROM scratch
COPY --from=kicbase / /
EXPOSE 22
# tell systemd that it is in docker (it will check for the container env)
# https://systemd.io/CONTAINER_INTERFACE/
ENV container=docker
# systemd exits on SIGRTMIN+3, not SIGTERM (which re-executes it)
# https://bugzilla.redhat.com/show_bug.cgi?id=1201657
STOPSIGNAL SIGRTMIN+3
# NOTE: this is *only* for documentation, the entrypoint is overridden later
ENTRYPOINT [ "/usr/local/bin/entrypoint", "/sbin/init" ]
