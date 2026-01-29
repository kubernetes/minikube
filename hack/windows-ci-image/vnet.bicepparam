using './vnet.bicep'

// TODO(mloskot): Move to https://learn.microsoft.com/en-us/azure/azure-resource-manager/bicep/patterns-shared-variable-file
param networkNsgName = 'nsg-minikube-ci'
param networkVnetName = 'vnet-minikube-ci'
