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

set -e

echo "1*** post-build.sh"
pwd
ls -lah

BOARD_DIR=$(dirname "$0")

echo "${BOARD_DIR}"

# Detect boot strategy, EFI or BIOS
if [ -d "$BINARIES_DIR/efi-part/" ]; then
    echo "1a*** boot strategy: EFI"
    pwd
    echo "BINARIES_DIR: $BINARIES_DIR"
    cp -f "$BOARD_DIR/grub-efi.cfg" "$BINARIES_DIR/efi-part/EFI/BOOT/grub.cfg"

    echo "1b*** copy done?"
    pwd
    ls -lah $BINARIES_DIR/efi-part/EFI/BOOT/

else
    echo "1a*** boot strategy: BIOS"
    cp -f "$BOARD_DIR/grub-bios.cfg" "$TARGET_DIR/boot/grub/grub.cfg"

    # Copy grub 1st stage to binaries, required for genimage
    cp -f "$TARGET_DIR/lib/grub/i386-pc/boot.img" "$BINARIES_DIR"
fi

echo "1z*** done"
