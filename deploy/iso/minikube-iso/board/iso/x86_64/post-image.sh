#!/bin/sh

set -e

GENIMAGE_CFG="$2"

support/scripts/genimage.sh -c "$GENIMAGE_CFG"

cd "$BINARIES_DIR"
mkdir -p root/boot
cp bzImage root/boot/vmlinuz
cp rootfs.cpio.gz root/boot/initrd.img
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