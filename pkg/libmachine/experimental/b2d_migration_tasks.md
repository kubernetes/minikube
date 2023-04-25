<!--[metadata]>
+++
draft=true
title = "Docker Machine"
description = "machine"
keywords = ["machine, orchestration, install, installation, docker, documentation"]
[menu.main]
parent="mn_install"
+++
<![end-metadata]-->

# Boot2Docker Migration

This document is a rough guide to what will need to be completed to support
migrating from boot2docker-cli to Machine.  It is not meant to be a user guide
but more so an internal guide to what we will want to support.

## Existing Boot2Docker Instances

We will need to import the disk to "migrate" the existing Docker data to the
new Machine.  This should not be too much work as instead of creating the 
virtual disk we will simply copy this one.  From there, provisioning should
happen as normal (cert regeneration, option configuration, etc).

## CLI

Currently almost every b2d command has a comparable Machine command.  I do not
feel we need to have the exact same naming but we will want to create a 
migration user guide to inform the users of what is different.

## Boot2Docker Host Alias

Boot2Docker also modifies the local system host file to create a `boot2docker`
alias that can be used by the host system.  We will need to decide if we want
to support this and, if so, how to implement.  Perhaps local aliases for each
Machine name?

## Installer and Initial Setup

There is a Boot2Docker installer that assists the users in getting started.
It installs VirtualBox along with the b2d CLI.  We will need something similar.
This will probably be part of a larger installation project with the various
Docker platform tools.

## Updates

Machine already supports the `upgrade` command to update the Machine instances.
I'm not sure if we want to add a mechanism to update the local Machine binary
and/or the Docker CLI binary as well.  We will need to discuss.
