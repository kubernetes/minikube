// Requires prior deployment of Shared Image Gallery (SIG) and Virtual Network (VNet)
targetScope = 'resourceGroup'

// vm.bicepparam
@sys.description('Name of existing VM image available from the Shared Image Gallery')
param sigImageDefinitionName string
@sys.description('Version of existing VM image available from the Shared Image Gallery')
param sigImageVersion string
@sys.description('Name of existing Shared Image Gallery containing the required VM image')
param sigName string

@sys.description('Name of existing Network Security Group where the new VMs will be associated.')
param networkNsgName string
@sys.description('Name of existing Virtual Network where the new VM will be attached.')
param networkVnetName string

@sys.description('Name of the new Virtual Machine, it will be used for the known VM hostname.')
param vmName string
@sys.description('Size of the new Virtual Machine.')
param vmSize string

var location string = resourceGroup().location
var nameSuffix string = uniqueString(subscription().id, resourceGroup().id)
var networkInterfaceName string = '${vmName}-nic-${nameSuffix}'
var publicIpName string = '${vmName}-pip-${nameSuffix}'
var virtualMachineName string = vmName // caller must ensure no names clash, so no suffix appended

resource sig 'Microsoft.Compute/galleries@2025-03-03' existing = {
  name: sigName
}

resource sigImageDefinition 'Microsoft.Compute/galleries/images@2025-03-03' existing = {
  name: sigImageDefinitionName
  parent: sig
}

resource virtualNetwork 'Microsoft.Network/virtualNetworks@2025-05-01' existing = {
  name: networkVnetName
}

resource networkSecurityGroup 'Microsoft.Network/networkSecurityGroups@2025-05-01' existing = {
  name: networkNsgName
}

resource publicIp 'Microsoft.Network/publicIPAddresses@2025-01-01' = {
  location: location
  name: publicIpName
  properties: {
    deleteOption: 'Delete'
    dnsSettings: {
      domainNameLabel: vmName // use as known user-specified hostname (e.g. GitHub workflow may randomise it with GitHub PR id)
    }
    publicIPAddressVersion: 'IPv4'
    publicIPAllocationMethod: 'Static' // Dynamic: Standard sku publicIp /subscriptions/... must have AllocationMethod set to Static.
  }
  sku: { name: 'Standard' } // Basic: Cannot create more than 0 IPv4 Basic SKU public IP addresses for this subscription in this region.
  tags: { owner: 'minikube', vm: virtualMachineName }
}

resource networkInterface 'Microsoft.Network/networkInterfaces@2025-05-01' = {
  location: location
  name: networkInterfaceName
  properties: {
    ipConfigurations: [
      {
        name: virtualMachineName
        properties: {
          primary: true
          subnet: {
            id: '${virtualNetwork.id}/subnets/${virtualNetwork.properties.subnets[0].name}'
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIp.id
            properties: {
              deleteOption: 'Delete'
            }
          }
        }
      }
    ]
    networkSecurityGroup: {
      id: networkSecurityGroup.id
    }
  }
  tags: { owner: 'minikube', vm: virtualMachineName }
}

resource virtualMachine 'Microsoft.Compute/virtualMachines@2025-04-01' = {
  location: location
  name: virtualMachineName
  properties: {
    hardwareProfile: {
      vmSize: vmSize
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: networkInterface.id
          properties: {
            deleteOption: 'Delete'
            primary: true
          }
        }
      ]
    }
    //osProfile: {} // Parameter OSProfile is not allowed with a specialized image.
    storageProfile: {
      imageReference: {
        id: '${sig.id}/images/${sigImageDefinition.name}/versions/${sigImageVersion}'
      }
      osDisk: {
        createOption: 'FromImage'
        managedDisk: {
          storageAccountType: 'StandardSSD_LRS'
        }
        deleteOption: 'Delete'
      }
    }
  }
  tags: { owner: 'minikube', vm: virtualMachineName }
}

output vmId string = virtualMachine.id
output vmHostname string = virtualMachine.name
output vmPublicFqdn string = publicIp.properties.dnsSettings.fqdn
output vmPublicIp string = publicIp.properties.ipAddress
