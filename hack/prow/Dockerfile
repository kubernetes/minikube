FROM docker:17.12.0-ce as static-docker-source
FROM debian:stretch
MAINTAINER Don McCasland <donmccasland@google.com>
CMD ["bash"]
RUN apt-get update \
    && apt-get install -qqy \
    pkg-config \
    libvirt-clients \
    libvirt-daemon \
    libvirt-daemon-system \
    libvirt-dev \
    qemu-kvm \
    gcc \
    g++ \
    libc6-dev \
    wget \
    make \
    git \
    patch \
    file \
    jq \
    gnupg \
    apt-utils \
    clang \
    bzip2 \
    cmake \
    curl \
    python-dev \
    python-setuptools \
    apt-transport-https \
    lsb-release \
    openssh-client
RUN wget https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.11.1.linux-amd64.tar.gz
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" &&  chmod -R 777 "$GOPATH"
ARG CLOUD_SDK_VERSION=221.0.0
ENV CLOUD_SDK_VERSION=$CLOUD_SDK_VERSION
COPY --from=static-docker-source /usr/local/bin/docker /usr/local/bin/docker
ENV APT_KEY_DONT_WARN_ON_DANGEROUS_USAGE=DontWarn
RUN easy_install -U pip && \
    pip install -U crcmod   && \
    export CLOUD_SDK_REPO="cloud-sdk-$(lsb_release -c -s)" && \
    echo "deb https://packages.cloud.google.com/apt $CLOUD_SDK_REPO main" > /etc/apt/sources.list.d/google-cloud-sdk.list && \
    curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key add - && \
    apt-get update && \
    apt-get install -y google-cloud-sdk=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-app-engine-python=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-app-engine-python-extras=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-app-engine-java=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-app-engine-go=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-datalab=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-datastore-emulator=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-pubsub-emulator=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-bigtable-emulator=${CLOUD_SDK_VERSION}-0 \
        google-cloud-sdk-cbt=${CLOUD_SDK_VERSION}-0 \
        kubectl && \
    gcloud config set core/disable_usage_reporting true && \
    gcloud config set component_manager/disable_update_check true && \
    gcloud config set metrics/environment github_docker_image && \
    gcloud --version && \
    docker --version && kubectl version --client
WORKDIR /go
ENV PWD /go/src/k8s.io/minikube
ENV GOARM=7
ENV KUBE_DYNAMIC_CROSSPLATFORMS="armhf arm64 s390x ppc64el"
ENV KUBE_CROSSPLATFORMS="linux/386 linux/arm linux/arm64 linux/ppc64le linux/s390x darwin/amd64 darwin/386 windows/amd64 windows/386"
RUN for platform in ${KUBE_CROSSPLATFORMS}; do GOOS=${platform%/*} GOARCH=${platform##*/} go install std; done
RUN echo "deb http://archive.ubuntu.com/ubuntu xenial main universe" > /etc/apt/sources.list.d/cgocrosscompiling.list \
   &&  apt-key adv --no-tty --keyserver keyserver.ubuntu.com --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32 \
   &&  apt-get update \
   &&  apt-get install -y build-essential \
   &&  for platform in ${KUBE_DYNAMIC_CROSSPLATFORMS}; do apt-get install -y crossbuild-essential-${platform}; done \
   &&  apt-get clean \
   &&  rm -rf /var/lib/apt/lists/*
ENV OSX_SDK=MacOSX10.11.sdk
ENV OSX_NDK_X86=/usr/local/osx-ndk-x86
ENV PATH=/go/bin:/osxcross/target/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/osx-ndk-x86/bin
RUN OSX_SDK_PATH=https://storage.googleapis.com/osx-sdk/$OSX_SDK.tar.bz2 \
   && git clone https://github.com/tpoechtrager/osxcross.git /osxcross \
#   && git clone https://github.com/tpoechtrager/osxcross.git /apple-libtapi \
#   && git clone https://github.com/tpoechtrager/cctools-port.git /cctools-port \
   && wget $OSX_SDK_PATH \
   && mv `basename $OSX_SDK_PATH` /osxcross/tarballs/ \
#   && cd /apple-libtapi \
#   && INSTALLPREFIX=/osxcross/target ./build.sh \
#   && cd /cctools-port/cctools \
#   && ./configure --prefix=/osxcross/target --target=x86_64-apple-darwin16 --with-libtapi=/osxcross/target \
#   && make \
#   && make install \
   &&  OCDEBUG=1 UNATTENDED=yes OSX_VERSION_MIN=10.6 /osxcross/build.sh \
   &&  mv /osxcross/target $OSX_NDK_X86 \
   &&  rm -rf /osxcross
VOLUME ["/root/.config"]
