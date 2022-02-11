#!/bin/sh

# Copyright 2022 The Kubernetes Authors All rights reserved.
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


#!/bin/sh

set -e

echo "2*** I am inside post-image-efi.sh"
pwd

# GENIMAGE_CFG="./genimage-efi.cfg"
# GENIMAGE_CFG="$2"

# echo "2a*** ${2}"
# echo "2b*** ${GENIMAGE_CFG}"

# support/scripts/genimage.sh -c "$GENIMAGE_CFG"

cd "$BINARIES_DIR"
echo "BINARIES_DIR: $BINARIES_DIR"
ls -lah
mkdir -p root/boot
cp bzImage root/boot/vmlinuz
cp rootfs.cpio.gz root/boot/initrd.img
mkdir -p root/EFI/BOOT
cp efi-part/EFI/BOOT/* root/EFI/BOOT/
# cp efiboot.img root/EFI/BOOT/

echo "2c*** mkisofs"
ls -lah

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

# -eltorito-boot EFI/BOOT/efiboot.img \

# make the new iso and put in root.
#  mkisofs
#    -o /new.iso
#    -b isolinux/isolinux.bin \
#    -c isolinux/boot.cat
#    -no-emul-boot
#    -boot-load-size 4 \
#    -boot-info-table 
#    -J -R -V disks .

# set -e

# UUID=$(dumpe2fs "$BINARIES_DIR/rootfs.ext2" 2>/dev/null | sed -n 's/^Filesystem UUID: *\(.*\)/\1/p')
# sed -i "s/UUID_TMP/$UUID/g" "$BINARIES_DIR/efi-part/EFI/BOOT/grub.cfg"
# sed "s/UUID_TMP/$UUID/g" board/pc/genimage-efi.cfg > "$BINARIES_DIR/genimage-efi.cfg"
# support/scripts/genimage.sh -c "$BINARIES_DIR/genimage-efi.cfg"
