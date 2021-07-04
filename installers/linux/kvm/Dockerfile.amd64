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

FROM gcr.io/gcp-runtimes/ubuntu_16_0_4

RUN apt-get update && apt-get install -y --no-install-recommends \
		gcc \
		libc6-dev \
		make \
		pkg-config \
		curl \
		libvirt-dev \
		git \
	&& rm -rf /var/lib/apt/lists/*

ARG GO_VERSION

RUN curl -sSL https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz | tar -C /usr/local -xzf -

ENV GOPATH /go
ENV PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/go/bin
