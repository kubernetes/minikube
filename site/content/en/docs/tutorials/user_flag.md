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

A good use case for the `--user` flag is if you have an application that starts and stops minikube clusters.
Assume the application will use an exsiting cluster if available, otherwise, it will start a new one.
The problem comes when the application is finished using the cluster, you only want to stop the running cluster if the application started the cluster, not if it was already existing.

This is where the user flag comes into play.
If the application was configured to pass a user flag on minikube commands (ex. `--user=app123`) then you could check to see what user executed the last `start` command looking at the audit log.
If the last user was `app123` you're safe to stop the cluster, otherwise leave it running.
