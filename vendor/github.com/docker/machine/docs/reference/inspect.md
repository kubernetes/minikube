<!--[metadata]>
+++
title = "inspect"
description = "Inspect information about a machine"
keywords = ["machine, inspect, subcommand"]
[menu.main]
identifier="machine.inspect"
parent="smn_machine_subcmds"
+++
<![end-metadata]-->

# inspect

    Usage: docker-machine inspect [OPTIONS] [arg...]

    Inspect information about a machine

    Description:
       Argument is a machine name.

    Options:
       --format, -f 	Format the output using the given go template.

By default, this will render information about a machine as JSON. If a format is
specified, the given template will be executed for each result.

Go's [text/template](http://golang.org/pkg/text/template/) package
describes all the details of the format.

In addition to the `text/template` syntax, there are some additional functions,
`json` and `prettyjson`, which can be used to format the output as JSON (documented below).

## Examples

**List all the details of a machine:**

This is the default usage of `inspect`.

    $ docker-machine inspect dev
    {
        "DriverName": "virtualbox",
        "Driver": {
            "MachineName": "docker-host-128be8d287b2028316c0ad5714b90bcfc11f998056f2f790f7c1f43f3d1e6eda",
            "SSHPort": 55834,
            "Memory": 1024,
            "DiskSize": 20000,
            "Boot2DockerURL": "",
            "IPAddress": "192.168.5.99"
        },
        ...
    }

**Get a machine's IP address:**

For the most part, you can pick out any field from the JSON in a fairly
straightforward manner.

    $ docker-machine inspect --format='{{.Driver.IPAddress}}' dev
    192.168.5.99

**Formatting details:**

If you want a subset of information formatted as JSON, you can use the `json`
function in the template.

    $ docker-machine inspect --format='{{json .Driver}}' dev-fusion
    {"Boot2DockerURL":"","CPUS":8,"CPUs":8,"CaCertPath":"/Users/hairyhenderson/.docker/machine/certs/ca.pem","DiskSize":20000,"IPAddress":"172.16.62.129","ISO":"/Users/hairyhenderson/.docker/machine/machines/dev-fusion/boot2docker-1.5.0-GH747.iso","MachineName":"dev-fusion","Memory":1024,"PrivateKeyPath":"/Users/hairyhenderson/.docker/machine/certs/ca-key.pem","SSHPort":22,"SSHUser":"docker","SwarmDiscovery":"","SwarmHost":"tcp://0.0.0.0:3376","SwarmMaster":false}

While this is usable, it's not very human-readable. For this reason, there is
`prettyjson`:

    $ docker-machine inspect --format='{{prettyjson .Driver}}' dev-fusion
    {
        "Boot2DockerURL": "",
        "CPUS": 8,
        "CPUs": 8,
        "CaCertPath": "/Users/hairyhenderson/.docker/machine/certs/ca.pem",
        "DiskSize": 20000,
        "IPAddress": "172.16.62.129",
        "ISO": "/Users/hairyhenderson/.docker/machine/machines/dev-fusion/boot2docker-1.5.0-GH747.iso",
        "MachineName": "dev-fusion",
        "Memory": 1024,
        "PrivateKeyPath": "/Users/hairyhenderson/.docker/machine/certs/ca-key.pem",
        "SSHPort": 22,
        "SSHUser": "docker",
        "SwarmDiscovery": "",
        "SwarmHost": "tcp://0.0.0.0:3376",
        "SwarmMaster": false
    }
