# vm-driver=none

## Overview

This document is written for system integrators who are familiar with minikube, and wish to run it within a customized VM environment.

`--vm-driver=none` allows advanced minikube users to skip VM creation, allowing minikube to be run on a user-supplied VM.

## What operating systems are supported?

`--vm-driver=none` supports releases of Debian, Fedora, and buildroot that are less than 2 years old. 

While the standard minikube guest VM uses buildroot, minikube integration tests are also regularly run against Debian 9 for compatibility. In practice, any systemd-based modern distribution is likely to work, and we will happily accept pull requests which improve compatibility with other systems.

## Should vm-driver=none be used on a personal development machine? No.

No. Please do not do this, ever.

minikube was designed to run Kubernetes within a dedicated VM, and when used with `--vm-driver=none`, may overwrite system binaries, configuration files, and system logs. Executing `minikube --vm-driver=none` outside of a VM could result in data loss, system instability and decreased security.

Usage of `--vm-driver=none` outside of a VM could also result in services being exposed in a way that may make them accessible to the public internet. Even if your host is protected by a firewall, these services still be vulnerable to [CSRF](https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)) or [DNS rebinding](https://en.wikipedia.org/wiki/DNS_rebinding) attacks.

## Can vm-driver=none be used outside of a VM?

Yes, but only after appropriate security and reliability precautions have been made. `minikube --vm-driver=none` assumes  complete control over the environment is is executing within, and may overwrite system binaries, configuration files, and system logs. 

The host running `minikube --vm-driver=none` should be:

* Isolated from the rest of the network with a firewall
* Disposable and easily reprovisioned, as this mode may overwrite system binaries, configuration files, and system logs

If you find yourself running a web browser on the same host running `--vm-driver=none`, please see __Should vm-driver=none be used on a personal development machine? No.__

## Known Issues

* You cannot run more than one `--vm-driver=none` instance on a single host #2781
* `--vm-driver=none` deletes other local docker images #2705
* `--vm-driver=none` fails on distro's which do not use systemd #2704
* Many `minikube` commands are not supported, such as: `dashboard`, `mount`, `ssh`, `stop` #3127


## Example Installation for Linux Continuous Integration

```shell
curl -Lo minikube https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64 && chmod +x minikube && sudo cp minikube /usr/local/bin/ && rm minikube
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && sudo cp kubectl /usr/local/bin/ && rm kubectl

export MINIKUBE_WANTUPDATENOTIFICATION=false
export MINIKUBE_WANTREPORTERRORPROMPT=false
export MINIKUBE_HOME=$HOME
export CHANGE_MINIKUBE_NONE_USER=true
mkdir -p $HOME/.kube
mkdir -p $HOME/.minikube
touch $HOME/.kube/config

export KUBECONFIG=$HOME/.kube/config
sudo -E minikube start --vm-driver=none

# this for loop waits until kubectl can access the api server that Minikube has created
for i in {1..150}; do # timeout for 5 minutes
   kubectl get po &> /dev/null
   if [ $? -ne 1 ]; then
      break
  fi
  sleep 2
done
```

