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

FROM ubuntu:24.04

RUN apt-get update \
	&& apt-get install -y apt dpkg apt-utils ca-certificates software-properties-common \
	&& apt-get upgrade -y \
	&& if [ "$(uname -ms)" = "Linux x86_64" ]; then apt-get install -y gcc-multilib; fi \
	&& apt-get install -y \
		build-essential \
		git \
		wget \
		cpio \
		python3 \
		unzip \
		bc \
		gcc \
		automake \
		libtool \
		gnupg2 \
		p7zip-full \
		locales \
		rsync \
		cmake \
		dumb-init \
		libpcre3-dev \
		mkisofs \
	&& rm -rf /var/lib/apt/lists/*

RUN mkdir /app
RUN chmod 777 /app

RUN localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG=en_US.utf8

# the default ccache dir is "$HOME/.buildroot-ccache"
ENV BR2_CCACHE_DIR=/var/cache/buildroot/ccache
RUN mkdir -p /var/cache/buildroot/ccache && chmod 1777 /var/cache/buildroot/ccache
# the default gocache dir is "$HOME/.cache/go-build"
ENV GOCACHE=/var/cache/buildroot/gocache
RUN mkdir -p /var/cache/buildroot/gocache && chmod 1777 /var/cache/buildroot/gocache
VOLUME ["/var/cache/buildroot"]

# dumb init will allow us to interrupt the build with ^C
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/bash"]
