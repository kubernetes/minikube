<!--[metadata]>
+++
title = "IBM Softlayer"
description = "IBM Softlayer driver for machine"
keywords = ["machine, IBM Softlayer, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# IBM Softlayer

Create machines on [Softlayer](http://softlayer.com).

You need to generate an API key in the softlayer control panel.
[Retrieve your API key](http://knowledgelayer.softlayer.com/procedure/retrieve-your-api-key)

    $ docker-machine create --driver softlayer --softlayer-user=user --softlayer-api-key=KEY --softlayer-domain=domain vm

Options:

-   `--softlayer-memory`: Memory for host in MB.
-   `--softlayer-disk-size`: A value of `0` will set the SoftLayer default.
-   `--softlayer-user`: **required** Username for your SoftLayer account, api key needs to match this user.
-   `--softlayer-api-key`: **required** API key for your user account.
-   `--softlayer-region`: SoftLayer region.
-   `--softlayer-cpu`: Number of CPUs for the machine.
-   `--softlayer-hostname`: Hostname for the machine.
-   `--softlayer-domain`: **required** Domain name for the machine.
-   `--softlayer-api-endpoint`: Change SoftLayer API endpoint.
-   `--softlayer-hourly-billing`: Specifies that hourly billing should be used, otherwise monthly billing is used.
-   `--softlayer-local-disk`: Use local machine disk instead of SoftLayer SAN.
-   `--softlayer-private-net-only`: Disable public networking.
-   `--softlayer-image`: OS Image to use.
-   `--softlayer-public-vlan-id`: Your public VLAN ID.
-   `--softlayer-private-vlan-id`: Your private VLAN ID.

The SoftLayer driver will use `UBUNTU_LATEST` as the image type by default.

Environment variables and default values:

| CLI option                     | Environment variable        | Default                     |
| ------------------------------ | --------------------------- | --------------------------- |
| `--softlayer-memory`           | `SOFTLAYER_MEMORY`          | `1024`                      |
| `--softlayer-disk-size`        | `SOFTLAYER_DISK_SIZE`       | `0`                         |
| **`--softlayer-user`**         | `SOFTLAYER_USER`            | -                           |
| **`--softlayer-api-key`**      | `SOFTLAYER_API_KEY`         | -                           |
| `--softlayer-region`           | `SOFTLAYER_REGION`          | `dal01`                     |
| `--softlayer-cpu`              | `SOFTLAYER_CPU`             | `1`                         |
| `--softlayer-hostname`         | `SOFTLAYER_HOSTNAME`        | `docker`                    |
| **`--softlayer-domain`**       | `SOFTLAYER_DOMAIN`          | -                           |
| `--softlayer-api-endpoint`     | `SOFTLAYER_API_ENDPOINT`    | `api.softlayer.com/rest/v3` |
| `--softlayer-hourly-billing`   | `SOFTLAYER_HOURLY_BILLING`  | `false`                     |
| `--softlayer-local-disk`       | `SOFTLAYER_LOCAL_DISK`      | `false`                     |
| `--softlayer-private-net-only` | `SOFTLAYER_PRIVATE_NET`     | `false`                     |
| `--softlayer-image`            | `SOFTLAYER_IMAGE`           | `UBUNTU_LATEST`             |
| `--softlayer-public-vlan-id`   | `SOFTLAYER_PUBLIC_VLAN_ID`  | `0`                         |
| `--softlayer-private-vlan-id`  | `SOFTLAYER_PRIVATE_VLAN_ID` | `0`                         |
