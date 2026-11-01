## Requirements

A Linux VM with the following:

* systemd or OpenRC
* a container runtime, such as containerd, CRI-O or Docker
* [cri-dockerd](https://github.com/Mirantis/cri-dockerd) (if using Kubernetes +v1.24 & `docker` container-runtime)
* a user with SSH key access and passwordless `sudo`

This VM must also meet the [kubeadm requirements](https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/install-kubeadm/), such as:

* 2 CPU's
* 2GB RAM
* iptables
* conntrack
* crictl
* cni-plugins
* SELinux permissive

## Usage

The ssh driver requires the IP address of the VM to use.

```shell
minikube start --driver=ssh --ssh-ip-address=vm.example.com
```

## Example: Ubuntu LTS with containerd

This example prepares a fresh Ubuntu Server LTS VM and starts minikube on it
using containerd. It was tested with Ubuntu 26.04 LTS (arm64) on a VM with
2 CPUs and 8 GB of RAM.

{{% alert title="Note" color="primary" %}}
Using `--container-runtime=containerd` with the ssh driver requires a minikube
version that includes [#23969](https://github.com/kubernetes/minikube/pull/23969).
{{% /alert %}}

### On the host: copy your SSH key to the VM

```shell
ssh-copy-id -i ~/.ssh/id_rsa.pub <user>@<vm-ip>
```

### On the VM: allow passwordless sudo

minikube runs `sudo` non-interactively over SSH.

```shell
echo "$USER ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/minikube
```

### On the VM: install containerd and dependencies

```shell
sudo apt-get update
sudo apt-get install -y containerd conntrack socat
```

Ubuntu's containerd package does not create `/etc/containerd/config.toml`,
which minikube needs. Create the default configuration:

```shell
sudo mkdir -p /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml > /dev/null
sudo systemctl restart containerd
```

### On the VM: install crictl and the CNI plugins

```shell
ARCH=$(dpkg --print-architecture)

CRICTL_VERSION=v1.35.0
curl -fsSL "https://github.com/kubernetes-sigs/cri-tools/releases/download/${CRICTL_VERSION}/crictl-${CRICTL_VERSION}-linux-${ARCH}.tar.gz" \
  | sudo tar -xz -C /usr/local/bin

CNI_PLUGIN_VERSION=v1.9.1
sudo mkdir -p /opt/cni/bin
curl -fsSL "https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGIN_VERSION}/cni-plugins-linux-${ARCH}-${CNI_PLUGIN_VERSION}.tgz" \
  | sudo tar -xz -C /opt/cni/bin
```

### On the host: start minikube

```shell
minikube start --driver=ssh \
  --ssh-ip-address=<vm-ip> --ssh-user=<user> --ssh-key=~/.ssh/id_rsa \
  --container-runtime=containerd
kubectl get nodes
```

The node should be `Ready` with `containerd` as its container runtime.
