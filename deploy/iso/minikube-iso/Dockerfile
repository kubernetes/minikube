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

# Beware ubuntu:16.10 and debian:testing have been tested and failed
FROM ubuntu:16.04

RUN apt-get update && apt-get install -y \
		build-essential \
		git \
		wget \
		cpio \
		python \
		unzip \
		bc \
		gcc-multilib \
		automake \
		libtool \
		gnupg2 \
		p7zip-full \
		locales \
		&& rm -rf /var/lib/apt/lists/*

RUN localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8
ENV LANG en_US.utf8

# dumb init will allow us to interrupt the build with ^C
RUN wget -O - https://github.com/Yelp/dumb-init/archive/v1.2.0.tar.gz | tar xvzf - \
	&& make -C dumb-init-1.2.0 \
	&& mv dumb-init-1.2.0/dumb-init /usr/bin \
	&& rm -rf dumb-init-1.2.0

RUN wget -O - https://redirector.gvt1.com/edgedl/go/go1.9.2.linux-amd64.tar.gz | tar -C /usr/local -xvz
ENV PATH $PATH:/usr/local/go/bin

ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["/bin/bash"]
