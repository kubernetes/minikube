# Copyright 2016 The Kubernetes Authors.
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

# Touching this file will trigger a new image build in Prow Staging https://testgrid.k8s.io/minikube-images
# Last update Jan 16, 2026 3:19 PM
FROM nginx:1.29.4-alpine-slim

RUN apk update && apk upgrade && apk add bash
COPY rootfs/bin/boot /bin/boot
COPY rootfs/etc/nginx /etc/nginx
CMD ["/bin/boot"]
