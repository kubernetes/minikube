# Proposal: Support multi-node clusters


## Contents

- [Motivation](#motivation)
- [Goals](#goals)
- [Non-goals](#non-goals)
- [Proposed Design](#proposed-design)
	- [Experimental status](#experimental-status)
	- [UX](#ux)
		- [Node configuration](#node-configuration)
	- [Driver support](#driver-support)
	- [Bootstrapper support](#bootstrapper-support)
	- [Networking](#networking)
	- [Building local images](#building-local-images)
	- [Shared storage](#shared-storage)


## Motivation

Running a minikube cluster with multiple nodes allows one to test and play around with Kubernetes features that are not available in a single-node cluster.  This was inspired by [kubernetes/minikube#94](https://github.com/kubernetes/minikube/issues/94).

Some examples:

* Observe and test scheduling behavior based on
  * Resource allocation
  * Pod Affinity/Anti-affinity
  * Node selectors
  * Node Taints

* Networking layer
  * Experimenting with different CNI plugins and configurations

* Self-healing
  * Simulating node failure and migration of pods (and storage if supported)

* Development
  * Use multi-node cluster for local kubernetes development
    * This would require the ability to inject custom versions of kube components like the kubelet
  * Help simulate behavior that requires multiple nodes

## Goals

* Enable creation of arbitrary number of worker nodes (as many a host will allow)
* Continue to support single-node cluster by default without any significant change in usability
* Create nodes as VMs, not docker-in-docker or similar emulations, to better simulate production multi-node connectivity
* Provide default networking layer, replaceable with custom configuration
  * Support standard networking options, eg. using CNI

## Non-goals

* Supporting multi-master configuration (including multi-node etcd cluster)


## Proposed Design

### Experimental status

_Would we want to treat this feature as experimental at first?_

If this feature changes single-node behavior (eg. adding CNI from the beginning, see [networking](#networking)), then a config flag may be needed to enable this.  If it does not change single-node behavior, then it can remain neatly isolated behind the `minikube worker` command.

### UX

Manipulating node addition/removal/configuration would be done via the `worker` command:

```
minikube [--profile=minikube] worker [subcommand]
```

Subcommands would include:

* `list`

  Lists all nodes with metadata: status, IP

* `create [worker_name]`

  Creates a node with optional name, defaulting to a naming scheme like `node-N`.

* `start <worker_name OR --all>`

  Starts node with desired name, creating it if necessary.
  `--all` starts all nodes already created.

* `stop <worker_name>`

  Stops a node keeping its state, does not remove it from cluster's API

* `delete <worker_name>`

  Deletes a node, attempting to delete it from the API if the master is running.

* `ssh <worker_name> -- [command]`

  Opens a shell on the node, optionally runs a command.

* `docker-env <worker_name>`

  Prints the docker environment settings for connecting the docker client
  to this node.

* `status [worker_name]`

  Prints status of specified node, or all nodes if not specified.

* `logs [worker_name]`

  Prints or tails logs of the Kube components running on the node,
  similar to `minikube logs`.

* `ip <worker_name>`

  Print the IP of the specified node


#### Node configuration

Nodes would inherit the configuration of the master node from `minikube start`, but could be customized via flags provided to the start command, allowing one to have different resource capacity on different nodes as well as testing different configurations, including different versions of worker components, eg:

```
  minikube worker create node-1 \
    --cpus=1 \
    --memory=1024 \
    --disk-size=10g \
    --extra-opts=... \
    --container-runtime... \
    --kubernetes-version=... \
    --iso-url=... \
    --docker-env=... \
    --docker-opt=... \
    --feature-gates=... \
    ...
```

In short, most configuration options that can be passed to `minikube start` should be applicable to a worker node.

### Driver support

_What drivers can support multi-node clusters?_

The primary requirement is that each node must have a unique IP that is routable among all nodes.  Some network plugins require layer 2 connectivity between the nodes, while others encapsulate inter-pod traffic inside IP packets and can operate without layer-2 connectivity between all nodes.  Driver support may impact CNI support if layer 2 networking is not available on all drivers we want to support. We could say that layer-2 is a requirement, so that encapsulation of packets is never needed.

For example, if `flanneld`'s `host-gw` backend will only work if all nodes can see each other via layer-2.  If that is not available, the `vxlan` backend can be used, at a performance and complexity cost (probably not significant at minikube scale).

If a driver supports the above, theoretically it supports a multi-node setup.  An initial implementation could start with `virtualbox` support only, since it is the default driver.


### Bootstrapper support

Multi-node will only be supported by the `kubeadm` bootstrapper, as it is designed for multi-node, while `localkube` is not.


### Networking

A multi-node cluster relies on a CNI plugin deployed as a DaemonSet to handle routing between nodes.  Minikube installs a default CNI distribution out-of-the-box.  This can be disabled, allowing a user to install their own choice of CNI plugin, eg. `--cni-plugin=none`.

_How should we transition from single-node to multi-node networking?_

* _Do we always install a CNI plugin, even for single-node clusters?_
* _Do we install a CNI only when the first worker node is created?_

  If so, existing pods need to be recreated so they are assigned a new IP by the CNI.

  Using a CNI plugin from the beginning keeps the experience consistent in both scenarios, at the expense of running more pods all the time.

_What CNI plugin should we install by default?_

`flanneld` with the `host-gw` backend is one option that works well and makes minimal changes to the IP stack, only manipulating the routing table.


### Building local images

It is common to point a Docker client at the minikube VM, enabling one to build local images directly on the node's Docker instance.  With multiple independent Docker instances, it becomes necessary to keep them in sync if one is building local images.

For an initial implementation, we could leave it to the user to work around this.  Some helper bash scripts might be enough, for example:

```shell
#
# Runs a docker command against all nodes, including the master
#
docker_on_all_nodes() {
  cmd=$0
  shift

  # Master node
  echo "Running against master node ..."
  eval "$(minikube docker-env)" && docker $cmd "$@"
  status=$?
  if [ $status -ne 0 ]; then
    return $status
  fi

  nodes=$(minikube node list --format='{{ .Name }}')
  for node in ${nodes[@]}; do
    echo "Running against node $node ..."
    eval "$(minikube node docker-env "$node")" && docker $cmd "$@"
    status=$?
    if [ $status -ne 0 ]; then
      return $status
    fi
  done
}

# Build image on all nodes
docker_on_all_nodes build -t myimage .
```

### Shared storage

Shared storage remains out of scope of this proposal.  Initial users can mount a shared folder on every host and use `hostPath` to make it accessible to Pods.
