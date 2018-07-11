#!/bin/bash -x

# ---- References: ----
# https://github.com/kubernetes/minikube

WITHOUT_VM=1

## -- With VM Driver
## VirtualBox or KVM or none
VM_DRIVER=VirtualBox

function replaceKeyValue() {
    if [ $# -lt 3 ]; then
        echo "ERROR: --- Usage: $0 <config_file> <key> <value> [<delimiter>] [<prefix-pattern>]"
        echo "e.g."
        echo './replaceKeyValue.sh \"elasticsearch.yml\" \"^network.host\" \"172.20.1.92\" \":\" \"# network\" '
        exit 1
    fi

    CONFIG_FILE=${1}
    TARGET_KEY=${2}
    REPLACEMENT_VALUE=${3}
    DELIMITER_TOKEN=${4:-:}
    PREFIX_PATTERN=${5:-}

    if [ ! -f "$CONFIG_FILE" ]; then
        echo "*** ERROR $CONFIG_FILE: Not found!"
        exit 1
    fi

    if grep -q "${TARGET_KEY} *${DELIMITER_TOKEN}" ${CONFIG_FILE}; then   
        #sudo sed -c -i "s/\($TARGET_KEY *= *\).*/\1$REPLACEMENT_VALUE/" $CONFIG_FILE
        sudo sed -i "s/\(${TARGET_KEY} *${DELIMITER_TOKEN} *\).*/\1${REPLACEMENT_VALUE}/" ${CONFIG_FILE}
    else
        if [ "$PREFIX_PATTERN" == "" ]; then
            #echo "$TARGET_KEY= $REPLACEMENT_VALUE" | sudo tee -a $CONFIG_FILE
            echo "${TARGET_KEY}${DELIMITER_TOKEN} ${REPLACEMENT_VALUE}" | sudo tee -a ${CONFIG_FILE}
        else
            sudo sed -i "/${PREFIX_PATTERN}/a \
    ${TARGET_KEY}${DELIMITER_TOKEN} ${REPLACEMENT_VALUE}" ${CONFIG_FILE}
        fi
    fi
}


# ------------- How to overcome issue with without VM Support -------------
#   https://github.com/kubernetes/minikube/issues/2575
#
# The workaround was not to specify a bridge IP for docker, as I had thought. Instead you need to start minikube like so:
# 
# minikube start--vm-driver=none --apiserver-ips 127.0.0.1 --apiserver-name localhost
#
# And then go and edit ~/.kube/config, replacing the server IP that was detected from the main network interface with 
# "localhost". For example, mine now looks like this:
# 
# - cluster:
#     certificate-authority: /home/jfeasel/.minikube/ca.crt
#     server: https://localhost:8443
#   name: minikube
# With this configuration, I can access my local cluster all of the time, even if the main network interface is disabled.
# 
# Also, we should note that it is required to have "socat" installed on the Linux environment. See this issue for details: 
# kubernetes/kubernetes#19765 I saw this when I tried to use helm to connect to my local cluster; I got errors with port-
# forwarding. Since I'm using Ubuntu all I had to do was sudo apt-get install socat and then everything worked as expected.

# -----------------------------------------------------------------
# Linux Continuous Integration without VM Support
# Example with kubectl installation:

MINIKUBE_URL=https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
INSTALL_DIR=/usr/local/bin
if [ ! -s minikube ]; then
    ## -- Local install --
    #curl -Lo minikube ${MINIKUBE_URL} && chmod +x minikube 
    ## -- System wide install --
    curl -Lo minikube ${MINIKUBE_URL} && chmod +x minikube && sudo sudo cp minikube ${INSTALL_DIR}/
fi
if [ ! -s ${INSTALL_DIR}/kubectl ]; then
    curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && sudo cp kubectl ${INSTALL_DIR}/

fi

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
export PATH="`pwd`/":$PATH

# ------ kube config file ------
mkdir -p $HOME/.kube
touch $HOME/.kube/config
export KUBECONFIG=$HOME/.kube/config

## ---- Wait for kube to come up ----
function waitForKubeUp() {
    # this for loop waits until kubectl can access the api server that Minikube has created
    for i in {1..150}; do # timeout for 5 minutes
       ./kubectl get po &> /dev/null
       if [ $? -ne 1 ]; then
          break
      fi
      sleep 2
    done
}

## ---- Start kube ----
if [ ${WITHOUT_VM} -eq 0 ]; then
    ## -- With VM Driver
    ## VirtualBox or KVM or none
    sudo -E minikube start --vm-driver=${VM_DRIVER}
    waitForKubeUp
else
    ## ****************************************************************************************
    ## **** WARNING: IT IS NOT RECOMMENDED TO RUN THE "none" DRIVER ON PERSONAL WORKSTATIONS!!
	## **** WARNING: The 'none' driver will run an insecure kubernetes apiserver as root that 
	## **** WARNING: may leave the host vulnerable to CSRF attacks
    ## ****************************************************************************************
    #VM_DRIVER=none
    #BRIDGE_SERVER_IP="`cat $HOME/.kube/config | grep server | grep 'localhost:8443' `"
    #if [ "${BRIDGE_SERVER_IP}" == "" ]; then
    #   sudo -E ./minikube start --vm-driver=none
    #   waitForKubeUp
    #   sudo -E ./minikube stop
    #   sleep 10
    #   ## -- Need to replace "Bridge IP=address to localhost 127.0.0.1"
    #   ## -- (see https://github.com/kubernetes/minikube/issues/2575)
    #   cp --backup=numbered $HOME/.kube/config $HOME/.kube/
    #   replaceKeyValue "$HOME/.kube/config" "server" "https\:\/\/localhost\:8443"
    #   sudo -E ./minikube start --vm-driver=none --apiserver-ips 127.0.0.1 --apiserver-name localhost
    #else
    #   sudo -E ./minikube start --vm-driver=none --apiserver-ips 127.0.0.1 --apiserver-name localhost
    #   waitForKubeUp
    #fi
    sudo -E minikube start --vm-driver=none
    waitForKubeUp
fi

# kubectl commands are now able to interact with Minikube cluster

#  Starting local Kubernetes v1.10.0 cluster...
#  Starting VM...
#  Getting VM IP address...
#  Moving files into cluster...
#  Setting up certs...
#  Connecting to cluster...
#  Setting up kubeconfig...
#  Starting cluster components...
#  Kubectl is now configured to use the cluster.
#  ===================
#  WARNING: IT IS RECOMMENDED NOT TO RUN THE NONE DRIVER ON PERSONAL WORKSTATIONS
#  	The 'none' driver will run an insecure kubernetes apiserver as root that may leave the host vulnerable to CSRF attacks


