<!--[metadata]>
+++
title = "restart"
description = "Restart a machine"
keywords = ["machine, restart, subcommand"]
[menu.main]
identifier="machine.restart"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# restart

    Usage: docker-machine restart [arg...]

    Restart a machine

    Description:
       Argument(s) are one or more machine names.
       
Restart a machine. Oftentimes this is equivalent to
`docker-machine stop; docker-machine start`. But some cloud driver try to implement a clever restart which keeps the same
ip address.

    $ docker-machine restart dev
    Waiting for VM to start...
