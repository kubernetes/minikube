# docker-machine-kvm
KVM driver for docker-machine

This driver leverages the new [plugin architecture](https://github.com/docker/machine/issues/1626) being
developed for Docker Machine.

# Quick start instructions

* Install `libvirt` and `qemu-kvm` on your system (e.g., `sudo apt-get install libvirt-bin qemu-kvm`)
    * Add yourself to the `libvirtd` group (may vary by linux distro) so you don't need to sudo
* Install [docker-machine](https://github.com/docker/machine/releases)
* Go to the
  [releases](https://github.com/dhiltgen/docker-machine-kvm/releases)
  page and download the docker-machine-driver-kvm binary, putting it
  in your PATH.
* You can now create virtual machines using this driver with
  `docker-machine create -d kvm myengine0`.

# Dependencies

This driver leverages [libvirt](http://libvirt.org/) and the [libvirt-go
library](https://github.com/alexzorin/libvirt-go) to create and manage
KVM based virtual machines.  It has been tested with Ubuntu 12.04 through 15.04
and should work on most platforms with KVM/libvirt support.  If you run into
compatibility problems, please file an [issue](https://github.com/dhiltgen/docker-machine-kvm/issues).

Typically you'll run `docker-machine` as yourself, so you'll want to
follow your distro specific instructions on allowing libvirt access
from your account.  For most distro's, you accomplish this by adding
your account to the `libvirtd` group.


# Capabilities
* **boot2docker.iso** based images
* **Dual Network**
    * **eth1** - A host private network called **docker-machines** is automatically created to ensure we always have connectivity to the VMs.  The `docker-machine ip` command will always return this IP address which is only accessible from your local system.
    * **eth0** - You can specify any libvirt named network.  If you don't specify one, the "default" named network will be used.
        * If you have exotic networking topolgies (openvswitch, etc.), you can use `virsh edit mymachinename` after creation, modify the first network definition by hand, then reboot the VM for the changes to take effect.
        * Typically this would be your "public" network accessible from external systems
        * To retrieve the IP address of this network, you can run a command like the following:
        ```bash
        docker-machine ssh mymachinename "ip -one -4 addr show dev eth0|cut -f7 -d' '"
        ```

* **Other Tunables**
    * Virtual CPU count via --kvm-cpu-count
    * Disk size via --kvm-disk-size
    * RAM via --kvm-memory

