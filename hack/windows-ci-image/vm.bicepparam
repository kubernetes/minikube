using './vm.bicep'

// Pre-existing Azure Shared Image Gallery (see sig.bicepparam)
param sigName = readEnvironmentVariable('MINIKUBE_AZ_SIG_NAME')
param sigImageDefinitionName = readEnvironmentVariable('MINIKUBE_AZ_IMAGE_NAME') // same as image name by convention
param sigImageVersion = readEnvironmentVariable('MINIKUBE_AZ_IMAGE_VERSION') // or 'latest'

// Pre-existing Azure Virtual Network (see vnet.bicepparam)
param networkNsgName = 'nsg-minikube-ci'
param networkVnetName = 'vnet-minikube-ci'

// Azure Virtual Machine
param vmName = 'vm-minikube-ci'
param vmSize = 'Standard_D2s_v3'
