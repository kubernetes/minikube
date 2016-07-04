# Driver plugin installation

Minikube uses Docker Machine to manage the Kubernetes VM so it benefits from the
driver plugin architecture that Docker Machine uses to provide a consistent way to
manage various VM providers. Minikube embeds VirtualBox and VMware Fusion drivers
so there are no additional steps to use them. However, other drivers require an
extra binary to be present in the host PATH.

The following drivers currently require driver plugin binaries to be present in
the host PATH:

* [KVM](#kvm-driver)
* [xhyve](#xhyve-driver)

#### KVM driver

Download the `docker-machine-driver-kvm` binary from
https://github.com/dhiltgen/docker-machine-kvm/releases and put it somewhere in
your PATH. Minikube is currently tested against `docker-machine-driver-kvm` 0.7.0.

#### xhyve driver

From https://github.com/zchee/docker-machine-driver-xhyve#install:

```
$ brew install docker-machine-driver-xhyve

# docker-machine-driver-xhyve need root owner and uid
$ sudo chown root:wheel $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
$ sudo chmod u+s $(brew --prefix)/opt/docker-machine-driver-xhyve/bin/docker-machine-driver-xhyve
```
