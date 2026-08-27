// runner.bicep - Ephemeral Azure VM for windows-node-selfhosted CI tests.
// Created fresh per run and deleted after. VM size Standard_D16s_v3 supports
// nested virtualization required for Hyper-V. Uses a marketplace Windows 11 image
// so no Shared Image Gallery access is required. Windows 11 includes the
// Hyper-V Default Switch out of the box, providing DHCP to nested VMs.
targetScope = 'resourceGroup'

param vmName string
param vmSize string = 'Standard_D16s_v3'
param adminUsername string = 'minikubeadmin'
@secure()
param adminPassword string

var location = resourceGroup().location
var nameSuffix = uniqueString(resourceGroup().id, vmName)
var networkInterfaceName = '${vmName}-nic-${nameSuffix}'
var networkSecurityGroupName = '${vmName}-nsg-${nameSuffix}'
var publicIpName = '${vmName}-pip-${nameSuffix}'
var subnetName = '${vmName}-snet-${nameSuffix}'
var virtualNetworkName = '${vmName}-vnet-${nameSuffix}'

resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-11-01' = {
  name: virtualNetworkName
  location: location
  tags: { project: 'minikube', purpose: 'ci-runner' }
  properties: {
    addressSpace: {
      addressPrefixes: ['10.0.0.0/16']
    }
    subnets: [
      {
        name: subnetName
        properties: {
          addressPrefix: '10.0.1.0/24'
          networkSecurityGroup: { id: networkSecurityGroup.id }
        }
      }
    ]
  }
}

resource publicIp 'Microsoft.Network/publicIPAddresses@2023-11-01' = {
  name: publicIpName
  location: location
  tags: { project: 'minikube', purpose: 'ci-runner' }
  sku: { name: 'Standard' }
  properties: {
    dnsSettings: { domainNameLabel: vmName }
    publicIPAddressVersion: 'IPv4'
    publicIPAllocationMethod: 'Static'
  }
}

resource networkSecurityGroup 'Microsoft.Network/networkSecurityGroups@2023-11-01' = {
  name: networkSecurityGroupName
  location: location
  tags: { project: 'minikube', purpose: 'ci-runner' }
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
  tags: { project: 'minikube', purpose: 'ci-runner' }
  properties: {
    ipConfigurations: [
      {
        name: 'internal'
        properties: {
          subnet: { id: '${virtualNetwork.id}/subnets/${subnetName}' }
          privateIPAllocationMethod: 'Dynamic'
          publicIPAddress: { id: publicIp.id }
        }
      }
    ]
    networkSecurityGroup: { id: networkSecurityGroup.id }
  }
}

resource virtualMachine 'Microsoft.Compute/virtualMachines@2023-09-01' = {
  name: vmName
  location: location
  tags: { project: 'minikube', purpose: 'ci-runner' }
  properties: {
    hardwareProfile: { vmSize: vmSize }
    osProfile: {
      computerName: take(vmName, 15)
      adminUsername: adminUsername
      adminPassword: adminPassword
      windowsConfiguration: {
        enableAutomaticUpdates: false
        patchSettings: { patchMode: 'Manual' }
      }
    }
    storageProfile: {
      imageReference: {
        publisher: 'MicrosoftWindowsDesktop'
        offer: 'windows-11'
        sku: 'win11-24h2-pro'
        version: 'latest'
      }
      osDisk: {
        createOption: 'FromImage'
        managedDisk: { storageAccountType: 'StandardSSD_LRS' }
        deleteOption: 'Delete'
      }
    }
    networkProfile: {
      networkInterfaces: [
        {
          id: networkInterface.id
          properties: { primary: true }
        }
      ]
    }
  }
}

// Enable OpenSSH server so the provisioning workflow can connect via SSH/SCP.
// Hyper-V is enabled separately in the provisioning workflow (requires a reboot).
resource enableSsh 'Microsoft.Compute/virtualMachines/extensions@2023-09-01' = {
  name: 'EnableOpenSSH'
  parent: virtualMachine
  location: location
  properties: {
    publisher: 'Microsoft.Compute'
    type: 'CustomScriptExtension'
    typeHandlerVersion: '1.10'
    autoUpgradeMinorVersion: true
    settings: {
      commandToExecute: 'powershell -ExecutionPolicy Bypass -Command "Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0; Start-Service sshd; Set-Service -Name sshd -StartupType Automatic; if (-not (Get-NetFirewallRule -Name sshd -ErrorAction SilentlyContinue)) { New-NetFirewallRule -Name sshd -DisplayName sshd -Enabled True -Direction Inbound -Protocol TCP -Action Allow -LocalPort 22 }"'
    }
  }
}

output vmId string = virtualMachine.id
output vmName string = virtualMachine.name
output hostname string = publicIp.properties.dnsSettings.fqdn
output publicIpAddress string = publicIp.properties.ipAddress
