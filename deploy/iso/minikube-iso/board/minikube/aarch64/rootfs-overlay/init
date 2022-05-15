#!/bin/sh
mkdir /sysroot
# the value 90% borrowed from tcl via boot2docker
mount -t tmpfs -o size=90% tmpfs /sysroot
# copy from rootfs, to be able to do switch_root(8)
tar -C / --exclude=sysroot -cf - . | tar -C /sysroot/ -xf -

# devtmpfs does not get automounted for initramfs
/bin/mount -t devtmpfs devtmpfs /sysroot/dev
exec 0</sysroot/dev/console
exec 1>/sysroot/dev/console
exec 2>/sysroot/dev/console
exec /sbin/switch_root /sysroot /sbin/init "$@"
