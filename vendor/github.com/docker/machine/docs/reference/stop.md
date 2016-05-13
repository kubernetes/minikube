<!--[metadata]>
+++
title = "stop"
description = "Gracefully stop a machine"
keywords = ["machine, stop, subcommand"]
[menu.main]
identifier="machine.stop"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# stop

    Usage: docker-machine stop [arg...]

    Gracefully Stop a machine

    Description:
       Argument(s) are one or more machine names.

For example:

    $ docker-machine ls
    NAME   ACTIVE   DRIVER       STATE     URL
    dev    *        virtualbox   Running   tcp://192.168.99.104:2376
    $ docker-machine stop dev
    $ docker-machine ls
    NAME   ACTIVE   DRIVER       STATE     URL
    dev    *        virtualbox   Stopped
