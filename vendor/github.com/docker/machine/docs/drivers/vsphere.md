<!--[metadata]>
+++
title = "VMware vSphere"
description = "VMware vSphere driver for machine"
keywords = ["machine, VMware vSphere, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# VMware vSphere

Creates machines on a [VMware vSphere](http://www.vmware.com/products/vsphere) Virtual Infrastructure. The machine must have a working vSphere ESXi installation. You can use a paid license or free 60 day trial license. Your installation may also include an optional VCenter server.
    $ docker-machine create --driver vmwarevsphere --vmwarevsphere-username=user --vmwarevsphere-password=SECRET vm

Options:

-   `--vmwarevsphere-username`: **required** vSphere Username.
-   `--vmwarevsphere-password`: **required** vSphere Password.
-   `--vmwarevsphere-cpu-count`: CPU number for Docker VM.
-   `--vmwarevsphere-memory-size`: Size of memory for Docker VM (in MB).
-   `--vmwarevsphere-disk-size`: Size of disk for Docker VM (in MB).
-   `--vmwarevsphere-boot2docker-url`: URL for boot2docker image.
-   `--vmwarevsphere-vcenter`: IP/hostname for vCenter (or ESXi if connecting directly to a single host).
-   `--vmwarevsphere-vcenter-port`: vSphere Port for vCenter.
-   `--vmwarevsphere-network`: Network where the Docker VM will be attached.
-   `--vmwarevsphere-datastore`: Datastore for Docker VM.
-   `--vmwarevsphere-datacenter`: Datacenter for Docker VM (must be set to `ha-datacenter` when connecting to a single host).
-   `--vmwarevsphere-pool`: Resource pool for Docker VM.
-   `--vmwarevsphere-hostsystem`: vSphere compute resource where the docker VM will be instantiated (use <cluster>/* or <cluster>/<host> if using a cluster).

The VMware vSphere driver uses the latest boot2docker image.

Environment variables and default values:

| CLI option                        | Environment variable      | Default                  |
| --------------------------------- | ------------------------- | ------------------------ |
| **`--vmwarevsphere-username`**    | `VSPHERE_USERNAME`        | -                        |
| **`--vmwarevsphere-password`**    | `VSPHERE_PASSWORD`        | -                        |
| `--vmwarevsphere-cpu-count`       | `VSPHERE_CPU_COUNT`       | `2`                      |
| `--vmwarevsphere-memory-size`     | `VSPHERE_MEMORY_SIZE`     | `2048`                   |
| `--vmwarevsphere-boot2docker-url` | `VSPHERE_BOOT2DOCKER_URL` | _Latest boot2docker url_ |
| `--vmwarevsphere-vcenter`         | `VSPHERE_VCENTER`         | -                        |
| `--vmwarevsphere-vcenter-port`    | `VSPHERE_VCENTER_PORT`    | 443                      |
| `--vmwarevsphere-disk-size`       | `VSPHERE_DISK_SIZE`       | `20000`                  |
| `--vmwarevsphere-network`         | `VSPHERE_NETWORK`         | -                        |
| `--vmwarevsphere-datastore`       | `VSPHERE_DATASTORE`       | -                        |
| `--vmwarevsphere-datacenter`      | `VSPHERE_DATACENTER`      | -                        |
| `--vmwarevsphere-pool`            | `VSPHERE_POOL`            | -                        |
| `--vmwarevsphere-hostsystem`      | `VSPHERE_HOSTSYSTEM`      | -                        |
