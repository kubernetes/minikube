set -e
MINIKUBE_INSTALL_HOME="$(./bin/minikube-linux-amd64 home)"

mkdir -p ${MINIKUBE_INSTALL_HOME}/bin
cp ./bin/docker-machine-driver-kvm-* ${MINIKUBE_INSTALL_HOME}/bin
cp ./bin/minikube-linux-amd64 ${MINIKUBE_INSTALL_HOME}/bin/minikube

cp -r ./addons ${MINIKUBE_INSTALL_HOME}
cp -r ./iso ${MINIKUBE_INSTALL_HOME}/cache
cp -r ./localkube ${MINIKUBE_INSTALL_HOME}/cache

PATH_CONFIG=${HOME}/.bashrc
touch ${PATH_CONFIG}
MINIKUBE_PATH_COMMENT='# The next line updates PATH for minikube.'
if ! grep -q "${MINIKUBE_PATH_COMMENT}" ${PATH_CONFIG}
  then
    echo '' >>${PATH_CONFIG}
    echo ${MINIKUBE_PATH_COMMENT} >>${PATH_CONFIG}
    echo 'if [ -f "'${MINIKUBE_INSTALL_HOME}'/path.bash.inc" ]; then source "'${MINIKUBE_INSTALL_HOME}'/path.bash.inc"; fi' >>${PATH_CONFIG}
fi
# TODO(aaron-prindle) install bash completion on the system
# TODO(aaron-prindle) support zsh and fish shells

# TODO(aaron-prindle) make sure that python installation is acceptable requirement on linux distros
# TODO(aaron-prindle) prompt for kvm/libvirt installation
if python -mplatform | grep -qi Ubuntu
  then
    if ! dpkg -s libvirt-bin > /dev/null
    then
      sudo apt -y install libvirt-bin qemu-kvm > /dev/null
    fi
    # For Ubuntu 17.04 change the group to `libvirt`
    if python -mplatform | grep -qi Ubuntu-17.04
    then
      sudo usermod -a -G libvirt $(whoami)
      # newgrp libvirt -
    else
      sudo usermod -a -G libvirtd $(whoami)    
      # newgrp libvirtd -
    fi
  else
    if ! yum list installed libvirt-daemon-kvm > /dev/null
    then
      sudo yum -y install libvirt-daemon-kvm kvm > /dev/null      
    fi
    sudo usermod -a -G libvirt $(whoami) 
    # newgrp libvirt -
fi

echo 'Installed minikube to '${MINIKUBE_INSTALL_HOME}'

Reload your current session (login/logout) for the changes made to the libvirtd group to take effect.
This is required to use minikube with the kvm driver.'
