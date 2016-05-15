#!/bin/bash
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

# Get ethotool
docker build -t ethtool -f Dockerfile.ethtool .
docker run ethtool cat ethtool > $tmpdir/ethtool
chmod +x $tmpdir/ethtool

# Do the build.
docker build -t iso .

# Output the iso.
docker run iso > $ISO

popd
mv $tmpdir/$ISO .

# Clean up.
rm -rf $tmpdir

echo "Iso available at ./$ISO"
