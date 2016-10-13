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

set -e

ISO=minikube.iso
tmpdir=$(mktemp -d)
echo "Building in $tmpdir."
cp -r . $tmpdir/

pushd $tmpdir

# Get nsenter.
docker run --rm jpetazzo/nsenter cat /nsenter > $tmpdir/nsenter && chmod +x $tmpdir/nsenter

# Get socat
docker build -t socat -f Dockerfile.socat .
docker run socat cat socat > $tmpdir/socat
chmod +x $tmpdir/socat

# Get ethtool
docker build -t ethtool -f Dockerfile.ethtool .
docker run ethtool cat ethtool > $tmpdir/ethtool
chmod +x $tmpdir/ethtool

# Get conntrack
docker build -t conntrack -f Dockerfile.conntrack .
docker run conntrack cat conntrack > $tmpdir/conntrack
chmod +x $tmpdir/conntrack

# Do the build.
docker build -t iso .

# Output the iso.
docker run iso > $ISO

popd
mv $tmpdir/$ISO .

# Clean up.
rm -rf $tmpdir
openssl sha256 "${ISO}" | awk '{print $2}' > "${ISO}.sha256"

echo "Iso available at ./$ISO"
echo "SHA sum available at ./$ISO.sha256"

