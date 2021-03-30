---
title: "kvm2"
weight: 2
description: >
  Linux KVM (Kernel-based Virtual Machine) driver
aliases:
    - /docs/reference/drivers/kvm2
---


## Overview

[KVM (Kernel-based Virtual Machine)](https://www.linux-kvm.org/page/Main_Page) is a full virtualization solution for Linux on x86 hardware containing virtualization extensions. To work with KVM, minikube uses the [libvirt virtualization API](https://libvirt.org/)

{{% readfile file="/docs/drivers/includes/kvm2_usage.inc" %}}

## Check virtualization support

{{% readfile file="/docs/drivers/includes/check_virtualization_linux.inc" %}}

## Special features

The `minikube start` command supports 5 additional KVM specific flags:

* **`--gpu`**: Enable experimental NVIDIA GPU support in minikube
* **`--hidden`**: Hide the hypervisor signature from the guest in minikube
* **`--kvm-network`**:  The KVM default network name
* **`--network`**:  The dedicated KVM private network name
* **`--kvm-qemu-uri`**: The KVM qemu uri, defaults to qemu:///system

## Issues

* `minikube` will repeatedly ask for the root password if user is not in the correct `libvirt` group [#3467](https://github.com/kubernetes/minikube/issues/3467)
* `Machine didn't return an IP after 120 seconds` when firewall prevents VM network access [#3566](https://github.com/kubernetes/minikube/issues/3566)
* `unable to set user and group to '65534:992` when `dynamic ownership = 1` in `qemu.conf` [#4467](https://github.com/kubernetes/minikube/issues/4467)
* KVM VM's cannot be used simultaneously with VirtualBox  [#4913](https://github.com/kubernetes/minikube/issues/4913)
* On some distributions, libvirt bridge networking may fail until the host reboots

Also see [co/kvm2 open issues](https://github.com/kubernetes/minikube/labels/co%2Fkvm2)

### Nested Virtulization

If you are running KVM in a nested virtualization environment ensure your config the kernel modules correctly follow either [this](https://stafwag.github.io/blog/blog/2018/06/04/nested-virtualization-in-kvm/) or [this](https://computingforgeeks.com/how-to-install-kvm-virtualization-on-debian/) tutorial.

## Troubleshooting

* Run `id` to confirm that user belongs to the libvirt[d] group (the output should contain entry similar to: 'groups=...,108(libvirt),...').
* Run `virsh domcapabilities --virttype="kvm"` to confirm that the host supports KVM virtualisation.
* Run `virt-host-validate` and check for the suggestions.
* Run ``ls -la `which virsh` ``, `virsh uri`, `sudo virsh net-list --all` and `ip a s` to collect additional information for debugging.
* Run `minikube start --alsologtostderr -v=9` to debug crashes.
* Run `docker-machine-driver-kvm2 version` to verify the kvm2 driver executes properly.
* Read [How to debug Virtualization problems](https://fedoraproject.org/wiki/How_to_debug_Virtualization_problems)

### Troubleshooting KVM/libvirt networks

For the most part, minikube will try to detect and resolve any issues with the KVM/libvirt networks for you.
However, there are some situations where manual intervention is needed, mostly because root privileges are required.

1.  Run `sudo virsh net-list --all` to list all interfaces.

example output:
```shell
 Name                     State    Autostart   Persistent
-----------------------------------------------------------
 default                  active   yes         yes
 mk-kvm0                  active   yes         yes
 mk-minikube              active   yes         yes
 my-custom-kvm-priv-net   active   yes         yes
```
where:
*  ***default*** is the default libvirt network,
*  ***mk-kvm0*** is a default libvirt network created for minikube ***kvm0*** profile (eg, using `minikube start -p kvm0 --driver=kvm2`),
*  ***mk-minikube*** is a network created for default minikube profile (eg, using `minikube start --driver=kvm2`) and
*  ***my-custom-kvm-priv-net*** is a custom private network name provided for minikube profile (eg, using `minikube start -p kvm1 --driver=kvm2 --network="my-custom-kvm-priv-net"`).

2.  Run `sudo virsh net-autostart <network>` to manually set **network** to autostart, if not already set.

3.  Run `sudo virsh net-start <network>` to manually start/activate **network**, if not already started/active.

    1.  In case that the ***default*** libvirt network is missing or is unable to start/activate - consult your OS/distro-specific libvirt docs; the following steps *might* help you to fix the issue:
        1.  Run `sudo virsh net-dumpxml default > default.xml` to backup the ***default*** libvirt network config.
        2.  Run `sudo virsh net-destroy default` to stop the ***default*** libvirt network.
        3.  Run `sudo virsh net-undefine default` to delete the ***default*** libvirt network.
        4.  Run `sudo virsh net-define /usr/share/libvirt/networks/default.xml` to recreate the ***default*** libvirt network.
            *  Note: repeat above steps ***b.*** and ***c.*** and then Run `sudo virsh net-define default.xml` to restore the original ***default*** libvirt network config, in case of any issue.
        5.  Run `sudo virsh net-start default` to start the ***default*** libvirt network.
        6.  Run `sudo virsh net-autostart default` to autostart the ***default*** libvirt network.

    2.  If ***non-default*** libvirt **network** is unable to start/activate, use the following steps:
        1.  Run `sudo virsh net-dumpxml <network>` to dump XML **network** config - note the `bridge name=<bridge>` and `ip address='<address>' netmask='<netmask>'` values. Example output:

        ```xml
        <network connections='1'>
          <name>mk-minikube</name>
          <uuid>cfcb37fb-fd75-4599-825a-14bee5d863f5</uuid>
          <bridge name='virbr1' stp='on' delay='0'/>
          <mac address='52:54:00:80:97:5a'/>
          <dns enable='no'/>
          <ip address='192.168.39.1' netmask='255.255.255.0'>
            <dhcp>
              <range start='192.168.39.2' end='192.168.39.254'/>
            </dhcp>
          </ip>
        </network>
        ```

        b.  Run `ip -4 -br -o a s` to show all interfaces with assigned IPs (in CIDR format), now compare the above IP **address** and **netmask** with those of the **bridge**. Example output:

        ```shell
        lo               UNKNOWN        127.0.0.1/8
        virbr0           UP             192.168.122.1/24
        wlp113s0         UP             192.168.42.17/24
        br-08ada8d5dfa4  DOWN           172.22.0.1/16
        docker0          DOWN           172.17.0.1/16
        virbr1           UP             192.168.39.1/24
        ```

        *  ***IF THEY MATCH, or THE IP ADDRESS ISN'T LISTED ANYWHERE***: Run `sudo ip link delete <bridge>` followed by `sudo virsh net-start <network>` and  `sudo virsh net-autostart <network>` to let libvirt recreate the **bridge** and [auto]start the **network**.
        *  ***IF THE IP ADDRESS BELONGS TO ANOTHER INTERFACE***: something else occupied the IP **address** creating the conflict, and you'll have to determine what and then choose between the two...

4.  Run `sudo systemctl restart libvirtd` or `sudo systemctl restart libvirt` (depending on your OS/distro) to restart the libvirt daemon.

Hopefully, by now you have libvirt network operational, and you will be successfully running minikube again.
