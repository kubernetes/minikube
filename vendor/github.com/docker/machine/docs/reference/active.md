<!--[metadata]>
+++
title = "active"
description = "Identify active machines"
keywords = ["machine, active, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# active

See which machine is "active" (a machine is considered active if the
`DOCKER_HOST` environment variable points to it).

    $ docker-machine ls
    NAME      ACTIVE   DRIVER         STATE     URL
    dev       -        virtualbox     Running   tcp://192.168.99.103:2376
    staging   *        digitalocean   Running   tcp://203.0.113.81:2376
    $ echo $DOCKER_HOST
    tcp://203.0.113.81:2376
    $ docker-machine active
    staging
