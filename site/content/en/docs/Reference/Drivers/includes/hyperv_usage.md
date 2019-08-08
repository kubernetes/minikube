## Requirements

* Windows 10 Pro 
* Hyper-V enabled
* A Hyper-V switch created

## Configuring Hyper-V

Open a PowerShell console as Administrator, and run the following command:

```powershell
Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
```

Reboot, and create a new external network switch:

```powershell
New-VMSwitch -name ExternalSwitch  -NetAdapterName Ethernet -AllowManagementOS $true 
```

Set this network switch as the minikube default:

```shell
minikube config set hyperv-virtual-switch ExternalSwitch
```

## Usage

```shell
minikube start --vm-driver=hyperv 
```
To make hyperv the default driver:

```shell
minikube config set vm-driver hyperv
```
