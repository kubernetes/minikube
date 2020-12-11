## Requirements

*VMware Workstation

## Driver Installation

First install the [VMware Workstation](https://www.vmware.com/in/products/workstation-pro/workstation-pro-evaluation.html) 

After installation, start a cluster using the vmware driver:

```shell
minikube start --memory=4096 --cpus=4 --vm-driver vmware
```
