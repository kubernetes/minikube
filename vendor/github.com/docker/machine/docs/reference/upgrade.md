<!--[metadata]>
+++
title = "upgrade"
description = "Upgrade Docker on a machine"
keywords = ["machine, upgrade, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# upgrade

Upgrade a machine to the latest version of Docker. How this upgrade happens
depends on the underlying distribution used on the created instance.

For example, if the machine uses Ubuntu as the underlying operating system, it
will run a command similar to `sudo apt-get upgrade docker-engine`, because
Machine expects Ubuntu machines it manages to use this package. As another
example, if the machine uses boot2docker for its OS, this command will download
the latest boot2docker ISO and replace the machine's existing ISO with the
latest.

    $ docker-machine upgrade default
    Stopping machine to do the upgrade...
    Upgrading machine default...
    Downloading latest boot2docker release to /home/username/.docker/machine/cache/boot2docker.iso...
    Starting machine back up...
    Waiting for VM to start...

> **Note**: If you are using a custom boot2docker ISO specified using
> `--virtualbox-boot2docker-url` or an equivalent flag, running an upgrade on
> that machine will completely replace the specified ISO with the latest
> "vanilla" boot2docker ISO available.
