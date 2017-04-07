# Minikube ISO Release Notes

## Version 0.18.0 - 4/6/2017
* ISO will now be versioned the same as minikube
* Added timezone data
* Added `jq` and `coreutils` packages
* Enabled RDB Kernel module
* Added dockerized build for iso image
* Enabled NFS_v4_2 in kernel
* Added CIFS-utils

## Version 1.0.7 - 3/2/2017
* Added ethtool
* Added bootlocal.sh script for custom startup options
* Added version info in /etc/VERSION
* Bumped rkt to v1.24.0
* Enabled user namespaces in kernel
* `/tmp/hostpath_pv` and `/tmp/hostpath-provisioner` are now persisted

## Version 1.0.6 - 2/2/2017
* Added back in curl, git, and rsync
* Enabled CONFIG_TUN in kernel
* Added NFS packages
* Enabled swapon on start/stop
* Updated CNI to v0.4.0
* Fix permissions for /data directory
* Updated RKT to v1.23.0
* Added in CoreOS toolbox binary
* Fixed vboxFS permission error
