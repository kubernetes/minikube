#!/bin/bash
#
# Script for minikube full installation on Debian/Ubuntu/Mint
# Install minikube dependencies (Docker and kubectl)
# Start minikube and minikube dashboard
# Options for minikube start (--vm-driver): virtualbox, kvm, none
# Usage: 
#   ./setup.sh
#   ./setup.sh none
#   ./setup.sh virtualbox
#   ./setup.sh kvm
#

set -e

if [ "$EUID" -eq 0 ]
  then echo "Run without sudo/root"
  exit
fi

# Check minikube driver (virtualbox, kvm, none)
if [ $1 = "none" ]; then
  DRIVER="--vm-driver=none"
elif [ $1 = "virtualbox" ]; then
  DRIVER="--vm-driver=virtualbox"
elif [ $1 = "kvm" ]; then
  DRIVER="--vm-driver=kvm"
fi

# Install Docker dependencies
echo "Installing Docker dependencies.."
sudo apt-get -y install apt-transport-https ca-certificates curl gnupg-agent software-properties-common

# Add Docker key (Ubuntu Bionic/Linux Mint) - TODO: Pass parameter to add key based on distro type
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(. /etc/os-release; echo "$UBUNTU_CODENAME") stable"
sudo apt-get update

# Install Docker
echo "Installing Docker.."
sudo apt-get -y install docker-ce docker-ce-cli containerd.io

# Add user to docker group
sudo usermod -aG docker $USER

# Install kubectl
echo "Installing kubectl.."
sudo curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
sudo chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
sudo rm -rf kubectl

# Download latest version of minikube
echo "Installing minikube.."
sudo curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo chmod +x minikube 
sudo cp minikube /usr/local/bin/
sudo rm -rf minikube
minikube version
kubectl version

# Set env variable for minikube
export CHANGE_MINIKUBE_NONE_USER=true

# Start minikube
echo "Starting minikube using driver ${DRIVER}.."
sudo minikube start ${DRIVER}

# Set minikube user chown permissions
sudo chown -R $USER $HOME/.kube $HOME/.minikube
sudo chown -R $USER $HOME/.minikube

# Enable minikube dashboard addon
sudo minikube addons enable dashboard

# Switch to use local docker daemon (needed if not using none driver)
eval $(minikube docker-env)

# Get minikube status
minikube status

# Set kubectl minikube context
kubectl config use-context minikube

# Start minikube dashboard
minikube dashboard &