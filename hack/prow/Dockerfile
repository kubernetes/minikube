From debian:stretch
MAINTAINER Don McCasland <donmccasland@google.com>
CMD ["bash"]
RUN apt-get update \
    && apt-get install -y pkg-config libvirt-clients libvirt-daemon libvirt-daemon-system libvirt-dev qemu-kvm gcc g++ libc6-dev wget make git patch file jq gnupg apt-utils clang bzip2 cmake
RUN wget https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.11.1.linux-amd64.tar.gz
ENV GOPATH=/go
ENV PATH=/go/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" &&  chmod -R 777 "$GOPATH"
WORKDIR /go
ENV PWD /go/src/k8s.io/minikube
ENV GOARM=7
ENV KUBE_DYNAMIC_CROSSPLATFORMS="armhf arm64 s390x ppc64el"
ENV KUBE_CROSSPLATFORMS="linux/386 linux/arm linux/arm64 linux/ppc64le linux/s390x darwin/amd64 darwin/386 windows/amd64 windows/386"
RUN for platform in ${KUBE_CROSSPLATFORMS}; do GOOS=${platform%/*} GOARCH=${platform##*/} go install std; done
RUN echo "deb http://archive.ubuntu.com/ubuntu xenial main universe" > /etc/apt/sources.list.d/cgocrosscompiling.list \
   &&  apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 40976EAF437D05B5 3B4FE6ACC0B21F32 \
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
