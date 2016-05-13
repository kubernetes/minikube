<!--[metadata]>
+++
title = "Provision hosts in the cloud"
description = "Using Docker Machine to provision hosts on cloud providers"
keywords = ["docker, machine, amazonec2, azure, digitalocean, google, openstack, rackspace, softlayer, virtualbox, vmwarefusion, vmwarevcloudair, vmwarevsphere, exoscale"]
[menu.main]
parent="workw_machine"
weight=-60
+++
<![end-metadata]-->

# Use Docker Machine to provision hosts on cloud providers

Docker Machine driver plugins are available for many cloud platforms, so you can use Machine to provision cloud hosts. When you use Docker Machine for provisioning, you create cloud hosts with Docker Engine installed on them.

You'll need to install and run Docker Machine, and create an account with the cloud provider.

Then you provide account verification, security credentials, and configuration options for the providers as flags to `docker-machine create`. The flags are unique for each cloud-specific driver.  For instance, to pass a Digital Ocean access token you use the `--digitalocean-access-token` flag. Take a look at the examples below for Digital Ocean and AWS.

## Examples

### Digital Ocean

For Digital Ocean, this command creates a Droplet (cloud host) called "docker-sandbox".

      $ docker-machine create --driver digitalocean --digitalocean-access-token xxxxx docker-sandbox

For a step-by-step guide on using Machine to create Docker hosts on Digital Ocean, see the [Digital Ocean Example](examples/ocean.md).

### Amazon Web Services (AWS)

For AWS EC2, this command creates an instance called "aws-sandbox":

      $ docker-machine create --driver amazonec2 --amazonec2-access-key AKI******* --amazonec2-secret-key 8T93C*******  aws-sandbox

For a step-by-step guide on using Machine to create Dockerized AWS instances, see the [Amazon Web Services (AWS) example](examples/aws.md).

## The docker-machine create command

The `docker-machine create` command typically requires that you specify, at a minimum:

* `--driver` - to indicate the provider on which to create the machine  (VirtualBox, DigitalOcean, AWS, and so on)

* Account verification and security credentials (for cloud providers), specific to the cloud service you are using

* `<machine>` - name of the host you want to create

For convenience, `docker-machine` will use sensible defaults for choosing settings such as the image that the server is based on, but you override the defaults using the respective flags (e.g. `--digitalocean-image`). This is useful if, for example, you want to create a cloud server with a lot of memory and CPUs (by default `docker-machine` creates a small server).

For a full list of the flags/settings available and their defaults, see the output of `docker-machine create -h` at the command line, the <a href="../reference/create/" target="_blank">create</a> command in the Machine <a href="../reference/" target="_blank">command line reference</a>, and <a href="https://docs.docker.com/machine/drivers/os-base/" target="_blank">driver options and operating system defaults</a> in the Machine driver reference.

## Drivers for cloud providers

When you install Docker Machine, you get a set of drivers for various cloud providers (like Amazon Web Services, Digital Ocean, or Microsoft Azure) and local providers (like Oracle VirtualBox, VMWare Fusion, or Microsoft Hyper-V).

See <a href="../drivers/" target="_blank">Docker Machine driver reference</a> for details on the drivers, including required flags and configuration options (which vary by provider).

## 3rd-party driver plugins

  Several Docker Machine driver plugins for use with other cloud platforms are available from 3rd party contributors. These are use-at-your-own-risk plugins, not maintained by or formally associated with Docker.

  See <a href="https://github.com/docker/machine/blob/master/docs/AVAILABLE_DRIVER_PLUGINS.md" target="_blank">Available driver plugins</a> in the docker/machine repo on GitHub.

## Adding a host without a driver

You can add a host to Docker which only has a URL and no driver. Then you can use the machine name you provide here for an existing host so you don’t have to type out the URL every time you run a Docker command.

    $ docker-machine create --url=tcp://50.134.234.20:2376 custombox
    $ docker-machine ls
    NAME        ACTIVE   DRIVER    STATE     URL
    custombox   *        none      Running   tcp://50.134.234.20:2376

## Using Machine to provision Docker Swarm clusters

Docker Machine can also provision <a href="https://docs.docker.com/swarm/overview/" target="_blank">Docker Swarm</a> clusters. This can be used with any driver and will be secured with TLS.

* To get started with Swarm, see <a href="https://docs.docker.com/swarm/get-swarm/" target="_blank">How to get Docker Swarm</a>.

* To learn how to use Machine to provision a Swarm cluster, see <a href="https://docs.docker.com/swarm/provision-with-machine/" target="_blank">Provision a Swarm cluster with Docker Machine</a>.

## Where to go next
-   Example: Provision Dockerized [Digital Ocean Droplets](examples/ocean.md)
-   Example: Provision Dockerized [AWS EC2 Instances](examples/aws.md)
-   [Understand Machine concepts](concepts.md)
-   [Docker Machine driver reference](drivers/index.md)
-   [Docker Machine subcommand reference](reference/index.md)
-   [Provision a Docker Swarm cluster with Docker Machine](https://docs.docker.com/swarm/provision-with-machine/)
