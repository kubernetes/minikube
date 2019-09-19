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

### Using Hyper-V Manager

1. Open the Hyper-V Manager. (On Windows 10, search for the Hyper-V Manager in the lower left search field.)
2. Select the Virtual Switch Manager on the right-hand Actions panel.
3. Set up a new external network switch to use, named `ExternalSwitch`
  * If you already have another network switch set up, use that one instead but make sure it is an external switch.

### Using PowerShell

**NOTE: This command is speculative, and may not work in all environments. You may need to use the Hyper-V Administrator UI for wireless interfaces.**

```powershell
New-VMSwitch -name ExternalSwitch -NetAdapterName Ethernet -AllowManagementOS $true
```

## Usage

Set the switch you created as the minikube default:

```shell
minikube config set hyperv-virtual-switch ExternalSwitch
```

```shell
minikube start --vm-driver=hyperv 
```
To make hyperv the default driver:

```shell
minikube config set vm-driver hyperv
```
