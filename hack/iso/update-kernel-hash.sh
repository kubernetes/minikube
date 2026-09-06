#!/bin/bash

# Copyright 2026 The Kubernetes Authors All rights reserved.
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

# After iso-menuconfig-<arch> sets BR2_LINUX_KERNEL_CUSTOM_VERSION_VALUE,
# fetch the kernel.org sha256 for that tarball and replace the linux-*.tar.xz
# line in linux.hash if it is not already present. License hashes stay.
# Linux only (GNU sed -i), same as the ISO build.

set -eu -o pipefail

if [ $# -ne 1 ]; then
	echo "usage: $0 aarch64|x86_64" >&2
	exit 1
fi

arch=$1
case "$arch" in
aarch64 | x86_64) ;;
*)
	echo "usage: $0 aarch64|x86_64" >&2
	exit 1
	;;
esac

cd "$(dirname "$0")/../.."

defconfig=deploy/iso/minikube-iso/configs/minikube_${arch}_defconfig
hash_file=deploy/iso/minikube-iso/patches/linux/linux.hash
sha256sums_url=https://cdn.kernel.org/pub/linux/kernel/v6.x/sha256sums.asc

version=$(sed -n 's/^BR2_LINUX_KERNEL_CUSTOM_VERSION_VALUE="\(.*\)"/\1/p' "$defconfig")
if [ -z "$version" ]; then
	echo "no BR2_LINUX_KERNEL_CUSTOM_VERSION_VALUE in $defconfig" >&2
	exit 1
fi

tarball=linux-$version.tar.xz
tarball_re=$(echo "$tarball" | sed 's/\./\\./g')

if grep -q "$tarball_re$" "$hash_file"; then
	echo "$tarball already in $hash_file"
	exit 0
fi

# kernel.org lists "HASH  linux-VERSION.tar.xz"; Buildroot wants "sha256  HASH  file".
new_hash=$(curl -fsSL "$sha256sums_url" | awk "/$tarball_re$/")
if [ -z "$new_hash" ]; then
	echo "no sha256 for $tarball in $sha256sums_url" >&2
	exit 1
fi

sed -i "s|^sha256 .* linux-.*\.tar\.xz$|sha256  $new_hash|" "$hash_file"

if ! grep -q "$tarball_re$" "$hash_file"; then
	echo "failed to replace linux-*.tar.xz line in $hash_file" >&2
	exit 1
fi

echo "updated $hash_file:"
echo "sha256  $new_hash"
