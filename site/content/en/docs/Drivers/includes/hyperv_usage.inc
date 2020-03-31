## Requirements

* Windows 10 Enterprise, Pro, or Education ([system requirements](https://docs.microsoft.com/en-us/virtualization/hyper-v-on-windows/reference/hyper-v-requirements))
* Hyper-V enabled

## Enabling Hyper-V

Open a PowerShell console as Administrator, and run the following command:

```powershell
Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
```

If Hyper-V was not previously active, you will need to reboot.

## Usage

```shell
minikube start --driver=hyperv 
```
To make hyperv the default driver:

```shell
minikube config set driver hyperv
```
