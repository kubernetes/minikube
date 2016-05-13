<!--[metadata]>
+++
title = "Generic"
description = "Generic driver for machine"
keywords = ["machine, Generic, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Generic

Create machines using an existing VM/Host with SSH.

This is useful if you are using a provider that Machine does not support
directly or if you would like to import an existing host to allow Docker
Machine to manage.

The driver will perform a list of tasks on create:

-   If docker is not running on the host, it will be installed automatically.
-   It will update the host packages (`apt-get update`, `yum update`...).
-   It will generate certificates to secure the docker daemon.
-   The docker daemon will be restarted, thus all running containers will be stopped.
-   The hostname will be changed to fit the machine name.

### Example

To create a machine instance, specify `--driver generic`, the IP address or DNS
name of the host and the path to the SSH private key authorized to connect
to the host.

    $ docker-machine create \
      --driver generic \
      --generic-ip-address=203.0.113.81 \
      --generic-ssh-key=~/.ssh/id_rsa \
      vm

### Password-protected SSH keys

When an SSH identity is not provided (with the `--generic-ssh-key` flag),
the SSH agent (if running) will be consulted. This makes it possible to
easily use password-protected SSH keys.

Note that this usage is _only_ supported if you're using the external SSH client,
which is the default behaviour when the `ssh` binary is available. If you're
using the native client (with `--native-ssh`), using the SSH agent is not yet
supported.

    $ docker-machine create \
      --driver generic \
      --generic-ip-address=203.0.113.81 \
      other

### Sudo privileges

The user that is used to SSH into the host can be specified with
`--generic-ssh-user` flag. This user has to have password-less sudo
privileges.
If it's not the case, you need to edit the `sudoers` file and configure the user
as a sudoer with `NOPASSWD`. See https://help.ubuntu.com/community/Sudoers.

### Options

-   `--generic-engine-port`: Port to use for Docker Daemon (Note: This flag will not work with boot2docker).
-   `--generic-ip-address`: **required** IP Address of host.
-   `--generic-ssh-key`: Path to the SSH user private key.
-   `--generic-ssh-user`: SSH username used to connect.
-   `--generic-ssh-port`: Port to use for SSH.

> **Note**: You must use a base operating system supported by Machine.

Environment variables and default values:

| CLI option                 | Environment variable | Default                   |
| -------------------------- | -------------------- | ------------------------- |
| `--generic-engine-port`    | `GENERIC_ENGINE_PORT`| `2376`                    |
| **`--generic-ip-address`** | `GENERIC_IP_ADDRESS` | -                         |
| `--generic-ssh-key`        | `GENERIC_SSH_KEY`    | _(defers to `ssh-agent`)_ |
| `--generic-ssh-user`       | `GENERIC_SSH_USER`   | `root`                    |
| `--generic-ssh-port`       | `GENERIC_SSH_PORT`   | `22`                      |
