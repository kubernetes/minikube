# Migrate from Boot2Docker CLI to Docker Machine

This guide explains migrating from the Boot2Docker CLI to Docker Machine.

This guide assumes basic knowledge of the Boot2Docker CLI and Docker Machine.  If you are not familiar, please review those docs prior to migrating.

There are a few differences between the Boot2Docker CLI commands and Machine.  Please review the table below for the Boot2Docker command and the corresponding Machine command.  You can also see details on Machine commands in the official [Docker Machine Docs](http://docs.docker.com/machine/#subcommands).

# Migrating

In order to migrate a Boot2Docker VM to Docker Machine, you must have Docker Machine installed.  If you do not have Docker Machine, please see the [install docs](http://docs.docker.com/machine/#installation) before proceeding.

> Note: when migrating to Docker Machine, this will also update Docker to the latest stable version

To migrate a Boot2Docker VM, run the following command where `<boot2docker-vm-name>` is the name of your Boot2Docker VM and `<new-machine-name>` is the name of the new Machine (i.e. `dev`):

> To get the name of your Boot2Docker VM, use the `boot2docker config` command.  Default: `boot2docker-vm`.

    docker-machine create -d virtualbox --virtualbox-import-boot2docker-vm <boot2docker-vm-name> <new-machine-name>

> Note: this will stop the Boot2Docker VM in order to safely copy the virtual disk

You should see output similar to the following:

    $> docker-machine create -d virtualbox --virtualbox-import-boot2docker-vm boot2docker-vm dev
    INFO[0000] Creating VirtualBox VM...
    INFO[0001] Starting VirtualBox VM...
    INFO[0001] Waiting for VM to start...
    INFO[0035] "dev" has been created and is now the active machine.
    INFO[0035] To point your Docker client at it, run this in your shell: eval "$(docker-machine env dev)"

You now should have a Machine that contains all of the Docker data from the Boot2Docker VM.  See the Docker Machine [usage docs](http://docs.docker.com/machine/#getting-started-with-docker-machine-using-a-local-vm) for details on working with Machine.

# Cleanup

When migrating a Boot2Docker VM to Docker Machine the Boot2Docker VM is left intact.  Once you have verified that all of your Docker data (containers, images, etc) are in the new Machine, you can remove the Boot2Docker VM using `boot2docker delete`.

# Command Comparison

| boot2docker cli | machine     | machine description                                                                    |
| --------------- | ----------- | -------------------------------------------------------------------------------------- |
| init            | create      | creates a new docker host                                                              |
| up              | start       | starts a stopped machine                                                               |
| ssh             | ssh         | runs a command or interactive ssh session on the machine                               |
| save            | -           | n/a                                                                                    |
| down            | stop        | stops a running machine                                                                |
| poweroff        | stop        | stops a running machine                                                                |
| reset           | restart     | restarts a running machine                                                             |
| config          | inspect (*) | shows details about machine                                                            |
| status          | ls (**)     | shows a list of all machines                                                           |
| info            | inspect (*) | shows details about machine                                                            |
| ip              | url (***)   | shows the Docker URL for the machine                                                   |
| shellinit       | env         | shows the environment configuration needed to configure the Docker CLI for the machine |
| delete          | rm          | removes a machine                                                                      |
| download        | -           |                                                                                        |
| upgrade         | upgrade     | upgrades Docker on the machine to the latest stable release                            |

\* provides similar functionality but not exact

** `ls` will show all machines including their status

*** the `url` command reports the entire Docker URL including the IP / Hostname
