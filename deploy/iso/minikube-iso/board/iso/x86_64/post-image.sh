#!/bin/sh

# Copyright 2021 The Kubernetes Authors All rights reserved.
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

GENIMAGE_CFG="$2"

support/scripts/genimage.sh -c "$GENIMAGE_CFG"

cd "$BINARIES_DIR"
mkdir -p root/boot
cp bzImage root/boot/bzimage
cp rootfs.cpio.gz root/boot/initrd
mkdir -p root/EFI/BOOT
cp efi-part/EFI/BOOT/* root/EFI/BOOT/
cp efiboot.img root/EFI/BOOT/

mkisofs \
   -o boot.iso \
   -R -J -v -d -N \
   -hide-rr-moved \
   -no-emul-boot \
   -eltorito-platform=efi \
   -eltorito-boot EFI/BOOT/efiboot.img \
   -V "EFIBOOTISO" \
   -A "EFI Boot ISO" \
   root
cd -
