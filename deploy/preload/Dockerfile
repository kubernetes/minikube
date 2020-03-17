# Copyright 2020 The Kubernetes Authors All rights reserved.
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

# Download gcloud
FROM alpine:3.10 as download-gcloud
ENV GCLOUD_VERSION 276.0.0
ENV GCLOUD_URL https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz
RUN wget -O gcloud.tar.gz "${GCLOUD_URL}"
RUN tar -zxf gcloud.tar.gz

FROM gcr.io/gcp-runtimes/ubuntu_16_0_4

# Install python and make
RUN apt-get update && \
    apt-get install --no-install-recommends --no-install-suggests -y \
    python make && \
    rm -rf /var/lib/apt/lists/*

# Install docker
COPY --from=docker:18.09.6 /usr/local/bin/docker /usr/local/bin/

# Install gcloud dependencies
COPY --from=download-gcloud google-cloud-sdk/ /google-cloud-sdk/
# Finish installation of gcloud
RUN CLOUDSDK_PYTHON="python2.7" /google-cloud-sdk/install.sh \
    --usage-reporting=false \
    --bash-completion=false \
    --disable-installation-options
ENV PATH=$PATH:/google-cloud-sdk/bin
RUN gcloud auth configure-docker

# Install go
COPY --from=golang:1.14 /usr/local/go /usr/local/go
ENV PATH=$PATH:/usr/local/go/bin

CMD make upload-preloaded-images-tar
