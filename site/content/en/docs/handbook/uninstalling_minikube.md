---
title: "Uninstall"
linkTitle: "Uninstall"
weight: 99
draft: true 
date: 2019-08-18
description: >
  Reference on uninstalling minikube
---

NOTE: To be moved to the FAQ

## Chocolatey

- Open a command prompt with Administrator privileges.
- Run `minikube delete --purge --all`
- Run, `choco uninstall minikube` to remove the minikube package from your system.

## Windows Installer

- Open a command prompt with Administrator privileges.
- Run `minikube delete --purge --all`
- Open the Run dialog box (**Win+R**), type in `appwiz.cpl` and hit **Enter** key.
- In there, find an entry for the Minikube installer, right click on it & click on **Uninstall**.

## Binary/Direct

- Open a command prompt with Administrator privileges.
- Run `minikube delete --purge --all`
- Delete the minikube binary.

## Debian/Ubuntu (Deb)

- Run `minikube delete --purge --all`
- Run `sudo dpkg -P minikube`

## Fedora/Red Hat (RPM)

- Run `minikube delete --purge --all`
- Run  `sudo rpm -e minikube`

## Brew

- Run `minikube delete --purge --all`
- Run `brew uninstall minikube`
