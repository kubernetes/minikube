<!--[metadata]>
+++
title = "Machine concepts and help"
description = "Understand concepts for Docker Machine, including drivers, base OS, IP addresses, environment variables"
keywords = ["docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarefusion, vmwarevcloudair, vmwarevsphere, exoscale"]
[menu.main]
parent="workw_machine"
weight=-40
+++
<![end-metadata]-->


# Understand Machine concepts and get help

Docker Machine allows you to provision Docker machines in a variety of environments, including virtual machines that reside on your local system, on cloud providers, or on bare metal servers (physical computers). Docker Machine creates a Docker host, and you use the Docker Engine client as needed to build images and create containers on the host.

## Drivers for creating machines

To create a virtual machine, you supply Docker Machine with the name of the driver you want use. The driver determines where the virtual machine is created. For example, on a local Mac or Windows system, the driver is typically Oracle VirtualBox. For provisioning physical machines, a generic driver is provided. For cloud providers, Docker Machine supports drivers such as AWS, Microsoft Azure, Digital Ocean, and many more. The Docker Machine reference includes a complete [list of supported drivers](drivers/index.md).

## Default base operating systems for local and cloud hosts

Since Docker runs on Linux, each VM that Docker Machine provisions relies on a
base operating system. For convenience, there are default base operating
systems. For the Oracle Virtual Box driver, this base operating system is <a href="https://github.com/boot2docker/boot2docker" target="_blank">boot2docker</a>. For drivers used to connect to cloud providers, the base operating system is Ubuntu 12.04+. You can change this default when you create a machine. The Docker Machine reference includes a complete [list of
supported operating systems](drivers/os-base.md).

## IP addresses for Docker hosts

For each machine you create, the Docker host address is the IP address of the
Linux VM. This address is assigned by the `docker-machine create` subcommand.
You use the `docker-machine ls` command to list the machines you have created.
The `docker-machine ip <machine-name>` command returns a specific host's IP
address.

## Configuring CLI environment variables for a Docker host

Before you can run a `docker` command on a machine, you need to configure your
command-line to point to that machine. The `docker-machine env <machine-name>`
subcommand outputs the configuration command you should use.

For a complete list of `docker-machine` subcommands, see the [Docker Machine subcommand reference](reference/index.md).

## Crash Reporting

Provisioning a host is a complex matter that can fail for a lot of reasons. Your
workstation may have a wide variety of shell, network configuration, VPN, proxy
or firewall issues.  There are also reasons from the other end of the chain:
your cloud provider or the network in between.

To help `docker-machine` be as stable as possible, we added a monitoring of
crashes whenever you try to `create` or `upgrade` a host. This will send, over
HTTPS, to Bugsnag some information about your `docker-machine` version, build,
OS, ARCH, the path to your current shell and, the history of the last command as
you could see it with a `--debug` option.  This data is sent to help us pinpoint
recurring issues with `docker-machine` and will only be transmitted in the case
of a crash of `docker-machine`.

If you wish to opt out of error reporting, you can create a `no-error-report`
file in your `$HOME/.docker/machine` directory, and Docker Machine will disable
this behavior.  e.g.:

    $ mkdir -p ~/.docker/machine && touch ~/.docker/machine/no-error-report

Leaving the file empty is fine -- Docker Machine just checks for its presence.

## Getting help

Docker Machine is still in its infancy and under active development. If you need
help, would like to contribute, or simply want to talk about the project with
like-minded individuals, we have a number of open channels for communication.

-   To report bugs or file feature requests: please use the [issue tracker on
    Github](https://github.com/docker/machine/issues).
-   To talk about the project with people in real time: please join the
    `#docker-machine` channel on IRC.
-   To contribute code or documentation changes: please [submit a pull request on
    Github](https://github.com/docker/machine/pulls).

For more information and resources, please visit
[our help page](https://docs.docker.com/project/get-help/).

## Where to go next

-   Create and run a Docker host on your [local system using VirtualBox](get-started.md)
-   Provision multiple Docker hosts [on your cloud provider](get-started-cloud.md)
-   <a href="../drivers/" target="_blank">Docker Machine driver reference</a>
-   <a href="../reference/" target="_blank">Docker Machine subcommand reference</a>
