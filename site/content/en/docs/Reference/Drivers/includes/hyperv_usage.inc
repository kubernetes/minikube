## Requirements

* Windows 10 Enterprise, Pro, or Education ([system requirements](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/reference/hyper-v-requirements))
* Hyper-V enabled
* An active Hyper-V switch

## Enabling Hyper-V

Open a PowerShell console as Administrator, and run the following command:

```powershell
Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
```

If Hyper-V was not previously active, you will need to reboot.

## Network Configuration

Create a Hyper-V external network switch  

{{% alert title="Untested" color="warning" %}}
This command is speculative, and may not work in all environments. You may need to use the Hyper-V Administrator UI for wireless interfaces.
{{% /alert %}}

```powershell
New-VMSwitch -name ExternalSwitch -NetAdapterName Ethernet -AllowManagementOS $true
```

Set this switch as the minikube default:

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
