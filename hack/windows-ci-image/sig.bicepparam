using './sig.bicep'

param sigImageDefinitionName = readEnvironmentVariable('MINIKUBE_AZ_IMAGE_NAME')
param sigName = readEnvironmentVariable('MINIKUBE_AZ_SIG_NAME')
