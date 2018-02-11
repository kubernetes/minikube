# multi-node prototype

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
* Theoretically, different CNIs can be used

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

* kube-dns Pods need to be deleted and re-created after CNI is installed

  This is because they will not have be in the right subnet when they are first started.

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

## Design concerns

### UX

* What should commands look like?

  `minikube node [subcommand]` ?

  Where `subcommand` is:

  * add
  * remove
  * start
  * stop
  * ssh
  * docker-env
  * status
  * list

### Drivers

* Can all drivers (except `none`) support multi-node?  Would it require a lot of custom code for each driver?

  From my experience with the Virtualbox setup, the main requirement is that each node have a unique IP that is routable among all nodes.  Most network plugins require layer 2 connectivity between the nodes.  This should not be a problem in a local VM network.

### Local docker images

* A useful pattern with a single node is to be able to run `eval "$(minikube docker-env)"` and push images directly to the local docker node's daemon.  With multiple nodes this becomes trickier.  

  There may be some clever solutions to this, but the simplest one is probably just some documentation or helper scrips that allow one to run `docker push` (or other commands) against all the nodes in a loop.

### Networking

Multi-node networking requires:

* subnet allocation for each node
* out-of-band network control/routing plane via CNI or out-of-band agents

For single-node setups, this is not required.

* What's a good way to accommodate multi-node without increasing complexity for single-node setups?

  Perhaps multi-node networking could be installed by minikube only after a worker node is added, so for single-node setups no extra complexity is introduced?

* Should minikube be responsible for configuring the network?

  In my opinion, yes, but it should be overrideable by the user.

* How should we set up networking?  CNI plugin, or out-of-band daemon?

  From what I understand, CNI allows the kubelet to set up networking on-demand, whereas an out-of-band daemon needs to hook into docker's configuration and set up `docker0`'s subnet at startup.

  It seems like CNI would be easier to implement, as it can be deployed as a daemonset via the Kube API.

* How do we ensure that networking is set up before any other pods are deployed?

  A taint might help with this.  It could be removed after CNI is installed

* How do we transition into multi-node networking from single-node if Pods are already running with node-local subnets?

  This is not an issue if CNI is always installed, and there is no distinction between single-node and multi-node networking.


### Bootstrappers

* Do we support both `localkube` and `kubeadm`?

kubeadm is designed for multi-node whereas localkube is not.  Is it reasonable to say that multi-node is only supported by `kubeadm`?
