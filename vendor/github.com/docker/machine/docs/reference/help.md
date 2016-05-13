<!--[metadata]>
+++
title = "help"
description = "Show command help"
keywords = ["machine, help, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# help

    Usage: docker-machine help [arg...]

    Shows a list of commands or help for one command

Usage: docker-machine help _subcommand_

For example:

    $ docker-machine help config
    Usage: docker-machine config [OPTIONS] [arg...]

    Print the connection config for machine

    Description:
       Argument is a machine name.

    Options:

       --swarm      Display the Swarm config instead of the Docker daemon
