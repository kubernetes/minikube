#!/bin/bash

# Copyright 2020 The Kubernetes Authors All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# cleanup shared between Linux and macOS
function check_jenkins() {
  jenkins_pid="$(pidof java)"
  if [[ "${jenkins_pid}" = "" ]]; then
          return
  fi
  pstree "${jenkins_pid}" \
        | egrep -i 'bash|integration|e2e|minikube' \
        && echo "tests are is running on pid ${jenkins_pid} ..." \
        && exit 1
}

check_jenkins
logger "cleanup_and_reboot running - may shutdown in 60 seconds"
echo "cleanup_and_reboot running - may shutdown in 60 seconds" | wall
sleep 10
check_jenkins
logger "cleanup_and_reboot is happening!"

# kill jenkins to avoid an incoming request
killall java

# clean minikube left overs
echo -e "\ncleanup minikube..."
killall minikube >/dev/null 2>&1 || true
USERS="$(lslogins --user-accs --noheadings --output=USER)"
for user in $USERS; do
    if sudo su - $user -c "minikube delete --all --purge" >/dev/null 2>&1; then
	echo "successfully cleaned up minikube for $user user"
    fi
done

# clean docker left overs
echo -e "\ncleanup docker..."
docker kill $(docker ps -aq) >/dev/null 2>&1 || true
docker system prune --volumes --force || true

# clean KVM left overs
echo -e "\ncleanup kvm..."
overview() {
	echo -e "\n - KVM domains:"
	sudo virsh list --all || true
	echo " - KVM pools:"
	sudo virsh pool-list --all || true
	echo " - KVM networks:"
	sudo virsh net-list --all || true
	echo " - host networks:"
	sudo ip link show || true
}
echo -e "\nbefore the cleanup:"
overview
for DOM in $( sudo virsh list --all --name ); do
	if sudo virsh destroy "${DOM}"; then
		if sudo virsh undefine "${DOM}"; then
			echo "successfully deleted KVM domain:" "${DOM}"
			continue
		fi
		echo "unable to delete KVM domain:" "${DOM}"
	fi
done
#for POOL in $( sudo virsh pool-list --all --name ); do  # better, but flag '--name' is not supported for 'virsh pool-list' command on older libvirt versions
for POOL in $( sudo virsh pool-list --all | awk 'NR>2 {print $1}' ); do
	for VOL in $( sudo virsh vol-list "${POOL}" ); do
		if sudo virsh vol-delete --pool "${POOL}" "${VOLUME}"; then  # flag '--delete-snapshots': "delete snapshots associated with volume (must be supported by storage driver)"
			echo "successfully deleted KVM pool/volume:" "${POOL}"/"${VOL}"
			continue
		fi
		echo "unable to delete KVM pool/volume:" "${POOL}"/"${VOL}"
	done
done
for NET in $( sudo virsh net-list --all --name ); do
	if [ "${NET}" != "default" ]; then
		if sudo virsh net-destroy "${NET}"; then
			if sudo virsh net-undefine "${NET}"; then
				echo "successfully deleted KVM network" "${NET}"
				continue
			fi
		fi
		echo "unable to delete KVM network" "${NET}"
	fi
done
# DEFAULT_BRIDGE is a bridge connected to the 'default' KVM network
DEFAULT_BRIDGE=$( sudo virsh net-info default | awk '{ if ($1 == "Bridge:") print $2 }' )
echo "bridge connected to the 'default' KVM network to leave alone:" "${DEFAULT_BRIDGE}"
for VIF in $( sudo ip link show | awk -v defvbr="${DEFAULT_BRIDGE}.*" -F': ' '$2 !~ defvbr { if ($2 ~ /virbr.*/ || $2 ~ /vnet.*/) print $2 }' ); do
	if sudo ip link delete "${VIF}"; then
		echo "successfully deleted KVM interface" "${VIF}"
		continue
	fi
	echo "unable to delete KVM interface" "${VIF}"
done
echo -e "\nafter the cleanup:"
overview

# Linux-specific cleanup

# disable localkube, kubelet
systemctl list-unit-files --state=enabled \
        | grep kube \
        | awk '{ print $1 }' \
        | xargs systemctl disable

# update and reboot
apt update -y && apt upgrade -y && apt-get autoclean && reboot
