# Docker Machine

![](https://docs.docker.com/machine/img/logo.png)

Machine lets you create Docker hosts on your computer, on cloud providers, and
inside your own data center. It creates servers, installs Docker on them, then
configures the Docker client to talk to them.

It works a bit like this:

```console
$ docker-machine create -d virtualbox default
Running pre-create checks...
Creating machine...
(default) Creating VirtualBox VM...
(default) Creating SSH key...
(default) Starting VM...
Waiting for machine to be running, this may take a few minutes...
Machine is running, waiting for SSH to be available...
Detecting operating system of created instance...
Detecting the provisioner...
Provisioning with boot2docker...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
Checking connection to Docker...
Docker is up and running!
To see how to connect Docker to this machine, run: docker-machine env default

$ docker-machine ls
NAME      ACTIVE   DRIVER       STATE     URL                         SWARM   DOCKER   ERRORS
default   -        virtualbox   Running   tcp://192.168.99.188:2376           v1.9.1

$ eval "$(docker-machine env default)"

$ docker run busybox echo hello world
Unable to find image 'busybox:latest' locally
511136ea3c5a: Pull complete
df7546f9f060: Pull complete
ea13149945cb: Pull complete
4986bf8c1536: Pull complete
hello world
```

In addition to local VMs, you can create and manage cloud servers:

```console
$ docker-machine create -d digitalocean --digitalocean-access-token=secret staging
Creating SSH key...
Creating Digital Ocean droplet...
To see how to connect Docker to this machine, run: docker-machine env staging

$ docker-machine ls
NAME      ACTIVE   DRIVER         STATE     URL                         SWARM   DOCKER   ERRORS
default   -        virtualbox     Running   tcp://192.168.99.188:2376           v1.9.1
staging   -        digitalocean   Running   tcp://203.0.113.81:2376             v1.9.1
```

## Installation and documentation

Full documentation [is available here](https://docs.docker.com/machine/).

## Contributing

Want to hack on Machine? Please start with the [Contributing Guide](https://github.com/docker/machine/blob/master/CONTRIBUTING.md).

## Driver Plugins

In addition to the core driver plugins bundled alongside Docker Machine, users
can make and distribute their own plugin for any virtualization technology or
cloud provider.  To browse the list of known Docker Machine plugins, please [see
this document in our
docs repo](https://github.com/docker/docker.github.io/blob/master/machine/AVAILABLE_DRIVER_PLUGINS.md).

## Troubleshooting

Docker Machine tries to do the right thing in a variety of scenarios but
sometimes things do not go according to plan.  Here is a quick troubleshooting
guide which may help you to resolve of the issues you may be seeing.

Note that some of the suggested solutions are only available on the Docker
Machine master branch.  If you need them, consider compiling Docker Machine from
source.
#### `docker-machine` hangs

A common issue with Docker Machine is that it will hang when attempting to start
up the virtual machine.  Since starting the machine is part of the `create`
process, `create` is often where these types of errors show up.

A hang could be due to a variety of factors, but the most common suspect is
networking.  Consider the following:

-   Are you using a VPN?  If so, try disconnecting and see if creation will
    succeed without the VPN.  Some VPN software aggressively controls routes and
    you may need to [manually add the route](https://github.com/docker/machine/issues/1500#issuecomment-121134958).
-   Are you connected to a proxy server, corporate or otherwise?  If so, take a
    look at the `--no-proxy` flag for `env` and at [setting environment variables
    for the created Docker Engine](https://docs.docker.com/machine/reference/create/#specifying-configuration-options-for-the-created-docker-engine).
-   Are there a lot of host-only interfaces listed by the command `VBoxManage list
    hostonlyifs`?  If so, this has sometimes been known to cause bugs.  Consider
    removing the ones you are not using (`VBoxManage hostonlyif remove name`) and
    trying machine creation again.

We are keenly aware of this as an issue and working towards a set of solutions
which is robust for all users, so please give us feedback and/or report issues,
workarounds, and desired workflows as you discover them.

#### Machine creation errors out before finishing

If you see messages such as "exit status 1" creating machines with VirtualBox,
this frequently indicates that there is an issue with VirtualBox itself.  Please
[file an issue](https://github.com/docker/machine/issues/new) and include a link
to a [Github Gist](https://gist.github.com/) with the output of the VirtualBox
log (usually located at
`$HOME/.docker/machine/machines/machinename/machinename/Logs/VBox.log`), as well
as the output of running the Docker Machine command which is failing with the
global `--debug` flag enabled.  This will help us to track down which versions
of VirtualBox are failing where, and under which conditions.

If you see messages such as "exit status 255", this frequently indicates there
has been an issue with SSH.  Please investigate your SSH configuration if you
have one, and/or [file an issue](https://github.com/docker/machine/issues).

#### "You may be getting rate limited by Github" error message

In order to `create` or `upgrade` virtual machines running Docker, Docker
Machine will check the Github API for the latest release of the [boot2docker
operating system](https://github.com/boot2docker/boot2docker).  The Github API
allows for a small number of unauthenticated requests from a given client, but
if you share an IP address with many other users (e.g. in an office), you may
get rate limited by their API, and Docker Machine will error out with messages
indicating this.

In order to work around this issue, you can [generate a
token](https://help.github.com/articles/creating-an-access-token-for-command-line-use/)
and pass it to Docker Machine using the global `--github-api-token` flag like
so:

```console
$ docker-machine --github-api-token=token create -d virtualbox newbox
```

This should eliminate any issues you've been experiencing with rate limiting.
