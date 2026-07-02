#!/bin/bash
set -e

VM_ID_OR_NAME="$1"
[ -z "${VM_ID_OR_NAME}" ] && { echo "VM id or name is empty"; exit 1; }

if [ "${VM_ID_OR_NAME:0:13}" = /subscription ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting VM ${VM_ID_OR_NAME}"
  VM_ID="${VM_ID_OR_NAME}"
elif [ "${VM_ID_OR_NAME:0:13}" != /subscription ]; then
  VM_RG="$2"
  [ -z "${VM_RG}" ] && VM_RG="${MINIKUBE_AZ_RESOURCE_GROUP}"
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting VM ${VM_ID_OR_NAME} in ${VM_RG}"
  VM_ID="$(az vm show --resource-group "${VM_RG}" --name "${VM_ID_OR_NAME}" --query "id" --output tsv)"
else
  echo "Usage: $(basename "${0}") <VM id>"
  echo "Usage: $(basename "${0}") <VM name> <VM resource group name>"
  exit 1
fi

VM_OSDISK_ON_PARENT_DELETE=$(az vm show --ids "${VM_ID}" --query 'storageProfile.osDisk.deleteOption' --output tsv)
if [ "${VM_OSDISK_ON_PARENT_DELETE}" != "Delete" ]; then
    VM_OSDISK=$(az vm show --ids "${VM_ID}" --query 'storageProfile.osDisk.name' --output tsv)
fi
VM_NIC_ON_PARENT_DELETE=$(az vm show --ids "${VM_ID}" --query 'networkProfile.networkInterfaces[].deleteOption' --output tsv)
VM_PIP_ON_PARENT_DELETE=${VM_NIC_ON_PARENT_DELETE}
if [ "${VM_NIC_ON_PARENT_DELETE}" != "Delete" ]; then
  VM_NIC=$(az vm show --ids "${VM_ID}" --query 'networkProfile.networkInterfaces[].id' --output tsv)
  VM_PIP_ON_PARENT_DELETE=$(az network nic show --ids "${VM_NIC}" --query 'ipConfigurations[].publicIPAddress.deleteOption' --output tsv)
  if [ "${VM_PIP_ON_PARENT_DELETE}" != "Delete" ]; then
    VM_PIP=$(az network nic show --ids "${VM_NIC}" --query 'ipConfigurations[].publicIPAddress.id' --output tsv)
  fi
fi

set | grep -E "^VM_*" | sort

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting VM: ${VM_ID}"
az vm delete --ids "${VM_ID}" --yes
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting VM done"

if [ -n "${VM_OSDISK}" ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting OSDisk: ${VM_OSDISK}"
  az disk delete --ids "${VM_OSDISK}" --yes
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting OSDisk done"
fi

if [ -n "${VM_NIC}" ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting NIC: ${VM_NIC}"
  az network nic delete --ids "${VM_NIC}" # --yes not available
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting done"
fi

if [ -n "${VM_PIP}" ]; then
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting PIP: ${VM_PIP}"
  az network public-ip delete --ids "${VM_PIP}" # --yes not available
  echo "[$(date '+%Y-%m-%d %H:%M:%S')] Deleting PIP done"
fi
