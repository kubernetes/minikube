<!--[metadata]>
+++
title = "rm"
description = "Remove a machine."
keywords = ["machine, rm, subcommand"]
[menu.main]
identifier="machine.rm"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# rm

Remove a machine. This will remove the local reference as well as delete it
on the cloud provider or virtualization management platform.

    $ docker-machine rm --help

    Usage: docker-machine rm [OPTIONS] [arg...]

    Remove a machine

    Description:
       Argument(s) are one or more machine names.

    Options:

       --force, -f	Remove local configuration even if machine cannot be removed, also implies an automatic yes (`-y`)
       -y		Assumes automatic yes to proceed with remove, without prompting further user confirmation

## Examples

    $ docker-machine ls
    NAME   ACTIVE   URL          STATE     URL                         SWARM   DOCKER   ERRORS
    bar    -        virtualbox   Running   tcp://192.168.99.101:2376           v1.9.1
    baz    -        virtualbox   Running   tcp://192.168.99.103:2376           v1.9.1
    foo    -        virtualbox   Running   tcp://192.168.99.100:2376           v1.9.1
    qix    -        virtualbox   Running   tcp://192.168.99.102:2376           v1.9.1


    $ docker-machine rm baz
    About to remove baz
    Are you sure? (y/n): y
    Successfully removed baz


    $ docker-machine ls
    NAME   ACTIVE   URL          STATE     URL                         SWARM   DOCKER   ERRORS
    bar    -        virtualbox   Running   tcp://192.168.99.101:2376           v1.9.1
    foo    -        virtualbox   Running   tcp://192.168.99.100:2376           v1.9.1
    qix    -        virtualbox   Running   tcp://192.168.99.102:2376           v1.9.1


    $ docker-machine rm bar qix
    About to remove bar, qix
    Are you sure? (y/n): y
    Successfully removed bar
    Successfully removed qix


    $ docker-machine ls
    NAME   ACTIVE   URL          STATE     URL                         SWARM   DOCKER   ERRORS
    foo    -        virtualbox   Running   tcp://192.168.99.100:2376           v1.9.1

    $ docker-machine rm -y foo
    About to remove foo
    Successfully removed foo
