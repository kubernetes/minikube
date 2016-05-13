<!--[metadata]>
+++
title = "Rackspace"
description = "Rackspace driver for machine"
keywords = ["machine, Rackspace, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Rackspace

Create machines on [Rackspace cloud](http://www.rackspace.com/cloud)

    $ docker-machine create --driver rackspace --rackspace-username=user --rackspace-api-key=KEY --rackspace-region=region vm

Options:

-   `--rackspace-username`: **required** Rackspace account username.
-   `--rackspace-api-key`: **required** Rackspace API key.
-   `--rackspace-region`: **required** Rackspace region name.
-   `--rackspace-endpoint-type`: Rackspace endpoint type (`adminURL`, `internalURL` or the default `publicURL`).
-   `--rackspace-image-id`: Rackspace image ID. Default: Ubuntu 15.10 (Wily Werewolf) (PVHVM).
-   `--rackspace-flavor-id`: Rackspace flavor ID. Default: General Purpose 1GB.
-   `--rackspace-ssh-user`: SSH user for the newly booted machine.
-   `--rackspace-ssh-port`: SSH port for the newly booted machine.
-   `--rackspace-docker-install`: Set if Docker has to be installed on the machine.

The Rackspace driver will use `59a3fadd-93e7-4674-886a-64883e17115f` (Ubuntu 15.10) by default.

Environment variables and default values:

| CLI option                   | Environment variable | Default                                |
| ---------------------------- | -------------------- | -------------------------------------- |
| **`--rackspace-username`**   | `OS_USERNAME`        | -                                      |
| **`--rackspace-api-key`**    | `OS_API_KEY`         | -                                      |
| **`--rackspace-region`**     | `OS_REGION_NAME`     | -                                      |
| `--rackspace-endpoint-type`  | `OS_ENDPOINT_TYPE`   | `publicURL`                            |
| `--rackspace-image-id`       | -                    | `59a3fadd-93e7-4674-886a-64883e17115f` |
| `--rackspace-flavor-id`      | `OS_FLAVOR_ID`       | `general1-1`                           |
| `--rackspace-ssh-user`       | -                    | `root`                                 |
| `--rackspace-ssh-port`       | -                    | `22`                                   |
| `--rackspace-docker-install` | -                    | `true`                                 |
