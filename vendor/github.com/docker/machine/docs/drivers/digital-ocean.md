<!--[metadata]>
+++
title = "Digital Ocean"
description = "Digital Ocean driver for machine"
keywords = ["machine, Digital Ocean, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Digital Ocean

Create Docker machines on [Digital Ocean](https://www.digitalocean.com/).

You need to create a personal access token under "Apps & API" in the Digital Ocean
Control Panel and pass that to `docker-machine create` with the `--digitalocean-access-token` option.

    $ docker-machine create --driver digitalocean --digitalocean-access-token=aa9399a2175a93b17b1c86c807e08d3fc4b79876545432a629602f61cf6ccd6b test-this

Options:

-   `--digitalocean-access-token`: **required** Your personal access token for the Digital Ocean API.
-   `--digitalocean-image`: The name of the Digital Ocean image to use.
-   `--digitalocean-region`: The region to create the droplet in, see [Regions API](https://developers.digitalocean.com/documentation/v2/#regions) for how to get a list.
-   `--digitalocean-size`: The size of the Digital Ocean droplet (larger than default options are of the form `2gb`).
-   `--digitalocean-ipv6`: Enable IPv6 support for the droplet.
-   `--digitalocean-private-networking`: Enable private networking support for the droplet.
-   `--digitalocean-backups`: Enable Digital Oceans backups for the droplet.
-   `--digitalocean-userdata`: Path to file containing User Data for the droplet.
-   `--digitalocean-ssh-user`: SSH username.
-   `--digitalocean-ssh-port`: SSH port.
-   `--digitalocean-ssh-key-fingerprint`: Use an existing SSH key instead of creating a new one, see [SSH keys](https://developers.digitalocean.com/documentation/v2/#ssh-keys).

The DigitalOcean driver will use `ubuntu-15-10-x64` as the default image.

Environment variables and default values:

| CLI option                          | Environment variable              | Default            |
| ----------------------------------- | --------------------------------- | ------------------ |
| **`--digitalocean-access-token`**   | `DIGITALOCEAN_ACCESS_TOKEN`       | -                  |
| `--digitalocean-image`              | `DIGITALOCEAN_IMAGE`              | `ubuntu-15-10-x64` |
| `--digitalocean-region`             | `DIGITALOCEAN_REGION`             | `nyc3`             |
| `--digitalocean-size`               | `DIGITALOCEAN_SIZE`               | `512mb`            |
| `--digitalocean-ipv6`               | `DIGITALOCEAN_IPV6`               | `false`            |
| `--digitalocean-private-networking` | `DIGITALOCEAN_PRIVATE_NETWORKING` | `false`            |
| `--digitalocean-backups`            | `DIGITALOCEAN_BACKUPS`            | `false`            |
| `--digitalocean-userdata`           | `DIGITALOCEAN_USERDATA`           | -                  |
| `--digitalocean-ssh-user`           | `DIGITALOCEAN_SSH_USER`           | `root`             |
| `--digitalocean-ssh-port`           | `DIGITALOCEAN_SSH_PORT`           | 22                 |
| `--digitalocean-ssh-key-fingerprint`| `DIGITALOCEAN_SSH_KEY_FINGERPRINT`| -                  |
