targetScope = 'resourceGroup'

// sig.bicepparam
@sys.description('Name of the new VM image definition to create in the newly created Shared Image Gallery.')
param sigImageDefinitionName string
@sys.description('Name of the new Shared Image Gallery to create')
param sigName string

var location string = resourceGroup().location
var imageOffer string = 'minikube-ci'
var imagePublisher string = 'kubernetes'

resource sig 'Microsoft.Compute/galleries@2024-03-03' = {
  name: sigName
  location: location
  tags: { owner: 'minikube' }
  properties: {
    description: 'Gallery for Minikube VM images'
    identifier: {}
  }
}

resource image 'Microsoft.Compute/galleries/images@2024-03-03' = {
  location: location
  name: sigImageDefinitionName
  parent: sig
  tags: { owner: 'minikube' }
  properties: {
    description: 'Minimal Windows 11 image for Minikube CI'
    identifier: {
      publisher: imagePublisher
      offer: imageOffer
      sku: 'windows-11'
    }
    architecture: 'x64'
    hyperVGeneration: 'V2'
    osState: 'Specialized'
    osType: 'Windows'
  }
}

output sigId string = sig.id
output sigImageDefinitionId string = image.id
