<!--[metadata]>
+++
title = "Migrate from Boot2Docker to Machine"
description = "Migrate from Boot2Docker to Docker Machine"
keywords = ["machine, commands, boot2docker, migrate, docker"]
[menu.main]
parent="workw_machine"
weight=-30
+++
<![end-metadata]-->

# Migrate from Boot2Docker to Docker Machine

If you were using Boot2Docker previously, you have a pre-existing Docker
`boot2docker-vm` VM on your local system.  To allow Docker Machine to manage
this older VM, you must migrate it.

1.  Open a terminal or the Docker CLI on your system.

2.  Type the following command.

        $ docker-machine create -d virtualbox --virtualbox-import-boot2docker-vm boot2docker-vm docker-vm

3.  Use the `docker-machine` command to interact with the migrated VM.  

## Subcommand comparison

The `docker-machine` subcommands are slightly different than the `boot2docker`
subcommands. The table below lists the equivalent `docker-machine` subcommand
and what it does:

| `boot2docker` | `docker-machine` | `docker-machine` description                                                      |
| ------------- | ---------------- | --------------------------------------------------------------------------------- |
| init          | create           | Creates a new docker host.                                                        |
| up            | start            | Starts a stopped machine.                                                         |
| ssh           | ssh              | Runs a command or interactive ssh session on the machine.                         |
| save          | -                | Not applicable.                                                                   |
| down          | stop             | Stops a running machine.                                                          |
| poweroff      | stop             | Stops a running machine.                                                          |
| reset         | restart          | Restarts a running machine.                                                       |
| config        | inspect          | Prints machine configuration details.                                             |
| status        | ls               | Lists all machines and their status.                                              |
| info          | inspect          | Displays a machine's details.                                                     |
| ip            | ip               | Displays the machine's ip address.                                                |
| shellinit     | env              | Displays shell commands needed to configure your shell to interact with a machine |
| delete        | rm               | Removes a machine.                                                                |
| download      | -                | Not applicable.                                                                   |
| upgrade       | upgrade          | Upgrades a machine's Docker client to the latest stable release.                  |
