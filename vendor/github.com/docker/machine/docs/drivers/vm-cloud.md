<!--[metadata]>
+++
title = "VMware vCloud Air"
description = "VMware vCloud Air driver for machine"
keywords = ["machine, VMware vCloud Air, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# VMware vCloud Air

Creates machines on [vCloud Air](http://vcloud.vmware.com) subscription service. You need an account within an existing subscription of vCloud Air VPC or Dedicated Cloud.

    $ docker-machine create --driver vmwarevcloudair --vmwarevcloudair-username=user --vmwarevcloudair-password=SECRET vm

Options:

-   `--vmwarevcloudair-username`: **required** vCloud Air Username.
-   `--vmwarevcloudair-password`: **required** vCloud Air Password.
-   `--vmwarevcloudair-computeid`: Compute ID (if using Dedicated Cloud).
-   `--vmwarevcloudair-vdcid`: Virtual Data Center ID.
-   `--vmwarevcloudair-orgvdcnetwork`: Organization VDC Network to attach.
-   `--vmwarevcloudair-edgegateway`: Organization Edge Gateway.
-   `--vmwarevcloudair-publicip`: Org Public IP to use.
-   `--vmwarevcloudair-catalog`: Catalog.
-   `--vmwarevcloudair-catalogitem`: Catalog Item.
-   `--vmwarevcloudair-provision`: Install Docker binaries.
-   `--vmwarevcloudair-cpu-count`: VM CPU Count.
-   `--vmwarevcloudair-memory-size`: VM Memory Size in MB.
-   `--vmwarevcloudair-ssh-port`: SSH port.
-   `--vmwarevcloudair-docker-port`: Docker port.

The VMware vCloud Air driver will use the `Ubuntu Server 12.04 LTS (amd64 20140927)` image by default.

Environment variables and default values:

| CLI option                        | Environment variable      | Default                                    |
| --------------------------------- | ------------------------- | ------------------------------------------ |
| **`--vmwarevcloudair-username`**  | `VCLOUDAIR_USERNAME`      | -                                          |
| **`--vmwarevcloudair-password`**  | `VCLOUDAIR_PASSWORD`      | -                                          |
| `--vmwarevcloudair-computeid`     | `VCLOUDAIR_COMPUTEID`     | -                                          |
| `--vmwarevcloudair-vdcid`         | `VCLOUDAIR_VDCID`         | -                                          |
| `--vmwarevcloudair-orgvdcnetwork` | `VCLOUDAIR_ORGVDCNETWORK` | `<vdcid>-default-routed`                   |
| `--vmwarevcloudair-edgegateway`   | `VCLOUDAIR_EDGEGATEWAY`   | `<vdcid>`                                  |
| `--vmwarevcloudair-publicip`      | `VCLOUDAIR_PUBLICIP`      | -                                          |
| `--vmwarevcloudair-catalog`       | `VCLOUDAIR_CATALOG`       | `Public Catalog`                           |
| `--vmwarevcloudair-catalogitem`   | `VCLOUDAIR_CATALOGITEM`   | `Ubuntu Server 12.04 LTS (amd64 20140927)` |
| `--vmwarevcloudair-provision`     | `VCLOUDAIR_PROVISION`     | `true`                                     |
| `--vmwarevcloudair-cpu-count`     | `VCLOUDAIR_CPU_COUNT`     | `1`                                        |
| `--vmwarevcloudair-memory-size`   | `VCLOUDAIR_MEMORY_SIZE`   | `2048`                                     |
| `--vmwarevcloudair-ssh-port`      | `VCLOUDAIR_SSH_PORT`      | `22`                                       |
| `--vmwarevcloudair-docker-port`   | `VCLOUDAIR_DOCKER_PORT`   | `2376`                                     |
