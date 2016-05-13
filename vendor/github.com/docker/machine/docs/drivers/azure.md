<!--[metadata]>
+++
title = "Microsoft Azure"
description = "Microsoft Azure driver for machine"
keywords = ["machine, Microsoft Azure, driver"]
[menu.main]
parent="smn_machine_drivers"
+++
<![end-metadata]-->

# Microsoft Azure

You will need an Azure Subscription to use this Docker Machine driver.
[Sign up for a free trial.][trial]

> **NOTE:** This documentation is for the new version of the Azure driver, which started
> shipping with v0.7.0. This driver is not backwards-compatible with the old
> Azure driver. If you want to continue managing your existing Azure machines, please
> download and use machine versions prior to v0.7.0.

[azure]: http://azure.microsoft.com/
[trial]: https://azure.microsoft.com/free/

## Authentication

The first time you try to create a machine, Azure driver will ask you to
authenticate:

    $ docker-machine create --driver azure --azure-subscription-id <subs-id> <machine-name>
    Running pre-create checks...
    Microsoft Azure: To sign in, use a web browser to open the page https://aka.ms/devicelogin.
    Enter the code [...] to authenticate.

After authenticating, the driver will remember your credentials up to two weeks.

> **KNOWN ISSUE:** There is a known issue with Azure Active Directory causing stored
> credentials to expire within hours rather than 14 days when the user logs in with
> personal Microsoft Account (formerly _Live ID_) instead of an Active Directory account.
> Currently, there is no ETA for resolution, however in the meanwhile you can
> [create an AAD account][aad-docs] and login with that as a workaround.

[aad-docs]: https://azure.microsoft.com/documentation/articles/virtual-machines-windows-create-aad-work-id/

## Options

Azure driver only has a single required argument to make things easier. Please
read the optional flags to configure machine details and placement further.

Required:

- `--azure-subscription-id`: **(required)** Your Azure Subscription ID.

Optional:

- `--azure-image`: Azure virtual machine image in the format of Publisher:Offer:Sku:Version [[?][vm-image]]
- `--azure-location`: Azure region to create the virtual machine. [[?][location]]
- `--azure-resource-group`: Azure Resource Group name to create the resources in.
- `--azure-size`: Size for Azure Virtual Machine. [[?][vm-size]]
- `--azure-ssh-user`: Username for SSH login.
- `--azure-vnet`: Azure Virtual Network name to connect the virtual machine. [[?][vnet]]
- `--azure-subnet`: Azure Subnet Name to be used within the Virtual Network.
- `--azure-subnet-prefix`: Private CIDR block. Used to create subnet if it does not exist. Must match in the case that the subnet does exist.
- `--azure-availability-set`: Azure Availability Set to place the virtual machine into. [[?][av-set]]
- `--azure-open-port`: Make additional port number(s) accessible from the Internet [[?][nsg]]
- `--azure-private-ip-address`: Specify a static private IP address for the machine.
- `--azure-use-private-ip`: Use private IP address of the machine to connect. It's useful for managing Docker machines from another machine on the same network e.g. while deploying Swarm.
- `--azure-no-public-ip`: Do not create a public IP address for the machine (implies `--azure-use-private-ip`). Should be used only when creating machines from an Azure VM within the same subnet.
- `--azure-static-public-ip`: Assign a static public IP address to the machine.
- `--azure-docker-port`: Port number for Docker engine [$AZURE_DOCKER_PORT]
- `--azure-environment`: Azure environment (e.g. `AzurePublicCloud`, `AzureChinaCloud`).

[vm-image]: https://azure.microsoft.com/en-us/documentation/articles/resource-groups-vm-searching/
[location]: https://azure.microsoft.com/en-us/regions/
[vm-size]:  https://azure.microsoft.com/en-us/documentation/articles/virtual-machines-size-specs/
[vnet]:     https://azure.microsoft.com/en-us/documentation/articles/virtual-networks-overview/
[av-set]:   https://azure.microsoft.com/en-us/documentation/articles/virtual-machines-manage-availability/

Environment variables and default values:

| CLI option                      | Environment variable          | Default            |
| ------------------------------- | ----------------------------- | ------------------ |
| **`--azure-subscription-id`**   | `AZURE_SUBSCRIPTION_ID`       | -                  |
| `--azure-environment`           | `AZURE_ENVIRONMENT`           | `AzurePublicCloud` |
| `--azure-image`                 | `AZURE_IMAGE`                 | `canonical:UbuntuServer:15.10:latest` |
| `--azure-location`              | `AZURE_LOCATION`              | `westus`           |
| `--azure-resource-group`        | `AZURE_RESOURCE_GROUP`        | `docker-machine`   |
| `--azure-size`                  | `AZURE_SIZE`                  | `Standard_A2`      |
| `--azure-ssh-user`              | `AZURE_SSH_USER`              | `docker-user`      |
| `--azure-vnet`                  | `AZURE_VNET`                  | `docker-machine`   |
| `--azure-subnet`                | `AZURE_SUBNET`                | `docker-machine`   |
| `--azure-subnet-prefix`         | `AZURE_SUBNET_PREFIX`         | `192.168.0.0/16`   |
| `--azure-availability-set`      | `AZURE_AVAILABILITY_SET`      | `docker-machine`   |
| `--azure-open-port`             | -                             | -                  |
| `--azure-private-ip-address`    | -                             | -                  |
| `--azure-use-private-ip`        | -                             | -                  |
| `--azure-no-public-ip`          | -                             | -                  |
| `--azure-static-public-ip`      | -                             | -                  |
| `--azure-docker-port`           | `AZURE_DOCKER_PORT`           | `2376`             |

## Notes

Azure runs fully on the new [Azure Resource Manager (ARM)][arm] stack. Each
machine created comes with a few more Azure resources associated with it:

* A [Virtual Network][vnet] and a subnet under it is created to place your
machines into. This establishes a local network between your docker machines.
* An [Availability Set][av-set] is created to maximize availability of your
machines.

These are created once when the first machine is created and reused afterwards.
Although they are free resources, driver does a best effort to clean them up
after the last machine using these resources is removed.

Each machine is created with a public dynamic IP address for external
connectivity. All its ports (except Docker and SSH) are closed by default. You
can use `--azure-open-port` argument to specify multiple port numbers to be
accessible from Internet. 

Once the machine is created, you can modify [Network Security Group][nsg]
rules and open ports of the machine from the [Azure Portal][portal].

[arm]:    https://azure.microsoft.com/en-us/documentation/articles/resource-group-overview/
[nsg]:    https://azure.microsoft.com/en-us/documentation/articles/virtual-networks-nsg/
[portal]: https://portal.azure.com/
