<!--[metadata]>
+++
title = "config"
description = "Show client configuration"
keywords = ["machine, config, subcommand"]
[menu.main]
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# config

    Usage: docker-machine config [OPTIONS] [arg...]

    Print the connection config for machine

    Description:
       Argument is a machine name.

    Options:

       --swarm      Display the Swarm config instead of the Docker daemon


For example: 

    $ docker-machine config dev
    --tlsverify
    --tlscacert="/Users/ehazlett/.docker/machines/dev/ca.pem"
    --tlscert="/Users/ehazlett/.docker/machines/dev/cert.pem"
    --tlskey="/Users/ehazlett/.docker/machines/dev/key.pem"
    -H tcp://192.168.99.103:2376
