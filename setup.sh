#!/bin/bash
set -e

if [ "$EUID" -eq 0 ]
  then echo "Run without sudo/root"
  exit
fi

# Install Docker dependencies
sudo apt-get -y install apt-transport-https ca-certificates curl gnupg-agent software-properties-common
sleep 1

# Add Docker key (Ubuntu Bionic/Linux Mint) - TODO: Pass parameter to add key based on distro type
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(. /etc/os-release; echo "$UBUNTU_CODENAME") stable"
sudo apt-get update

# Install Docker
sudo apt-get -y install docker-ce docker-ce-cli containerd.io
sleep 1

# Add user to docker group
sudo usermod -aG docker $USER
sleep 1

# Install kubectl
sudo curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
sudo chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
sudo rm -rf kubectl
kubectl version
sleep 1

# Download latest version of minikube and install - TODO: Get the latest version of minikube instead of hard coding the url
sudo curl -Lo minikube https://storage.googleapis.com/minikube/releases/v1.2.0/minikube-linux-amd64
sudo chmod +x minikube 
sudo cp minikube /usr/local/bin/
sudo rm -rf minikube
sleep 1

# Set env variable for minikube
export CHANGE_MINIKUBE_NONE_USER=true

# Start minikube using none driver (Linux) - TODO: Receive driver type as script parameter and other parameters as well
sudo minikube start --vm-driver=none
sleep 1

# Set minikube user chown permissions
sudo chown -R $USER $HOME/.kube $HOME/.minikube
sudo chown -R $USER $HOME/.minikube
sleep 1

# Enable minikube dashboard addon
sudo minikube addons enable dashboard
sleep 1

# Switch to use local docker daemon (needed if not using none driver)
eval $(minikube docker-env)
sleep 1

# Get minikube status
minikube status

# Start minikube dashboard
minikube dashboard &


