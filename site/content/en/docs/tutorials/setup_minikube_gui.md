---
title: "Setting Up minikube GUI"
linkTitle: "Setting Up minikube GUI"
weight: 1
date: 2022-02-25
---

## Overview

- This guide will show you how to setup the minikube GUI
- **WARNING!** This GUI is a prototype and therefore may be unstable or contain bugs. Please use at your own risk, we are not responsible for damages.
- If you experience any bugs or have suggestions to improve the GUI feel free to create a [GitHub Issue](https://github.com/kubernetes/minikube/issues/new/choose).
- Please note that the SSH functionality currently only works on Linux.

## Before You Begin

- You will need to already have minikube setup on your machine, follow the [Getting Start doc]({{< ref "/docs/commands/start" >}}) if not already done.

## Steps

{{% tabs %}}
{{% mactab %}}
1. Download the zipped folder
```shell
curl -LO https://storage.googleapis.com/minikube-gui/nightly/minikube-gui-mac.zip
```

2. Unzip
```shell
unzip minikube-gui-mac.zip
```

3. Open the application
```shell
open dist/systray.app
```

4. If you see the following, click cancel.

![Mac unverified developer](/images/gui/mac.png)

5. Open System Preferences and go to Security & Privacy -> General and click "Open Anyway".
{{% /mactab %}}
{{% windowstab %}}
1. Download the zipped folder via PowerShell (below) or via your [browser](https://storage.googleapis.com/minikube-gui/nightly/minikube-gui-windows.zip) (faster)
```shell
Invoke-WebRequest -Uri 'https://storage.googleapis.com/minikube-gui/nightly/minikube-gui-windows.zip' -UseBasicParsing
```

2. Unzip
```shell
Expand-Archive minikube-gui-windows.zip
```

3. Open the application
```shell
.\minikube-gui-windows\dist\systray.exe
```

4. If you see the following, click `More info` and then `Run anyway`

![Windows unreconized app](/images/gui/windows.png)
{{% /windowstab %}}
{{% linuxtab %}}
1. Download the zipped folder
```shell
curl -LO https://storage.googleapis.com/minikube-gui/nightly/minikube-gui-linux.zip
```

2. Unzip
```shell
unzip minikube-gui-linux.zip
```

3. Open the application
```shell
dist/systray
```
{{% /linuxtab %}}
{{% /tabs %}}

