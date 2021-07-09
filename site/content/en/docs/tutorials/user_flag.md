---
title: "Using the User Flag"                     
linkTitle: "Using the User Flag"
weight: 1
date: 2021-06-15
description: >
  Using the User Flag to Keep an Audit Log
---

## Overview

In minikube, all executed commands are logged to a local audit log in the minikube home directory (default: `~/.minikube/logs/audit.json`).
These commands are logged with additional information including the user that ran them, which by default is the OS user.
However, there is a global flag `--user` that will set the user who ran the command in the audit log.

## Prerequisites

- minikube v1.17.1 or newer

## What does the flag do?

Assuming the OS user is `johndoe`, running `minikube start` will add the following to the audit log:
```
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
|    Command    |          Args            |           Profile           |     User     |    Version     |          Start Time           |           End Time            |
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
| start         |                          | minikube                    | johndoe      | v1.21.0        | Tue, 15 Jun 2021 09:00:00 MST | Tue, 15 Jun 2021 09:01:00 MST |
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
```
As you can see, minikube pulled the OS user and listed them as the user for the command.

Running the same command with `--user=mary` appended to the command will add the following to the audit log:
```
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
|    Command    |          Args            |           Profile           |     User     |    Version     |          Start Time           |           End Time            |
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
| start         | --user=mary              | minikube                    | mary         | v1.21.0        | Tue, 15 Jun 2021 09:00:00 MST | Tue, 15 Jun 2021 09:01:00 MST |
|---------------|--------------------------|-----------------------------|--------------|----------------|-------------------------------|-------------------------------|
```
Here you can see that passing `--user=mary` overwrote the OS user with `mary` as the user for the command.

## Example use case

- Embedded use of minikube by multiple users (IDEs, Plugins, etc.)
- A machine shared by multiple users using the same home folder

## How do I use minikube in a script?

If you are using minikube in a script or plugin it is recommeneded to add `--user=your_script_name` to all operations.

Example:
```
minikube start --user=plugin_name
minikube profile list --user=plugin_name
minikube stop --user=plugin_name
```
