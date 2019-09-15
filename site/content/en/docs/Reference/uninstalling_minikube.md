---
title: "Uninstall minikube"
linkTitle: "Uninstall minikube"
weight: 6
date: 2019-08-18
description: >
  Reference on uninstalling minikube from your system completely.
---

# Uninstall minikube on Windows
Following are the ways you can install minikube on Windows. Depending on how you installed minikube, please follow the guide appropriately.

## Chocolatey
If you have installed minikube using Chocolatey package manager, follow the below steps to completely uninstall minikube from your system -
- Open a command prompt with Administrator privileges.
- We need to delete the cluster which was created by minikube - `minikube delete`
- Run, `choco uninstall minikube` to remove the minikube package from your system.
- Now, navigate to your User Folder - `C:\Users\YOUR_USER_NAME` (You can also find the path by expanding the environment variable `%USERPROFILE%`)
- In this folder, delete the `.minikube` folder.

## Windows Installer
If you have downloaded and installed minikube using the Windows Installer provided in our Releases, kindly follow the below steps -
- Open a command prompt with Administrator privileges.
- We need to delete the cluster which was created by minikube - `minikube delete`
- Now, open the Run dialog box (**Win+R**), type in `appwiz.cpl` and hit **Enter** key.
- In there, find an entry for the Minikube installer, right click on it & click on **Uninstall**.
- Follow the onscreen prompts to uninstall minikube from your system.
- Now, navigate to your User Folder - `C:\Users\YOUR_USER_NAME` (You can also find the path by expanding the environment variable `%USERPROFILE%`)
- In this folder, delete the `.minikube` folder.

## Binary/Direct
If you have downloaded just the binary and are using it to run minikube, please follow the below steps -
- Open a command prompt with Administrator privileges.
- We need to delete the cluster which was created by minikube - `minikube delete`
- Delete the minikube binary.
- Now, navigate to your User Folder - `C:\Users\YOUR_USER_NAME` (You can also find the path by expanding the environment variable `%USERPROFILE%`)
- In this folder, delete the `.minikube` folder.


# Uninstall minikube on Linux
## Binary/Direct
If you have installed minikube using the direct download method, follow the below steps to uninstall minikube completely from your system -
- In the shell, type in `minikube delete` to delete the minikube cluster.
- Remove the binary using `rm /usr/local/bin/minikube`
- Remove the directory containing the minikube configuration `rm -rf ~/.minikube`

## Debian/Ubuntu (Deb)
If you have installed minikube using the (deb) file, follow the below instructions -
- In the shell, type in `minikube delete` to delete the minikube cluster.
- Uninstall the minikube package completely - `sudo dpkg -P minikube`
- Remove the minikube configuration directory - `rm -rf ~/.minikube`

## Fedora/Red Hat (RPM)
If you have installed minikube using RPM, follow the below steps -
- In the shell, type in `minikube delete` to delete the minikube cluster.
- Uninstall the minikube package - `sudo rpm -e minikube`
- Remove the minikube configuration directory - `rm -rf ~/.minikube`


# Uninstall minikube on MacOS
## Binary/Direct
If you have installed minikube using the direct download method, follow the below steps to uninstall minikube completely from your system -
- In the shell, type in `minikube delete` to delete the minikube cluster.
- Remove the binary using `rm /usr/local/bin/minikube`
- Remove the directory containing the minikube configuration `rm -rf ~/.minikube`


## Brew
If you have installed minikube using the direct download method, follow the below steps to uninstall minikube completely from your system -
- In the shell, type in `minikube delete` to delete the minikube cluster.
- Uninstall the minikube package using `brew uninstall minikube`
- Remove the directory containing the minikube configuration `rm -rf ~/.minikube`
