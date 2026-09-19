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

* **`--kvm-gpu`**: Enable experimental NVIDIA GPU support in minikube
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
* Network connectivity lost after installing Docker [#23589](https://github.com/kubernetes/minikube/issues/23589)

Also see [co/kvm2-driver open issues](https://github.com/kubernetes/minikube/labels/co%2Fkvm2-driver).

### Nested Virtualization

If you are running KVM in a nested virtualization environment ensure your config the kernel modules correctly follow either [this](https://stafwag.github.io/blog/blog/2018/06/04/nested-virtualization-in-kvm/) or [this](https://computingforgeeks.com/how-to-install-kvm-virtualization-on-debian/) tutorial.

## Troubleshooting

* Run `id` to confirm that user belongs to the libvirt[d] group (the output should contain entry similar to: 'groups=...,108(libvirt),...').
* Run `virsh domcapabilities --virttype="kvm"` to confirm that the host supports KVM virtualisation.
* Run `virt-host-validate` and check for the suggestions.
* Run ``ls -la `which virsh` ``, `virsh uri`, `sudo virsh net-list --all` and `ip a s` to collect additional information for debugging.
* Run `minikube start --alsologtostderr -v=9` to debug crashes.
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

### Network connectivity lost after installing Docker

#### Symptoms

When starting a minikube cluster with the `kvm2` driver on a host with Docker installed, minikube may display a connectivity warning:

```shell
❗  Failing to connect to https://registry.k8s.io/ from inside the minikube VM
💡  To pull new external images, you may need to configure a proxy: https://minikube.sigs.k8s.io/docs/reference/networking/proxy/
```

Inside the VM, external network connectivity and DNS lookups fail:

```shell
$ minikube ssh -- ping -c 3 8.8.8.8
PING 8.8.8.8 (8.8.8.8) 56(84) bytes of data.
--- 8.8.8.8 ping statistics ---
3 packets transmitted, 0 received, 100% packet loss, time 2048ms

$ minikube ssh -- nslookup kubernetes.io
;; connection timed out; no servers could be reached
```

#### Root Cause

When the Docker daemon starts on the host, it modifies the host firewall by setting the default policy of the `FORWARD` chain to `DROP`. Because minikube's `kvm2` driver relies on libvirt virtual bridges (such as `virbr0` and `virbr1`), traffic forwarded across these bridges is blocked by Docker's default drop policy unless explicitly permitted.

#### How to Detect

Check the host's `FORWARD` and `DOCKER-USER` firewall chains:

```shell
$ sudo iptables -S FORWARD
-P FORWARD DROP
-A FORWARD -j DOCKER-USER
-A FORWARD -j DOCKER-ISOLATION-STAGE-1
...

$ sudo iptables -S DOCKER-USER
-N DOCKER-USER
-A DOCKER-USER -j RETURN
```

If the `FORWARD` policy is `-P FORWARD DROP` and `DOCKER-USER` only contains `-j RETURN` without explicit rules allowing traffic on `virbr+` interfaces, Docker is blocking forwarded packets from the VM.

#### Immediate Fix

To allow forwarded traffic across libvirt bridges, add rules to Docker's custom `DOCKER-USER` chain (which Docker evaluates before its own rules):

```shell
sudo iptables -I DOCKER-USER -i virbr+ -j ACCEPT
sudo iptables -I DOCKER-USER -o virbr+ -j ACCEPT
```

{{% alert title="Note" color="primary" %}}
The `virbr+` wildcard interface syntax matches any libvirt bridge (such as `virbr0`, `virbr1`, etc.). On modern Linux distributions using `nftables`, the `iptables` command transparently interacts with `iptables-nft`.
{{% /alert %}}

#### Persistent Fix (Libvirt Network Hook)

Because Docker recreates its chains upon restart, you can persist these rules across reboots and libvirt network state changes by creating a libvirt network hook script.

Create `/etc/libvirt/hooks/network` with root privileges:

```bash
#!/bin/bash
if [ "${2}" = "started" ] || [ "${2}" = "reloaded" ]; then
    iptables -I DOCKER-USER -i virbr+ -j ACCEPT
    iptables -I DOCKER-USER -o virbr+ -j ACCEPT
elif [ "${2}" = "stopped" ]; then
    iptables -D DOCKER-USER -i virbr+ -j ACCEPT
    iptables -D DOCKER-USER -o virbr+ -j ACCEPT
fi
```

Make the script executable:

```shell
sudo chmod +x /etc/libvirt/hooks/network
```

#### Alternative: Firewalld Configuration

If your distribution uses `firewalld` (such as Fedora, RHEL, or CentOS), ensure that the libvirt bridge is assigned to the `libvirt` or `trusted` zone, and enable forwarding policies:

```shell
sudo firewall-cmd --zone=libvirt --add-interface=virbr0 --permanent
sudo firewall-cmd --reload
```

