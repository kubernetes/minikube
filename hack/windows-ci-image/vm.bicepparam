using './vm.bicep'

// Azure Shared Image Gallery
param sigName = readEnvironmentVariable('MINIKUBE_AZ_SIG_NAME')
param sigImageDefinitionName = readEnvironmentVariable('MINIKUBE_AZ_IMAGE_NAME') // same as image name by convention
param sigImageVersion = readEnvironmentVariable('MINIKUBE_AZ_IMAGE_VERSION') // or 'latest'

// Azure Virtual Machine
param vmName = 'vm-minikube-ci'
param vmSize = 'Standard_D16s_v3'
