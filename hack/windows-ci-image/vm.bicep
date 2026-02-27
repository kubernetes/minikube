targetScope = 'resourceGroup'

// vm.bicepparam
param sigName string
param sigImageDefinitionName string
param sigImageVersion string
param vmName string
param vmSize string

var location string = resourceGroup().location
var namePrefix string = vmName
var nameSuffix string = uniqueString(resourceGroup().location)
var networkInterfaceName string = '${namePrefix}-nic-${nameSuffix}'
var networkSecurityGroupName string = '${namePrefix}-nsg-${nameSuffix}'
var publicIpName string = '${namePrefix}-pip-${nameSuffix}'
var subnetName string = '${namePrefix}-snet-${nameSuffix}'
var virtualMachineName string = '${vmName}-${nameSuffix}'
var virtualNetworkName string = '${namePrefix}-vnet-${nameSuffix}'

resource sig 'Microsoft.Compute/galleries@2024-03-03' existing = {
  name: sigName
}

resource sigImageDefinition 'Microsoft.Compute/galleries/images@2022-08-03' existing = {
  name: sigImageDefinitionName
  parent: sig
}

resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-11-01' = {
  name: virtualNetworkName
  location: location
  tags: { project: 'minikube' }
  properties: {
    addressSpace: {
      addressPrefixes: [
        '10.0.0.0/16'
      ]
    }
    subnets: [
      {
        name: subnetName
        properties: {
          addressPrefix: '10.0.1.0/24'
          networkSecurityGroup: {
            id: networkSecurityGroup.id
          }
        }
      }
    ]
  }
}

resource publicIp 'Microsoft.Network/publicIPAddresses@2023-11-01' = {
  name: publicIpName
  location: location
  tags: { project: 'minikube' }
  sku: { name: 'Standard' } // Basic: Cannot create more than 0 IPv4 Basic SKU public IP addresses for this subscription in this region.
  properties: {
    dnsSettings: {
      domainNameLabel: vmName
    }
    publicIPAddressVersion: 'IPv4'
    publicIPAllocationMethod: 'Static' // Dynamic: Standard sku publicIp /subscriptions/... must have AllocationMethod set to Static.
  }
}

resource networkSecurityGroup 'Microsoft.Network/networkSecurityGroups@2023-11-01' = {
  name: networkSecurityGroupName
  location: location
  tags: { project: 'minikube' }
  properties: {
    securityRules: [
      {
        name: 'AllowRDP'
        properties: {
          access: 'Allow'
          destinationAddressPrefix: '*'
          destinationPortRange: '3389'
          direction: 'Inbound'
          priority: 1000
          protocol: 'Tcp'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
        }
      }
      {
        name: 'AllowSSH'
        properties: {
          access: 'Allow'
          destinationAddressPrefix: '*'
          destinationPortRange: '22'
          direction: 'Inbound'
          priority: 1001
          protocol: 'Tcp'
          sourceAddressPrefix: '*'
          sourcePortRange: '*'
        }
      }
    ]
  }
}

resource networkInterface 'Microsoft.Network/networkInterfaces@2023-11-01' = {
  name: networkInterfaceName
  location: location
  tags: { project: 'minikube' }
  properties: {
    ipConfigurations: [
      {
        name: 'internal'
        properties: {
          subnet: {
            id: '${virtualNetwork.id}/subnets/${subnetName}'
          }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: {
            id: publicIp.id
          }
        }
      }
    ]
    networkSecurityGroup: {
      id: networkSecurityGroup.id
    }
  }
}

resource virtualMachine 'Microsoft.Compute/virtualMachines@2023-09-01' = {
  name: virtualMachineName
  location: location
  tags: { project: 'minikube' }
  properties: {
    hardwareProfile: {
      vmSize: vmSize
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
    networkProfile: {
      networkInterfaces: [
        {
          id: networkInterface.id
          properties: {
            primary: true
          }
        }
      ]
    }
  }
}

output vmId string = virtualMachine.id
output vmName string = virtualMachine.name
output vmHostname string = publicIp.properties.dnsSettings.fqdn
output publicIpAddress string = publicIp.properties.ipAddress
