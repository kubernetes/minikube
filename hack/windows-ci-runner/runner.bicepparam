using './runner.bicep'

// Ephemeral test VM name — overridden at deploy time with a unique per-run name
param vmName = 'vm-minikube-win11'

// Standard_D16s_v3: 16 vCPUs, 64 GiB RAM, supports nested virtualization for Hyper-V
param vmSize = 'Standard_D16s_v3'

param adminUsername = 'minikubeadmin'

param adminPassword = readEnvironmentVariable('MINIKUBE_AZ_VM_PASSWORD', '')
