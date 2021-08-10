# Vagrant

_A virtual machine version of KIC._

See <https://www.vagrantup.com/>

Currently only tested with the VirtualBox hypervisor and the Docker container runtime...

Only missing configuration and packages, for other virtualizations systems and runtimes.

## Start the virtual machine

```shell
vagrant up
```

uses the [Vagrantfile](Vagrantfile)

```text
Bringing machine 'default' up with 'virtualbox' provider...
==> default: Importing base box 'ubuntu/focal64'...
==> default: Matching MAC address for NAT networking...
==> default: Checking if box 'ubuntu/focal64' version '20210803.0.0' is up to date...
==> default: Setting the name of the VM: kicbase_default_1628606177659_66826
==> default: Clearing any previously set network interfaces...
==> default: Preparing network interfaces based on configuration...
    default: Adapter 1: nat
    default: Adapter 2: hostonly
==> default: Forwarding ports...
    default: 22 (guest) => 2222 (host) (adapter 1)
==> default: Running 'pre-boot' VM customizations...
==> default: Booting VM...
==> default: Waiting for machine to boot. This may take a few minutes...
    default: SSH address: 127.0.0.1:2222
    default: SSH username: vagrant
    default: SSH auth method: private key
    default: 
    default: Vagrant insecure key detected. Vagrant will automatically replace
    default: this with a newly generated keypair for better security.
    default: 
    default: Inserting generated public key within guest...
    default: Removing insecure key from the guest if it's present...
    default: Key inserted! Disconnecting and reconnecting using new SSH key...
==> default: Machine booted and ready!
==> default: Checking for guest additions in VM...
==> default: Setting hostname...
==> default: Configuring and enabling network interfaces...
==> default: Running provisioner: shell...
    default: Running: inline script
    ...
    default: Done.
```

## Access the virtual machine

```shell
vagrant ssh
```

Uses ssh tunnel port.

## Start minikube using it

```shell
minikube start --driver=ssh --native-ssh=false \
               --ssh-ip-address=192.168.50.4 --ssh-user=vagrant --ssh-port=22 \
               --ssh-key=$PWD/.vagrant/machines/default/virtualbox/private_key
```

Use the OpenSSH `ssh` binary, rather than the native Go SSH implementation.

Note: needs the private network (192.168.50.4), not the NAT network (10.0.2.15).

```text
😄  [vagrant] minikube v1.22.0 on Ubuntu 20.04
✨  Using the ssh driver based on user configuration
👍  Starting control plane node vagrant in cluster vagrant
🔗  Running remotely (CPUs=2, Memory=1987MB, Disk=39643MB) ...
🐳  Preparing Kubernetes v1.21.2 on Docker 20.10.8 ...
    ▪ Generating certificates and keys ...
    ▪ Booting up control plane ...
    ▪ Configuring RBAC rules ...
🔎  Verifying Kubernetes components...
    ▪ Using image gcr.io/k8s-minikube/storage-provisioner:v5
🌟  Enabled addons: storage-provisioner, default-storageclass
🏄  Done! kubectl is now configured to use "vagrant" cluster and "default" namespace by default
```

## System information

* OS: Ubuntu 20.04 LTS (`focal`)
* Box: <https://app.vagrantup.com/ubuntu/boxes/focal64>
