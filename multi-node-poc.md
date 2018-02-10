# multi-node PoC

## How to use

* Check out this branch
* Build minikube using `make out/minikube` (see [Build Guide](docs/contributors/build_guide.md) for more info)
* Run custom binary for all operations, eg:

  ```shell
  out/minikube start
  out/minikube node add node-1
  out/minikube node start
  ```

## Current setup and limitations

### Features

* Pods can communicate across nodes
* Nodes get non-overlapping subnets for pod IPs
* Theoretically, various CNIs can be used

### Limitations

* Tested only with this combination:
  * `minikube-darwin-amd64` build, based off v0.24.1 revision
  * virtualbox driver
  * kubernetes v1.8.0
  * flanneld CNI plugin v0.9.1
* Only works with `kubeadm` bootstrapper (using `kubeadm join` command)
* Code is hacked-in, minimal refactoring done
  * A much cleaner refactoring would be needed to implement in a sustainable way
* CNI (in this case, flannel) is deployed manually, out-of-band (see [kube-flannel.yml](demo/kube-flannel.yml))
* Pods must be started only after CNI is up and running, otherwise they get a host-local IP instead of a cluster-routable one.
* Hard-coded pod subnets in CNI config (set up for `flannel`)
  * Same subnet must be used in `controller-manager` config _AND_ in `flannel` config
  * Currently hard-coded to `10.244.0.0/16` for flannel
* Some components are explicitly told to use `eth1` address, as `eth0` is not externally routable in the Virtualbox config provided by `libmachine`:
  * `kubelet`
  * `flannel`
* All nodes share the same configuration, including CPU, disk, and memory capacity

### Bugs

* Deleting the cluster with `minikube delete` will not delete node data and will cause subsequent node starts to crash if the nodes have the same name.  

To fix this you must:

* Remove the nodes one by one with `minikube node remove <node_name>` _before_ deleting the minikube cluster

_or_

* Delete Virtualbox VMs manually
* Delete the machine data folders manually with:

  ```shell
  rm -r ~/.minikube/machines/minikube-*/
  ```

  where your minikube config path is `~/.minikube` and the minikube profile name is `minikube` (the default)
