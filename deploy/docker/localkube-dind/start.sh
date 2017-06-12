#!/bin/bash

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

mount --make-shared /

export CNI_BRIDGE_NETWORK_OFFSET="0.0.1.0"
/dindnet &> /dev/null &



mkdir -p /etc/localkube
base=/etc/localkube/docker
/usr/bin/dockerd \
  --exec-root=$base.exec \
  --graph=$base.graph \
  --host=unix://$base.socket \
  --pidfile=$base.pid &> /dev/null &


mkdir -p /etc/kubernetes/manifests
/localkube start \
--apiserver-insecure-address=0.0.0.0 \
--apiserver-insecure-port=8080 \
--network-plugin=cni \
--extra-config=kubelet.DockerEndpoint=unix:///$base.socket \
--extra-config=kubelet.KubeletFlags.ContainerRuntimeOptions.CNIConfDir="/etc/cni/net.d" \
--extra-config=kubelet.KubeletFlags.ContainerRuntimeOptions.CNIBinDir="/opt/cni/bin" \
--extra-config=kubelet.ClusterDNS="10.96.0.10" \
--extra-config=kubelet.ClusterDomain="cluster.local" \
--extra-config=kubelet.AllowPrivileged="true"
