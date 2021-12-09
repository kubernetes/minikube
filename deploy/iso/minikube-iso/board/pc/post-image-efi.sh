#!/bin/sh

set -e

UUID=$(dumpe2fs "$BINARIES_DIR/rootfs.ext2" 2>/dev/null | sed -n 's/^Filesystem UUID: *\(.*\)/\1/p')

sed -i "s/UUID_TMP/$UUID/g" "$BINARIES_DIR/efi-part/EFI/BOOT/grub.cfg"

sed "s/UUID_TMP/$UUID/g" board/pc/genimage-efi.cfg > "$BINARIES_DIR/genimage-efi.cfg"

support/scripts/genimage.sh -c "$BINARIES_DIR/genimage-efi.cfg"