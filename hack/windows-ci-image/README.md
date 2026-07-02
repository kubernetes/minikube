# Build VM image with Windows 11 for CI on Azure

> [!CAUTION]
> Minikube was granted access to Azure resource group `SIG-CLUSTER-LIFECYCLE-MINIKUBE` managed by SIG-Cluster-Lifecycle/SIG-Windows.
> We must **NOT** attempt to delete this resource group!
>
> Unfortunately, locking the resource group would not help, becuase it also locks any contained resources:
>
>    ```bash
>    az group lock create --lock-type CanNotDelete \
>       --name "CanNotDelete-${MINIKUBE_AZ_RESOURCE_GROUP}" \
>       --resource-group "${MINIKUBE_AZ_RESOURCE_GROUP}" \
>       --notes "Group managed by SIG-Windows"
>    ```
>

## Features

The goal of the project is to build Windows-based golden image for Minikube in order
to create create short-lived ephemeral Azure Virtual Machines used for testing.

The project provides:

1. Bicep templates to create Shared Image Gallery on Azure in given resource group.
2. Packer templates to build single Virtual Machine image:
    - based on Windows 11 with Windows settings optimized and bloatware removed
    - [specialized, not generalized](https://learn.microsoft.com/en-us/azure/virtual-machines/shared-image-galleries#generalized-and-specialized-images), and [shallowly replicated](https://learn.microsoft.com/en-us/azure/virtual-machines/shared-image-galleries#shallow-replication)
    - provisioned with local administrator (user-provided username (default: `minikubeadmin`) and password)
    - provisioned with user-provided public SSH key for streamlined authentication
    - provisioned with [tools required to run Minikube tests](packer/provisioners/))

3. Packer configuration to publish it to the Azure Shared Image Gallery.
4. Bicep templates to create Azure Virtual Machine from the published image to verify the output of the image build.

## Usage

Help: `make help`

### Configure

1. Set up Azure CLI:

    ```bash
    az login
    ```

    or keep the Azure CLI profile in project-specific location (recommended? safety?):

    ```bash
    export AZURE_CONFIG_DIR=$HOME/minikube/.azure
    az login
    ```

2. Set up environment variables

    Run `make preflight` to check list of all required `MINIKUBE_AZ_*` environment variables.

    Review the reasonable defaults in [env.sh](env.sh).

    Create your user-specific `env.local.sh` setting:

    - `MINIKUBE_AZ_VM_ADMIN_PASSWORD` with password of your choice (see )
    - `MINIKUBE_AZ_VM_ADMIN_SSH_PUBLIC_KEY=$(cat ~/.ssh/minikubeadmin.id_ecdsa.pub | base64 -w 0)` or equivalent
    - optionally, any of the others you wish to override their default values

    Load the environment variables:

    ```bash
    source ./env.sh
    ```

    Alternatively, install [mise](https://mise.jdx.dev) and let it load [.mise.toml](.mise.toml).

### Build

Run `make help` to learn about available commands.

```bash
source ./env.sh
make preflight                  # verify your environment
make sig-deploy                 # create Shared Image Gallery on Azure
make packer-build-and-publish   # build VM image and push it to the gallery
```

### Test

Find IP and FQDN of VM:

```bash
make vm-deploy          # create VM from the image
make vm-fqdn            # print public IP and FQDN
```

Connect to PowerShell on VM:
```bash
make vm-generate-ssh    # generate script with ssh command
./vm-generate-ssh.sh    # uses the insecure SSH key
```

Connect to PowerShell on VM using the insecure SSH key:

```bash
ssh -i ~/.ssh/minikube-ci.id_ecdsa  -o StrictHostKeyChecking=no minikubeadmin@vm-minikube-ci.${AZURE_DEFAULTS_LOCATION}.cloudapp.azure.com
```

Connect to PowerShell on VM using password:

```bash
export SSHPASS="${MINIKUBE_AZ_VM_ADMIN_PASSWORD}"
sshpass -e ssh -o StrictHostKeyChecking=no minikubeadmin@vm-minikube-ci.${AZURE_DEFAULTS_LOCATION}.cloudapp.azure.com
```

On VM, test if Docker is functioning properly:

```powershell
docker images ls
docker run hello-world
```

## Run Minikube

Connect to PowerShell on VM via SSH and run the following commands according to the [docs](https://minikube.sigs.k8s.io/docs/start/?arch=%2Fwindows%2Fx86-64%2Fstable%2F.exe+download):

```powershell
$ProgressPreference = 'SilentlyContinue'; Invoke-WebRequest -OutFile 'c:\minikube\minikube.exe' -Uri 'https://github.com/kubernetes/minikube/releases/latest/download/minikube-windows-amd64.exe' -UseBasicParsing
```

```powershell
$env:PATH += ';C:\Minikube
```

```powershell
minikube start --container-runtime=docker --vm=true
```
