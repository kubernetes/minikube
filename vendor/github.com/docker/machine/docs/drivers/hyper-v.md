<!--[metadata]>
+++
title = "Microsoft Hyper-V"
description = "Microsoft Hyper-V driver for machine"
keywords = ["machine, Microsoft Hyper-V, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Microsoft Hyper-V

Creates a Boot2Docker virtual machine locally on your Windows machine
using Hyper-V. [See here](http://windows.microsoft.com/en-us/windows-8/hyper-v-run-virtual-machines)
for instructions to enable Hyper-V. You will need to use an
Administrator level account to create and manage Hyper-V machines.

> **Note**: You will need an existing virtual switch to use the
> driver. Hyper-V can share an external network interface (aka
> bridging), see [this blog](http://blogs.technet.com/b/canitpro/archive/2014/03/11/step-by-step-enabling-hyper-v-for-use-on-windows-8-1.aspx).
> If you would like to use NAT, create an internal network, and use
> [Internet Connection
> Sharing](http://www.packet6.com/allowing-windows-8-1-hyper-v-vm-to-work-with-wifi/).

    $ docker-machine create --driver hyperv vm

Options:

-   `--hyperv-boot2docker-url`: The URL of the boot2docker ISO.
-   `--hyperv-virtual-switch`: Name of the virtual switch to use.
-   `--hyperv-disk-size`: Size of disk for the host in MB.
-   `--hyperv-memory`: Size of memory for the host in MB.
-   `--hyperv-cpu-count`: Number of CPUs for the host.
-   `--hyperv-static-macaddress`: Hyper-V network adapter's static MAC address.
-   `--hyperv-vlan-id`: Hyper-V network adapter's VLAN ID if any.

Environment variables and default values:

| CLI option                   | Environment variable       | Default                  |
| ---------------------------- | -------------------------- | ------------------------ |
| `--hyperv-boot2docker-url`   | `HYPERV_BOOT2DOCKER_URL`   | _Latest boot2docker url_ |
| `--hyperv-virtual-switch`    | `HYPERV_VIRTUAL_SWITCH`    | _first found_            |
| `--hyperv-disk-size`         | `HYPERV_DISK_SIZE`         | `20000`                  |
| `--hyperv-memory`            | `HYPERV_MEMORY`            | `1024`                   |
| `--hyperv-cpu-count`         | `HYPERV_CPU_COUNT`         | `1`                      |
| `--hyperv-static-macaddress` | `HYPERV_STATIC_MACADDRESS` | _undefined_              |
| `--hyperv-cpu-count`         | `HYPERV_VLAN_ID`           | _undefined_              |
